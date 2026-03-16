package com.hexslayer.service;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Random;
import java.util.Set;
import java.util.UUID;
import java.util.stream.Collectors;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import com.hexslayer.dto.ZoneMonsterResponse;
import com.hexslayer.model.MapMonster;
import com.hexslayer.model.MonsterType;
import com.hexslayer.repository.MapMonsterRepository;
import com.hexslayer.repository.MonsterTypeRepository;
import com.uber.h3core.H3Core;

@Service
public class ZoneService {

    private static final Logger log = LoggerFactory.getLogger(ZoneService.class);
    private static final int ZONE_RESOLUTION = 6;
    private static final int ENTITY_RESOLUTION = 12;
    private static final int ZONE_MONSTER_CAP = 300;

    private final MapMonsterRepository mapMonsterRepo;
    private final MonsterTypeRepository monsterTypeRepo;
    private final H3Core h3;

    public ZoneService(MapMonsterRepository mapMonsterRepo, MonsterTypeRepository monsterTypeRepo) {
        this.mapMonsterRepo = mapMonsterRepo;
        this.monsterTypeRepo = monsterTypeRepo;
        try {
            this.h3 = H3Core.newInstance();
        } catch (IOException e) {
            throw new RuntimeException("Failed to initialize H3", e);
        }
    }

    public record ZoneResult(String h3Zone, List<ZoneMonsterResponse> monsters) {}

    public ZoneResult getOrCreateMonsters(double lat, double lng) {
        long zoneAddr = h3.latLngToCell(lat, lng, ZONE_RESOLUTION);
        String zoneStr = h3.h3ToString(zoneAddr);

        long aliveCount = mapMonsterRepo.countByH3ZoneAndIsAliveTrue(zoneStr);
        int threshold = ZONE_MONSTER_CAP / 5;
        int toSpawn = ZONE_MONSTER_CAP - (int) aliveCount;

        log.info("zone {}: aliveCount={} threshold={} toSpawn={}", zoneStr, aliveCount, threshold, toSpawn);

        if (aliveCount < threshold) {
            spawnMonsters(zoneStr, zoneAddr, toSpawn);
        }

        List<MapMonster> monsters = mapMonsterRepo.findByH3ZoneAndIsAliveTrue(zoneStr);
        log.info("zone {}: returning {} monsters", zoneStr, monsters.size());

        List<ZoneMonsterResponse> result = monsters.stream().map(m -> new ZoneMonsterResponse(
                m.getId(),
                m.getH3Index(),
                m.getMonsterType().getName(),
                m.getMonsterType().getIcon(),
                m.getCurrentHP(),
                m.getMonsterType().getMaxHP(),
                m.isAlive()
        )).toList();

        return new ZoneResult(zoneStr, result);
    }

    private void spawnMonsters(String zoneStr, long zoneAddr, int count) {
        List<MonsterType> monsterTypes = monsterTypeRepo.findAll();
        if (monsterTypes.isEmpty()) return;

        List<Long> children = h3.cellToChildren(zoneAddr, ENTITY_RESOLUTION);
        if (children.isEmpty()) return;

        Set<String> occupied = new HashSet<>(mapMonsterRepo.findOccupiedCellsByZone(zoneStr));

        List<Long> available = children.stream()
                .filter(c -> !occupied.contains(h3.h3ToString(c)))
                .collect(Collectors.toList());

        if (available.isEmpty()) return;
        Collections.shuffle(available);

        Random rand = new Random();
        List<MapMonster> toSave = new ArrayList<>();

        for (int i = 0; i < count && i < available.size(); i++) {
            MonsterType mt = monsterTypes.get(rand.nextInt(monsterTypes.size()));
            String cellStr = h3.h3ToString(available.get(i));

            MapMonster monster = new MapMonster();
            monster.setId(UUID.randomUUID().toString());
            monster.setH3Zone(zoneStr);
            monster.setH3Index(cellStr);
            monster.setMonsterType(mt);
            monster.setCurrentHP(mt.getMaxHP());
            monster.setAlive(true);
            toSave.add(monster);
        }

        mapMonsterRepo.saveAll(toSave);
    }
}
