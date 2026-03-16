package com.hexslayer.ws;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;
import org.springframework.web.util.UriComponentsBuilder;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.hexslayer.model.Character;
import com.hexslayer.model.CharacterEngagement;
import com.hexslayer.model.MapMonster;
import com.hexslayer.model.Player;
import com.hexslayer.repository.CharacterEngagementRepository;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.repository.MapMonsterRepository;
import com.hexslayer.repository.PlayerRepository;

@Component
public class GameWebSocketHandler extends TextWebSocketHandler {

    private static final Logger log = LoggerFactory.getLogger(GameWebSocketHandler.class);
    private static final ObjectMapper mapper = new ObjectMapper();

    private final WebSocketHub hub;
    private final PlayerRepository playerRepo;
    private final CharacterRepository characterRepo;
    private final CharacterEngagementRepository engagementRepo;
    private final MapMonsterRepository mapMonsterRepo;

    public GameWebSocketHandler(WebSocketHub hub, PlayerRepository playerRepo,
                                 CharacterRepository characterRepo,
                                 CharacterEngagementRepository engagementRepo,
                                 MapMonsterRepository mapMonsterRepo) {
        this.hub = hub;
        this.playerRepo = playerRepo;
        this.characterRepo = characterRepo;
        this.engagementRepo = engagementRepo;
        this.mapMonsterRepo = mapMonsterRepo;
    }

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        String token = extractToken(session);
        if (token == null) {
            session.close(CloseStatus.POLICY_VIOLATION);
            return;
        }

        Optional<Player> player = playerRepo.findBySessionToken(token);
        if (player.isEmpty()) {
            session.close(CloseStatus.POLICY_VIOLATION);
            return;
        }

        session.getAttributes().put("player", player.get());
        log.info("websocket connected: player={}", player.get().getId());

        sendJson(session, Map.of("type", "connected", "message", "welcome to hexslayer"));
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        Player player = (Player) session.getAttributes().get("player");
        if (player == null) return;

        Map<String, Object> msg = mapper.readValue(message.getPayload(), Map.class);
        String msgType = (String) msg.get("type");
        log.info("ws message: player={} type={}", player.getId(), msgType);

        switch (msgType != null ? msgType : "") {
            case "ping" -> sendJson(session, Map.of("type", "pong"));

            case "subscribe_zone" -> {
                String zone = (String) msg.get("h3_zone");
                if (zone == null || zone.isEmpty()) {
                    sendJson(session, Map.of("type", "error", "message", "h3_zone required"));
                    return;
                }
                hub.subscribe("zone:" + zone, session);
                sendJson(session, Map.of("type", "subscribed", "h3_zone", zone));
                sendZoneSnapshot(session, zone);
            }

            case "unsubscribe_zone" -> {
                String zone = (String) msg.get("h3_zone");
                if (zone == null || zone.isEmpty()) {
                    sendJson(session, Map.of("type", "error", "message", "h3_zone required"));
                    return;
                }
                hub.unsubscribe("zone:" + zone, session);
                sendJson(session, Map.of("type", "unsubscribed", "h3_zone", zone));
            }

            default -> sendJson(session, Map.of("type", "error", "message", "unknown message type"));
        }
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        hub.unsubscribeAll(session);
    }

    private void sendZoneSnapshot(WebSocketSession session, String zone) throws IOException {
        List<Character> characters = characterRepo.findByH3ZoneAndIsAliveTrue(zone);

        List<String> charIds = characters.stream().map(Character::getId).toList();
        Map<String, String> engagementByChar = new HashMap<>();
        List<String> engagedMonsterIds = new ArrayList<>();

        if (!charIds.isEmpty()) {
            List<CharacterEngagement> engagements = engagementRepo.findByCharacterIdIn(charIds);
            for (CharacterEngagement e : engagements) {
                engagementByChar.put(e.getCharacter().getId(), e.getMonster().getId());
                engagedMonsterIds.add(e.getMonster().getId());
            }
        }

        List<Map<String, Object>> charData = characters.stream().map(c -> {
            Map<String, Object> entry = new HashMap<>(Map.of(
                    "id", c.getId(),
                    "name", c.getName(),
                    "hp", c.getHp(),
                    "max_hp", c.getMaxHP(),
                    "player_id", c.getPlayerId(),
                    "h3_index", c.getH3Index()
            ));
            String monsterId = engagementByChar.get(c.getId());
            if (monsterId != null) {
                entry.put("fighting_monster_id", monsterId);
            }
            return entry;
        }).toList();

        List<Map<String, Object>> monsterData = new ArrayList<>();
        if (!engagedMonsterIds.isEmpty()) {
            List<MapMonster> monsters = mapMonsterRepo.findAllById(engagedMonsterIds);
            monsterData = monsters.stream().map(m -> Map.<String, Object>of(
                    "id", m.getId(),
                    "current_hp", m.getCurrentHP()
            )).toList();
        }

        sendJson(session, Map.of(
                "type", "zone_snapshot",
                "zone", zone,
                "characters", charData,
                "monsters", monsterData
        ));
    }

    private String extractToken(WebSocketSession session) {
        String query = session.getUri().getQuery();
        if (query == null) return null;
        return UriComponentsBuilder.newInstance().query(query).build()
                .getQueryParams().getFirst("token");
    }

    private void sendJson(WebSocketSession session, Object payload) throws IOException {
        synchronized (session) {
            session.sendMessage(new TextMessage(mapper.writeValueAsString(payload)));
        }
    }
}
