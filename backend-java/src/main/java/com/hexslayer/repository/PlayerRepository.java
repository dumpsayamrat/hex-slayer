package com.hexslayer.repository;

import com.hexslayer.model.Player;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;

public interface PlayerRepository extends JpaRepository<Player, String> {
    Optional<Player> findBySessionToken(String sessionToken);
}
