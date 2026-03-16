package com.hexslayer.game;

public final class GameConfig {
    public static final int ZONE_MONSTER_CAP = 300;
    public static final int MAX_CHARACTERS_ALIVE = 2;
    public static final int TICK_INTERVAL_SECONDS = 2;
    public static final int ZONE_RESOLUTION = 6;
    public static final int ENTITY_RESOLUTION = 12;
    public static final int GRID_DISK_RADIUS = 2;
    public static final int ZONE_MAX_DURATION_MINS = 30;
    public static final double STEP_DISTANCE_DEG = 0.0005;

    // Wander settings
    public static final double WANDER_BEARING_DRIFT_MAX = 30.0;
    public static final double WANDER_BIG_TURN_CHANCE = 0.05;
    public static final double WANDER_BIG_TURN_MAX = 90.0;

    // Hunting settings
    public static final int DETECTION_RADIUS = 25;

    private GameConfig() {}
}
