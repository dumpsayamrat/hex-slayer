package com.hexslayer.dto;

public record ZoneMonsterResponse(
        String id,
        String h3Index,
        String type,
        String icon,
        int currentHp,
        int maxHp,
        boolean isAlive
) {}
