package com.hexslayer.game;

import com.hexslayer.model.Character;
import com.hexslayer.model.MapMonster;
import com.uber.h3core.H3Core;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ThreadLocalRandom;

import static com.hexslayer.game.GameConfig.*;

public class Movement {

    private final H3Core h3;

    public Movement(H3Core h3) {
        this.h3 = h3;
    }

    public String wander(Character ch) {
        long charCell = h3.stringToH3(ch.getH3Index());
        var ll = h3.cellToLatLng(charCell);
        double lat = ll.lat;
        double lng = ll.lng;

        ThreadLocalRandom rand = ThreadLocalRandom.current();

        // Drift the bearing gradually
        double drift = (rand.nextDouble() * 2 - 1) * WANDER_BEARING_DRIFT_MAX;
        if (rand.nextDouble() < WANDER_BIG_TURN_CHANCE) {
            drift = (rand.nextDouble() * 2 - 1) * WANDER_BIG_TURN_MAX;
        }
        ch.setWanderBearing(normalizeBearing(ch.getWanderBearing() + drift));

        // Move one step in the bearing direction
        double bearingRad = ch.getWanderBearing() * Math.PI / 180.0;
        double newLat = lat + STEP_DISTANCE_DEG * Math.cos(bearingRad);
        double newLng = lng + STEP_DISTANCE_DEG * Math.sin(bearingRad) / Math.cos(lat * Math.PI / 180.0);

        long newCell = h3.latLngToCell(newLat, newLng, ENTITY_RESOLUTION);

        // Check still within zone
        long parent = h3.cellToParent(newCell, ZONE_RESOLUTION);
        if (!h3.h3ToString(parent).equals(ch.getH3Zone())) {
            // Hit zone boundary — turn around
            ch.setWanderBearing(normalizeBearing(ch.getWanderBearing() + 150 + rand.nextDouble() * 60));
            return ch.getH3Index();
        }

        return h3.h3ToString(newCell);
    }

    public record MoveResult(String h3Index, int distance) {}

    public MoveResult moveToward(Character ch, MapMonster target) {
        long charCell = h3.stringToH3(ch.getH3Index());
        long targetCell = h3.stringToH3(target.getH3Index());

        var charLL = h3.cellToLatLng(charCell);
        var targetLL = h3.cellToLatLng(targetCell);

        // Compute bearing toward target
        double dLng = (targetLL.lng - charLL.lng) * Math.PI / 180.0;
        double charLatRad = charLL.lat * Math.PI / 180.0;
        double targetLatRad = targetLL.lat * Math.PI / 180.0;

        double y = Math.sin(dLng) * Math.cos(targetLatRad);
        double x = Math.cos(charLatRad) * Math.sin(targetLatRad) - Math.sin(charLatRad) * Math.cos(targetLatRad) * Math.cos(dLng);
        double bearing = Math.atan2(y, x) * 180.0 / Math.PI;

        ch.setWanderBearing(normalizeBearing(bearing));

        double bearingRad = bearing * Math.PI / 180.0;
        double newLat = charLL.lat + STEP_DISTANCE_DEG * Math.cos(bearingRad);
        double newLng = charLL.lng + STEP_DISTANCE_DEG * Math.sin(bearingRad) / Math.cos(charLL.lat * Math.PI / 180.0);

        long newCell = h3.latLngToCell(newLat, newLng, ENTITY_RESOLUTION);

        // Check still in zone
        long parent = h3.cellToParent(newCell, ZONE_RESOLUTION);
        if (!h3.h3ToString(parent).equals(ch.getH3Zone())) {
            return new MoveResult(ch.getH3Index(), 999);
        }

        int dist;
        try {
            dist = (int) h3.gridDistance(newCell, targetCell);
        } catch (Exception e) {
            dist = 999;
        }

        return new MoveResult(h3.h3ToString(newCell), dist);
    }

    public MapMonster findNearestFreeMonster(Character ch, List<MapMonster> monsters, Map<String, Boolean> engaged) {
        long charCell = h3.stringToH3(ch.getH3Index());
        if (!h3.isValidCell(charCell)) return null;

        // Build monster lookup by h3_index (only alive + not engaged)
        Map<String, List<MapMonster>> byCell = new java.util.HashMap<>();
        for (MapMonster m : monsters) {
            if (!m.isAlive() || engaged.containsKey(m.getId())) continue;
            byCell.computeIfAbsent(m.getH3Index(), k -> new java.util.ArrayList<>()).add(m);
        }

        // Scan outward ring by ring
        for (int k = 1; k <= DETECTION_RADIUS; k++) {
            try {
                List<Long> ring = h3.gridRingUnsafe(charCell, k);
                for (long cell : ring) {
                    String cellStr = h3.h3ToString(cell);
                    List<MapMonster> ms = byCell.get(cellStr);
                    if (ms != null && !ms.isEmpty()) {
                        return ms.get(0);
                    }
                }
            } catch (Exception e) {
                // gridRingUnsafe can fail for pentagons, skip
            }
        }

        return null;
    }

    public static double normalizeBearing(double b) {
        b = b % 360;
        if (b < 0) b += 360;
        return b;
    }
}
