package com.hexslayer.controller;

import com.hexslayer.game.GameEngine;
import com.hexslayer.model.Character;
import com.hexslayer.model.Player;
import com.hexslayer.service.CharacterService;
import com.hexslayer.ws.WebSocketHub;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/character")
public class CharacterController {

    private final CharacterService characterService;
    private final GameEngine engine;
    private final WebSocketHub hub;

    public CharacterController(CharacterService characterService, GameEngine engine, WebSocketHub hub) {
        this.characterService = characterService;
        this.engine = engine;
        this.hub = hub;
    }

    @PostMapping("/deploy")
    public ResponseEntity<Map<String, Object>> deploy(
            @RequestBody Map<String, String> body,
            HttpServletRequest request) {

        Player player = (Player) request.getAttribute("player");
        if (player == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "unauthorized"));
        }

        String h3Zone = body.get("h3_zone");
        if (h3Zone == null || h3Zone.isEmpty()) {
            return ResponseEntity.badRequest().body(Map.of("error", "h3_zone is required"));
        }

        Character ch = characterService.deploy(player.getId(), h3Zone);

        // Start zone tick loop + broadcast to subscribers
        engine.ensureZoneLoop(ch.getH3Zone());
        hub.broadcast("zone:" + ch.getH3Zone(), Map.of(
                "type", "char_deployed",
                "id", ch.getId(),
                "name", ch.getName(),
                "player_id", ch.getPlayerId(),
                "h3_zone", ch.getH3Zone(),
                "h3_index", ch.getH3Index(),
                "hp", ch.getHp(),
                "max_hp", ch.getMaxHP()
        ));

        Map<String, Object> resp = new java.util.HashMap<>();
        resp.put("id", ch.getId());
        resp.put("name", ch.getName());
        resp.put("h3_zone", ch.getH3Zone());
        resp.put("h3_index", ch.getH3Index());
        resp.put("hp", ch.getHp());
        resp.put("max_hp", ch.getMaxHP());
        resp.put("base_damage", ch.getBaseDamage());
        resp.put("damage_amp", ch.getDamageAmp());
        resp.put("damage_reduction", ch.getDamageReduction());
        resp.put("crit_chance", ch.getCritChance());
        resp.put("crit_multiplier", ch.getCritMultiplier());

        return ResponseEntity.status(HttpStatus.CREATED).body(resp);
    }
}
