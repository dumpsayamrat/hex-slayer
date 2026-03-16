package com.hexslayer.controller;

import com.hexslayer.model.Player;
import com.hexslayer.service.PlayerService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/player")
public class PlayerController {

    private final PlayerService playerService;

    public PlayerController(PlayerService playerService) {
        this.playerService = playerService;
    }

    @PostMapping("/init")
    public ResponseEntity<Map<String, String>> initPlayer() {
        Player player = playerService.create();
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of(
                "playerId", player.getId(),
                "sessionToken", player.getSessionToken(),
                "name", player.getName()
        ));
    }
}
