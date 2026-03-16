package com.hexslayer.model;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "map_monsters")
public class MapMonster {

    @Id
    private String id;

    @Column(name = "h3_zone", nullable = false)
    private String h3Zone;

    @Column(name = "h3_index", nullable = false)
    private String h3Index;

    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "monster_type_id", nullable = false)
    private MonsterType monsterType;

    @Column(name = "current_hp", nullable = false)
    private int currentHP;

    @Column(name = "is_alive", nullable = false)
    private boolean isAlive = true;

    @Column(name = "respawn_at")
    private Instant respawnAt;

    @Column(name = "created_at")
    private Instant createdAt;

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }

    public MapMonster() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getH3Zone() { return h3Zone; }
    public void setH3Zone(String h3Zone) { this.h3Zone = h3Zone; }
    public String getH3Index() { return h3Index; }
    public void setH3Index(String h3Index) { this.h3Index = h3Index; }
    public MonsterType getMonsterType() { return monsterType; }
    public void setMonsterType(MonsterType monsterType) { this.monsterType = monsterType; }
    public int getCurrentHP() { return currentHP; }
    public void setCurrentHP(int currentHP) { this.currentHP = currentHP; }
    public boolean isAlive() { return isAlive; }
    public void setAlive(boolean alive) { isAlive = alive; }
    public Instant getRespawnAt() { return respawnAt; }
    public void setRespawnAt(Instant respawnAt) { this.respawnAt = respawnAt; }
    public Instant getCreatedAt() { return createdAt; }
    public void setCreatedAt(Instant createdAt) { this.createdAt = createdAt; }
}
