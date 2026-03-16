package com.hexslayer.service;

import com.hexslayer.exception.ValidationException;
import com.hexslayer.model.Character;
import com.hexslayer.model.MapMonster;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.repository.MapMonsterRepository;
import com.uber.h3core.H3Core;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.util.*;
import java.util.concurrent.ThreadLocalRandom;
import java.util.stream.Collectors;

@Service
public class CharacterService {

    private static final int MAX_CHARACTERS_ALIVE = 2;
    private static final int ENTITY_RESOLUTION = 12;

    private static final String[] FIRST_NAMES = {
            "Shadow", "Storm", "Iron", "Frost", "Blaze",
            "Crimson", "Silent", "Dark", "Swift", "Wild"
    };
    private static final String[] LAST_NAMES = {
            "Fang", "Blade", "Claw", "Strike", "Hunter",
            "Walker", "Slayer", "Reaper", "Bane", "Warden"
    };

    private final CharacterRepository characterRepo;
    private final MapMonsterRepository mapMonsterRepo;
    private final H3Core h3;

    public CharacterService(CharacterRepository characterRepo, MapMonsterRepository mapMonsterRepo) {
        this.characterRepo = characterRepo;
        this.mapMonsterRepo = mapMonsterRepo;
        try {
            this.h3 = H3Core.newInstance();
        } catch (IOException e) {
            throw new RuntimeException("Failed to initialize H3", e);
        }
    }

    public Character deploy(String playerId, String h3Zone) {
        long aliveCount = characterRepo.countByPlayerIdAndIsAliveTrue(playerId);
        if (aliveCount >= MAX_CHARACTERS_ALIVE) {
            throw new ValidationException("max " + MAX_CHARACTERS_ALIVE + " alive characters allowed");
        }

        long monsterCount = mapMonsterRepo.countByH3ZoneAndIsAliveTrue(h3Zone);
        if (monsterCount == 0) {
            throw new ValidationException("no active monsters in zone " + h3Zone);
        }

        long zoneAddr = h3.stringToH3(h3Zone);
        if (!h3.isValidCell(zoneAddr)) {
            throw new ValidationException("invalid h3_zone: " + h3Zone);
        }

        List<Long> children = h3.cellToChildren(zoneAddr, ENTITY_RESOLUTION);
        String cellStr = h3.h3ToString(children.get(ThreadLocalRandom.current().nextInt(children.size())));

        ThreadLocalRandom rand = ThreadLocalRandom.current();
        int maxHP = rand.nextInt(500, 701);

        Character ch = new Character();
        ch.setId(UUID.randomUUID().toString());
        ch.setPlayerId(playerId);
        ch.setName(randomName());
        ch.setH3Zone(h3Zone);
        ch.setH3Index(cellStr);
        ch.setBaseDamage(rand.nextInt(30, 61));
        ch.setDamageAmp(randRange(1.0, 1.3));
        ch.setDamageReduction(randRange(0.35, 0.55));
        ch.setCritChance(randRange(0.10, 0.25));
        ch.setCritMultiplier(randRange(1.5, 2.0));
        ch.setHp(maxHP);
        ch.setMaxHP(maxHP);
        ch.setAlive(true);

        return characterRepo.save(ch);
    }

    private String randomName() {
        ThreadLocalRandom rand = ThreadLocalRandom.current();
        return FIRST_NAMES[rand.nextInt(FIRST_NAMES.length)] + " " + LAST_NAMES[rand.nextInt(LAST_NAMES.length)];
    }

    private double randRange(double min, double max) {
        return min + ThreadLocalRandom.current().nextDouble() * (max - min);
    }
}
