package com.hexslayer.model;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "players")
public class Player {

    @Id
    private String id;

    @Column(name = "session_token", nullable = false, unique = true)
    private String sessionToken;

    @Column(nullable = false)
    private String name = "Adventurer";

    @Column(name = "created_at")
    private Instant createdAt;

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }

    public Player() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getSessionToken() { return sessionToken; }
    public void setSessionToken(String sessionToken) { this.sessionToken = sessionToken; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public Instant getCreatedAt() { return createdAt; }
    public void setCreatedAt(Instant createdAt) { this.createdAt = createdAt; }
}
