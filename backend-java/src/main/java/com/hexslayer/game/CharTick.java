package com.hexslayer.game;

import com.hexslayer.model.Character;
import com.hexslayer.model.CharacterEngagement;
import com.hexslayer.model.MapMonster;
import com.hexslayer.repository.CharacterEngagementRepository;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.repository.MapMonsterRepository;

import java.time.Instant;
import java.util.*;

import static com.hexslayer.game.GameConfig.*;

public class CharTick {

    private final CharacterRepository charRepo;
    private final MapMonsterRepository monsterRepo;
    private final CharacterEngagementRepository engagementRepo;
    private final Movement movement;

    private final Character ch;
    private final Map<String, MapMonster> monsterByID;
    private final List<MapMonster> monsterList;
    private final Map<String, Boolean> engaged;
    private final Map<String, CharacterEngagement> engagementByChar;

    private final List<Map<String, Object>> events = new ArrayList<>();

    public CharTick(CharacterRepository charRepo, MapMonsterRepository monsterRepo,
                    CharacterEngagementRepository engagementRepo, Movement movement,
                    Character ch, Map<String, MapMonster> monsterByID, List<MapMonster> monsterList,
                    Map<String, Boolean> engaged, Map<String, CharacterEngagement> engagementByChar) {
        this.charRepo = charRepo;
        this.monsterRepo = monsterRepo;
        this.engagementRepo = engagementRepo;
        this.movement = movement;
        this.ch = ch;
        this.monsterByID = monsterByID;
        this.monsterList = monsterList;
        this.engaged = engaged;
        this.engagementByChar = engagementByChar;
    }

    public List<Map<String, Object>> process() {
        if (!ch.isAlive()) return events;

        CharacterEngagement eng = engagementByChar.get(ch.getId());

        // STATE: COMBAT
        if (eng != null) {
            processCombat(eng);
            return events;
        }

        // STATE: HUNTING
        if (ch.getTargetMonsterId() != null) {
            if (processHunting()) return events;
        }

        // STATE: SCANNING
        if (processScanning()) return events;

        // STATE: WANDERING
        processWandering();
        return events;
    }

    private void processCombat(CharacterEngagement eng) {
        MapMonster monster = monsterByID.get(eng.getMonster().getId());

        if (monster == null || !monster.isAlive()) {
            disengage();
            ch.setTargetMonsterId(null);
            events.addAll(wanderStep());
            return;
        }

        engaged.put(monster.getId(), true);
        List<Map<String, Object>> logs = Combat.doCombat(charRepo, monsterRepo, ch, monster);
        events.addAll(logs);

        handleCombatOutcome(monster);
    }

    private boolean processHunting() {
        MapMonster target = monsterByID.get(ch.getTargetMonsterId());

        if (target == null || !target.isAlive() || engaged.containsKey(target.getId())) {
            ch.setTargetMonsterId(null);
            charRepo.save(ch);
            return false;
        }

        Movement.MoveResult result = movement.moveToward(ch, target);
        if (!result.h3Index().equals(ch.getH3Index())) {
            ch.setH3Index(result.h3Index());
            charRepo.save(ch);
            events.add(Map.of(
                    "type", "char_move",
                    "character_id", ch.getId(),
                    "h3_index", result.h3Index()
            ));
        }

        if (result.distance() > GRID_DISK_RADIUS) {
            return true; // Keep walking
        }

        // Close enough — engage!
        engaged.put(target.getId(), true);
        CharacterEngagement newEng = new CharacterEngagement();
        newEng.setId(UUID.randomUUID().toString());
        newEng.setCharacter(ch);
        newEng.setMonster(target);
        newEng.setEngagedAt(Instant.now());
        engagementRepo.save(newEng);
        engagementByChar.put(ch.getId(), newEng);

        events.add(Map.of(
                "type", "combat_engage",
                "character_id", ch.getId(),
                "monster_id", target.getId()
        ));

        // First strike
        List<Map<String, Object>> logs = Combat.doCombat(charRepo, monsterRepo, ch, target);
        events.addAll(logs);

        handleCombatOutcome(target);
        return true;
    }

    private boolean processScanning() {
        MapMonster target = movement.findNearestFreeMonster(ch, monsterList, engaged);
        if (target == null) return false;

        ch.setTargetMonsterId(target.getId());
        charRepo.save(ch);
        return true;
    }

    private void processWandering() {
        String newIndex = movement.wander(ch);
        if (newIndex.equals(ch.getH3Index())) return;

        ch.setH3Index(newIndex);
        charRepo.save(ch);
        events.add(Map.of(
                "type", "char_move",
                "character_id", ch.getId(),
                "h3_index", newIndex
        ));
    }

    private void handleCombatOutcome(MapMonster monster) {
        if (!monster.isAlive()) {
            ch.setKills(ch.getKills() + 1);
            charRepo.save(ch);

            Map<String, Object> evt = new HashMap<>();
            evt.put("type", "monster_died");
            evt.put("monster_id", monster.getId());
            evt.put("killed_by", ch.getName());
            events.add(evt);

            disengage();
            ch.setTargetMonsterId(null);
            events.addAll(wanderStep());
        }

        if (!ch.isAlive()) {
            Map<String, Object> evt = new HashMap<>();
            evt.put("type", "character_died");
            evt.put("character_id", ch.getId());
            evt.put("killed_by", monster.getMonsterType().getName());
            events.add(evt);

            disengage();
        }
    }

    private void disengage() {
        CharacterEngagement eng = engagementByChar.remove(ch.getId());
        if (eng != null) {
            engagementRepo.delete(eng);
        }
    }

    private List<Map<String, Object>> wanderStep() {
        String newIndex = movement.wander(ch);
        if (newIndex.equals(ch.getH3Index())) return List.of();

        ch.setH3Index(newIndex);
        ch.setTargetMonsterId(null);
        charRepo.save(ch);

        return List.of(Map.of(
                "type", "char_move",
                "character_id", ch.getId(),
                "h3_index", newIndex
        ));
    }
}
