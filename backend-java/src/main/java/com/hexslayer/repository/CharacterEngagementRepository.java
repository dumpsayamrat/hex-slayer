package com.hexslayer.repository;

import com.hexslayer.model.CharacterEngagement;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface CharacterEngagementRepository extends JpaRepository<CharacterEngagement, String> {
    List<CharacterEngagement> findByCharacterIdIn(List<String> characterIds);
}
