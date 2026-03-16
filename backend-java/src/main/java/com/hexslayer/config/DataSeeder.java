package com.hexslayer.config;

import com.hexslayer.model.MonsterType;
import com.hexslayer.repository.MonsterTypeRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

@Component
public class DataSeeder implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(DataSeeder.class);
    private final MonsterTypeRepository monsterTypeRepo;

    public DataSeeder(MonsterTypeRepository monsterTypeRepo) {
        this.monsterTypeRepo = monsterTypeRepo;
    }

    @Override
    public void run(String... args) {
        if (monsterTypeRepo.count() > 0) {
            log.info("monster_types already seeded, skipping");
            return;
        }

        monsterTypeRepo.save(monsterType("Slime", 15, 1.0, 0.05, 0.05, 1.3, 80, "slime.png"));
        monsterTypeRepo.save(monsterType("Goblin", 25, 1.1, 0.08, 0.10, 1.5, 120, "goblin.png"));
        monsterTypeRepo.save(monsterType("Orc", 35, 1.2, 0.15, 0.12, 1.6, 200, "orc.png"));
        monsterTypeRepo.save(monsterType("Troll", 45, 1.3, 0.20, 0.15, 1.8, 300, "troll.png"));
        monsterTypeRepo.save(monsterType("Dragon", 35, 1.3, 0.25, 0.15, 1.8, 500, "dragon.png"));

        log.info("seeded 5 monster types");
    }

    private MonsterType monsterType(String name, int baseDamage, double damageAmp,
            double damageReduction, double critChance, double critMultiplier, int maxHP, String icon) {
        MonsterType mt = new MonsterType();
        mt.setName(name);
        mt.setBaseDamage(baseDamage);
        mt.setDamageAmp(damageAmp);
        mt.setDamageReduction(damageReduction);
        mt.setCritChance(critChance);
        mt.setCritMultiplier(critMultiplier);
        mt.setMaxHP(maxHP);
        mt.setIcon(icon);
        return mt;
    }
}
