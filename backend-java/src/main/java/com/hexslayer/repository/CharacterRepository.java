package com.hexslayer.repository;

import com.hexslayer.model.Character;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface CharacterRepository extends JpaRepository<Character, String> {
    long countByPlayerIdAndIsAliveTrue(String playerId);
    List<Character> findByH3ZoneAndIsAliveTrue(String h3Zone);
}
