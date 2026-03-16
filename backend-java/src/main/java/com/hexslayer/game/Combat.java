package com.hexslayer.game;

import com.hexslayer.model.Character;
import com.hexslayer.model.MapMonster;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.repository.MapMonsterRepository;

import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ThreadLocalRandom;

public class Combat {

    public record CombatResult(int damage, boolean isCrit) {}

    public record Combatant(int baseDamage, double damageAmp, double damageReduction,
                             double critChance, double critMultiplier) {}

    public static Combatant fromCharacter(Character c) {
        return new Combatant(c.getBaseDamage(), c.getDamageAmp(), c.getDamageReduction(),
                c.getCritChance(), c.getCritMultiplier());
    }

    public static Combatant fromMonster(MapMonster m) {
        return new Combatant(m.getMonsterType().getBaseDamage(), m.getMonsterType().getDamageAmp(),
                m.getMonsterType().getDamageReduction(), m.getMonsterType().getCritChance(),
                m.getMonsterType().getCritMultiplier());
    }

    public static CombatResult attack(Combatant attacker, Combatant defender) {
        double rawDamage = attacker.baseDamage() * attacker.damageAmp();
        boolean isCrit = ThreadLocalRandom.current().nextDouble() < attacker.critChance();
        if (isCrit) {
            rawDamage *= attacker.critMultiplier();
        }
        double finalDamage = rawDamage * (1.0 - defender.damageReduction());
        return new CombatResult((int) Math.round(finalDamage), isCrit);
    }

    public static List<Map<String, Object>> doCombat(CharacterRepository charRepo,
                                                      MapMonsterRepository monsterRepo,
                                                      Character ch, MapMonster monster) {
        List<Map<String, Object>> logs = new ArrayList<>();
        Combatant charStats = fromCharacter(ch);
        Combatant monStats = fromMonster(monster);

        // Character attacks monster
        CombatResult hit = attack(charStats, monStats);
        monster.setCurrentHP(monster.getCurrentHP() - hit.damage());
        if (monster.getCurrentHP() <= 0) {
            monster.setCurrentHP(0);
            monster.setAlive(false);
        }
        monsterRepo.save(monster);

        Map<String, Object> log1 = new HashMap<>();
        log1.put("type", "combat_log");
        log1.put("attacker", ch.getName());
        log1.put("attacker_id", ch.getId());
        log1.put("defender", monster.getMonsterType().getName());
        log1.put("defender_id", monster.getId());
        log1.put("damage", hit.damage());
        log1.put("is_crit", hit.isCrit());
        log1.put("character_id", ch.getId());
        log1.put("character_hp", ch.getHp());
        log1.put("monster_id", monster.getId());
        log1.put("monster_hp", monster.getCurrentHP());
        logs.add(log1);

        // Monster attacks character (only if both still alive)
        if (monster.isAlive() && ch.isAlive()) {
            CombatResult hit2 = attack(monStats, charStats);
            ch.setHp(ch.getHp() - hit2.damage());
            if (ch.getHp() <= 0) {
                ch.setHp(0);
                ch.setAlive(false);
                ch.setDiedAt(Instant.now());
            }
            charRepo.save(ch);

            Map<String, Object> log2 = new HashMap<>();
            log2.put("type", "combat_log");
            log2.put("attacker", monster.getMonsterType().getName());
            log2.put("attacker_id", monster.getId());
            log2.put("defender", ch.getName());
            log2.put("defender_id", ch.getId());
            log2.put("damage", hit2.damage());
            log2.put("is_crit", hit2.isCrit());
            log2.put("character_id", ch.getId());
            log2.put("character_hp", ch.getHp());
            log2.put("monster_id", monster.getId());
            log2.put("monster_hp", monster.getCurrentHP());
            logs.add(log2);
        }

        return logs;
    }
}
