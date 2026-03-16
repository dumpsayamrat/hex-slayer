package com.hexslayer.service;

import com.hexslayer.model.Player;
import com.hexslayer.repository.PlayerRepository;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
public class PlayerService {

    private final PlayerRepository playerRepo;

    public PlayerService(PlayerRepository playerRepo) {
        this.playerRepo = playerRepo;
    }

    public Player create() {
        Player player = new Player();
        player.setId(UUID.randomUUID().toString());
        player.setSessionToken(UUID.randomUUID().toString());
        return playerRepo.save(player);
    }
}
