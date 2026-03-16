package com.hexslayer.model;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "character_engagements")
public class CharacterEngagement {

    @Id
    private String id;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "character_id", nullable = false, unique = true)
    private Character character;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "monster_id", nullable = false)
    private MapMonster monster;

    @Column(name = "engaged_at")
    private Instant engagedAt;

    @PrePersist
    protected void onCreate() {
        engagedAt = Instant.now();
    }

    public CharacterEngagement() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public Character getCharacter() { return character; }
    public void setCharacter(Character character) { this.character = character; }
    public MapMonster getMonster() { return monster; }
    public void setMonster(MapMonster monster) { this.monster = monster; }
    public Instant getEngagedAt() { return engagedAt; }
    public void setEngagedAt(Instant engagedAt) { this.engagedAt = engagedAt; }
}
