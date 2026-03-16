package com.hexslayer.model;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "monster_types")
public class MonsterType {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String name;

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

    @Column(name = "max_hp", nullable = false)
    private int maxHP;

    @Column(nullable = false)
    private String icon;

    @Column(name = "created_at")
    private Instant createdAt;

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }

    public MonsterType() {}

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
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
    public int getMaxHP() { return maxHP; }
    public void setMaxHP(int maxHP) { this.maxHP = maxHP; }
    public String getIcon() { return icon; }
    public void setIcon(String icon) { this.icon = icon; }
    public Instant getCreatedAt() { return createdAt; }
    public void setCreatedAt(Instant createdAt) { this.createdAt = createdAt; }
}
