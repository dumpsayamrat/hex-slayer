package com.hexslayer.repository;

import com.hexslayer.model.MapMonster;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;

import java.util.List;

public interface MapMonsterRepository extends JpaRepository<MapMonster, String> {
    long countByH3ZoneAndIsAliveTrue(String h3Zone);
    List<MapMonster> findByH3ZoneAndIsAliveTrue(String h3Zone);

    @Query("SELECT m.h3Index FROM MapMonster m WHERE m.h3Zone = :h3Zone AND m.isAlive = true")
    List<String> findOccupiedCellsByZone(String h3Zone);
}
