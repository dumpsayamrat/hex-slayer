package com.hexslayer.model;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "characters")
public class Character {

    @Id
    private String id;

    @Column(name = "player_id", nullable = false)
    private String playerId;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "player_id", insertable = false, updatable = false)
    private Player player;

    @Column(nullable = false)
    private String name;

    @Column(name = "h3_zone", nullable = false)
    private String h3Zone;

    @Column(name = "h3_index", nullable = false)
    private String h3Index;

    @Column(nullable = false)
    private int hp;

    @Column(name = "max_hp", nullable = false)
    private int maxHP;

    @Column(name = "base_damage", nullable = false)
    private int baseDamage;

    @Column(name = "damage_amp", nullable = false)
    private double damageAmp;

    @Column(name = "damage_reduction", nullable = false)
    private double damageReduction;

    @Column(name = "crit_chance", nullable = false)
    private double critChance;

    @Column(name = "crit_multiplier", nullable = false)
    private double critMultiplier;

    @Column(nullable = false)
    private int kills = 0;

    @Column(name = "is_alive", nullable = false)
    private boolean isAlive = true;

    @Column(name = "wander_bearing", nullable = false)
    private double wanderBearing = 0;

    @Column(name = "target_monster_id")
    private String targetMonsterId;

    @Column(name = "deployed_at")
    private Instant deployedAt;

    @Column(name = "died_at")
    private Instant diedAt;

    @PrePersist
    protected void onCreate() {
        deployedAt = Instant.now();
    }

    public Character() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getPlayerId() { return playerId; }
    public void setPlayerId(String playerId) { this.playerId = playerId; }
    public Player getPlayer() { return player; }
    public void setPlayer(Player player) { this.player = player; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public String getH3Zone() { return h3Zone; }
    public void setH3Zone(String h3Zone) { this.h3Zone = h3Zone; }
    public String getH3Index() { return h3Index; }
    public void setH3Index(String h3Index) { this.h3Index = h3Index; }
    public int getHp() { return hp; }
    public void setHp(int hp) { this.hp = hp; }
    public int getMaxHP() { return maxHP; }
    public void setMaxHP(int maxHP) { this.maxHP = maxHP; }
    public int getBaseDamage() { return baseDamage; }
    public void setBaseDamage(int baseDamage) { this.baseDamage = baseDamage; }
    public double getDamageAmp() { return damageAmp; }
    public void setDamageAmp(double damageAmp) { this.damageAmp = damageAmp; }
    public double getDamageReduction() { return damageReduction; }
    public void setDamageReduction(double damageReduction) { this.damageReduction = damageReduction; }
    public double getCritChance() { return critChance; }
    public void setCritChance(double critChance) { this.critChance = critChance; }
    public double getCritMultiplier() { return critMultiplier; }
    public void setCritMultiplier(double critMultiplier) { this.critMultiplier = critMultiplier; }
    public int getKills() { return kills; }
    public void setKills(int kills) { this.kills = kills; }
    public boolean isAlive() { return isAlive; }
    public void setAlive(boolean alive) { isAlive = alive; }
    public double getWanderBearing() { return wanderBearing; }
    public void setWanderBearing(double wanderBearing) { this.wanderBearing = wanderBearing; }
    public String getTargetMonsterId() { return targetMonsterId; }
    public void setTargetMonsterId(String targetMonsterId) { this.targetMonsterId = targetMonsterId; }
    public Instant getDeployedAt() { return deployedAt; }
    public void setDeployedAt(Instant deployedAt) { this.deployedAt = deployedAt; }
    public Instant getDiedAt() { return diedAt; }
    public void setDiedAt(Instant diedAt) { this.diedAt = diedAt; }
}
