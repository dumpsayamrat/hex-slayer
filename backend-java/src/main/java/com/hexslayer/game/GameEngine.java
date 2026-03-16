package com.hexslayer.game;

import com.hexslayer.model.Character;
import com.hexslayer.model.CharacterEngagement;
import com.hexslayer.model.MapMonster;
import com.hexslayer.repository.CharacterEngagementRepository;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.repository.MapMonsterRepository;
import com.hexslayer.ws.WebSocketHub;
import com.uber.h3core.H3Core;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.*;
import java.util.concurrent.*;

import static com.hexslayer.game.GameConfig.*;

@Component
public class GameEngine {

    private static final Logger log = LoggerFactory.getLogger(GameEngine.class);

    private final CharacterRepository charRepo;
    private final MapMonsterRepository monsterRepo;
    private final CharacterEngagementRepository engagementRepo;
    private final WebSocketHub hub;
    private final Movement movement;

    private final Map<String, ScheduledFuture<?>> activeZones = new ConcurrentHashMap<>();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(4);

    public GameEngine(CharacterRepository charRepo, MapMonsterRepository monsterRepo,
                      CharacterEngagementRepository engagementRepo, WebSocketHub hub) {
        this.charRepo = charRepo;
        this.monsterRepo = monsterRepo;
        this.engagementRepo = engagementRepo;
        this.hub = hub;

        H3Core h3;
        try {
            h3 = H3Core.newInstance();
        } catch (IOException e) {
            throw new RuntimeException("Failed to initialize H3", e);
        }
        this.movement = new Movement(h3);
    }

    public void start() {
        hub.setOnSubscribe(topic -> {
            String zone = topicToZone(topic);
            if (zone != null) ensureZoneLoop(zone);
        });
        log.info("game engine: ready (zone loops start on subscriber)");
    }

    public void ensureZoneLoop(String zone) {
        if (activeZones.containsKey(zone)) return;

        int[] tickCount = {0};
        int maxTicks = (ZONE_MAX_DURATION_MINS * 60) / TICK_INTERVAL_SECONDS;

        ScheduledFuture<?> future = scheduler.scheduleAtFixedRate(() -> {
            try {
                tickCount[0]++;
                if (tickCount[0] > maxTicks) {
                    stopZone(zone, "hit " + ZONE_MAX_DURATION_MINS + " min limit");
                    return;
                }
                if (!tickZone(zone)) {
                    stopZone(zone, "no alive characters");
                }
            } catch (Exception e) {
                log.error("game engine: error in zone loop {}", zone, e);
            }
        }, TICK_INTERVAL_SECONDS, TICK_INTERVAL_SECONDS, TimeUnit.SECONDS);

        activeZones.put(zone, future);
        log.info("game engine: started zone loop {}", zone);
    }

    private void stopZone(String zone, String reason) {
        ScheduledFuture<?> future = activeZones.remove(zone);
        if (future != null) {
            future.cancel(false);
        }
        log.info("game engine: stopped zone loop {} ({})", zone, reason);
    }

    private boolean tickZone(String zone) {
        String topic = "zone:" + zone;

        List<Character> characters = charRepo.findByH3ZoneAndIsAliveTrue(zone);
        if (characters.isEmpty()) return false;

        List<MapMonster> monsters = monsterRepo.findByH3ZoneAndIsAliveTrue(zone);

        Map<String, MapMonster> monsterByID = new HashMap<>();
        for (MapMonster m : monsters) {
            monsterByID.put(m.getId(), m);
        }

        List<String> charIds = characters.stream().map(Character::getId).toList();
        List<CharacterEngagement> engagements = charIds.isEmpty() ?
                List.of() : engagementRepo.findByCharacterIdIn(charIds);

        Map<String, CharacterEngagement> engagementByChar = new HashMap<>();
        for (CharacterEngagement e : engagements) {
            engagementByChar.put(e.getCharacter().getId(), e);
        }

        Map<String, Boolean> engaged = new HashMap<>();

        List<Map<String, Object>> allEvents = new ArrayList<>();
        for (Character ch : characters) {
            CharTick ct = new CharTick(charRepo, monsterRepo, engagementRepo, movement,
                    ch, monsterByID, monsters, engaged, engagementByChar);
            allEvents.addAll(ct.process());
        }

        for (Map<String, Object> evt : allEvents) {
            hub.broadcast(topic, evt);
        }

        return true;
    }

    private String topicToZone(String topic) {
        if (topic.startsWith("zone:")) {
            return topic.substring(5);
        }
        return null;
    }
}
