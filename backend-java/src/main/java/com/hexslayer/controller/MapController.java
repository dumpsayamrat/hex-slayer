package com.hexslayer.controller;

import com.hexslayer.dto.ZoneMonsterResponse;
import com.hexslayer.model.Character;
import com.hexslayer.repository.CharacterRepository;
import com.hexslayer.service.ZoneService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/map")
public class MapController {

    private final ZoneService zoneService;
    private final CharacterRepository characterRepo;

    public MapController(ZoneService zoneService, CharacterRepository characterRepo) {
        this.zoneService = zoneService;
        this.characterRepo = characterRepo;
    }

    @GetMapping("/zones")
    public ResponseEntity<Map<String, Object>> getZones(
            @RequestParam double lat,
            @RequestParam double lng) {

        ZoneService.ZoneResult result = zoneService.getOrCreateMonsters(lat, lng);

        List<Character> chars = characterRepo.findByH3ZoneAndIsAliveTrue(result.h3Zone());
        List<Map<String, Object>> charData = chars.stream().map(ch -> Map.<String, Object>of(
                "id", ch.getId(),
                "name", ch.getName(),
                "hp", ch.getHp(),
                "max_hp", ch.getMaxHP(),
                "player_id", ch.getPlayer().getId(),
                "h3_index", ch.getH3Index()
        )).toList();

        return ResponseEntity.ok(Map.of(
                "h3_zone", result.h3Zone(),
                "monsters", result.monsters(),
                "characters", charData
        ));
    }
}
