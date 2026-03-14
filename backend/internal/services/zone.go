package services

import (
	"math/rand"

	"hexslayer/internal/config"
	"hexslayer/internal/db"
	"hexslayer/internal/models"

	"github.com/google/uuid"

	h3light "github.com/ThingsIXFoundation/h3-light"
	"github.com/ziprecruiter/h3-go/pkg/h3"
)

// ZoneMonsterResponse is the lean monster data sent to the frontend.
type ZoneMonsterResponse struct {
	ID        string `json:"id"`
	H3Index   string `json:"h3_index"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	CurrentHP int    `json:"current_hp"`
	MaxHP     int    `json:"max_hp"`
	IsAlive   bool   `json:"is_alive"`
}

// GetOrCreateZoneMonsters computes the res-6 zone from lat/lng,
// ensures monsters are spawned up to cap, and returns all monsters in the zone.
func GetOrCreateZoneMonsters(lat, lng float64) (string, []ZoneMonsterResponse, error) {
	ll := h3.NewLatLng(lat, lng)
	zone, err := h3.NewCellFromLatLng(ll, config.ZoneResolution)
	if err != nil {
		return "", nil, err
	}
	zoneStr := zone.String()

	// Count living monsters in this zone
	var aliveCount int64
	db.DB.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Count(&aliveCount)

	// Only spawn when alive count drops below 20% of cap, then fill back to cap
	threshold := config.ZoneMonsterCap / 5
	toSpawn := config.ZoneMonsterCap - int(aliveCount)
	if int(aliveCount) < threshold {
		if err := spawnMonsters(zoneStr, zone, toSpawn); err != nil {
			return "", nil, err
		}
	}

	// Fetch all alive monsters in zone with their type
	var monsters []models.MapMonster
	if err := db.DB.Preload("MonsterType").
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Find(&monsters).Error; err != nil {
		return "", nil, err
	}

	// Map to lean response
	result := make([]ZoneMonsterResponse, len(monsters))
	for i, m := range monsters {
		result[i] = ZoneMonsterResponse{
			ID:        m.ID,
			H3Index:   m.H3Index,
			Type:      m.MonsterType.Name,
			Icon:      m.MonsterType.Icon,
			CurrentHP: m.CurrentHP,
			MaxHP:     m.MonsterType.MaxHP,
			IsAlive:   m.IsAlive,
		}
	}

	return zoneStr, result, nil
}

// randomChildCell generates a random res-12 cell within a res-6 zone
// by scattering random points near the zone center and verifying parentage.
// Uses h3-light for Cell→LatLon (ziprecruiter/h3-go lacks this).
func randomChildCell(zoneStr string, zone h3.Cell) string {
	// Get zone center via h3-light (only lib that has Cell→LatLon)
	lightCell := h3light.MustCellFromString(zoneStr)
	centerLat, centerLon := lightCell.LatLon()

	const spread = 0.036
	for {
		lat := centerLat + (rand.Float64()*2-1)*spread
		lon := centerLon + (rand.Float64()*2-1)*spread
		ll := h3.NewLatLng(lat, lon)
		child, err := h3.NewCellFromLatLng(ll, config.EntityResolution)
		if err != nil {
			continue
		}
		parent, err := child.Parent(config.ZoneResolution)
		if err != nil {
			continue
		}
		if parent == zone {
			return child.String()
		}
	}
}

func spawnMonsters(zoneStr string, zone h3.Cell, count int) error {
	var monsterTypes []models.MonsterType
	if err := db.DB.Find(&monsterTypes).Error; err != nil {
		return err
	}
	if len(monsterTypes) == 0 {
		return nil
	}

	monsters := make([]models.MapMonster, count)
	for i := 0; i < count; i++ {
		mt := monsterTypes[rand.Intn(len(monsterTypes))]
		cellStr := randomChildCell(zoneStr, zone)

		monsters[i] = models.MapMonster{
			ID:            uuid.New().String(),
			H3Zone:        zoneStr,
			H3Index:       cellStr,
			MonsterTypeID: mt.ID,
			CurrentHP:     mt.MaxHP,
			IsAlive:       true,
		}
	}

	return db.DB.CreateInBatches(monsters, 100).Error
}
