package services

import (
	"fmt"
	"math/rand"
	"time"

	"hexslayer/internal/db"
	"hexslayer/internal/models"

	"github.com/google/uuid"
	"github.com/ziprecruiter/h3-go/pkg/h3"
)

const MaxCharactersPerPlayer = 2

var firstNames = []string{
	"Shadow", "Storm", "Iron", "Frost", "Blaze",
	"Crimson", "Silent", "Dark", "Swift", "Wild",
}

var lastNames = []string{
	"Fang", "Blade", "Claw", "Strike", "Hunter",
	"Walker", "Slayer", "Reaper", "Bane", "Warden",
}

func randomName() string {
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randRangeInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func DeployCharacter(playerID, h3Zone string) (*models.Character, error) {
	var aliveCount int64
	db.DB.Model(&models.Character{}).
		Where("player_id = ? AND is_alive = true", playerID).
		Count(&aliveCount)

	if aliveCount >= MaxCharactersPerPlayer {
		return nil, fmt.Errorf("max %d alive characters allowed", MaxCharactersPerPlayer)
	}

	var monsterCount int64
	db.DB.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", h3Zone).
		Count(&monsterCount)

	if monsterCount == 0 {
		return nil, fmt.Errorf("no active monsters in zone %s", h3Zone)
	}

	zone, err := h3.NewCellFromString(h3Zone)
	if err != nil {
		return nil, fmt.Errorf("invalid h3_zone: %w", err)
	}
	cellStr := randomChildCell(h3Zone, zone)

	maxHP := randRangeInt(250, 350)
	char := models.Character{
		ID:              uuid.New().String(),
		PlayerID:        playerID,
		Name:            randomName(),
		H3Zone:          h3Zone,
		H3Index:         cellStr,
		BaseDamage:      randRangeInt(80, 120),
		DamageAmp:       randRange(1.0, 1.5),
		DamageReduction: randRange(0.05, 0.20),
		CritChance:      randRange(0.10, 0.25),
		CritMultiplier:  randRange(1.5, 2.0),
		HP:              maxHP,
		MaxHP:           maxHP,
		IsAlive:         true,
		DeployedAt:      time.Now(),
	}

	if err := db.DB.Create(&char).Error; err != nil {
		return nil, err
	}

	return &char, nil
}
