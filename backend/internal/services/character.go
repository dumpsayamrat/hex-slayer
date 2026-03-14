package services

import (
	"fmt"
	"math/rand"
	"time"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	"github.com/google/uuid"
	h3 "github.com/uber/h3-go/v4"
	"gorm.io/gorm"
)

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

type CharacterService struct {
	db *gorm.DB
}

func NewCharacterService(db *gorm.DB) *CharacterService {
	return &CharacterService{db: db}
}

func (s *CharacterService) Deploy(playerID, h3Zone string) (*models.Character, error) {
	var aliveCount int64
	s.db.Model(&models.Character{}).
		Where("player_id = ? AND is_alive = true", playerID).
		Count(&aliveCount)

	if aliveCount >= config.MaxCharactersAlive {
		return nil, fmt.Errorf("max %d alive characters allowed", config.MaxCharactersAlive)
	}

	var monsterCount int64
	s.db.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", h3Zone).
		Count(&monsterCount)

	if monsterCount == 0 {
		return nil, fmt.Errorf("no active monsters in zone %s", h3Zone)
	}

	zone := h3.CellFromString(h3Zone)
	if !h3.IsValidIndex(zone) {
		return nil, fmt.Errorf("invalid h3_zone: %s", h3Zone)
	}
	cellStr := randomChildCell(zone)

	maxHP := randRangeInt(500, 700)
	char := models.Character{
		ID:              uuid.New().String(),
		PlayerID:        playerID,
		Name:            randomName(),
		H3Zone:          h3Zone,
		H3Index:         cellStr,
		BaseDamage:      randRangeInt(30, 60),
		DamageAmp:       randRange(1.0, 1.3),
		DamageReduction: randRange(0.35, 0.55),
		CritChance:      randRange(0.10, 0.25),
		CritMultiplier:  randRange(1.5, 2.0),
		HP:              maxHP,
		MaxHP:           maxHP,
		IsAlive:         true,
		DeployedAt:      time.Now(),
	}

	if err := s.db.Create(&char).Error; err != nil {
		return nil, err
	}

	return &char, nil
}
