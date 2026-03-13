package services

import (
	"math/rand"

	"hexslayer/internal/db"
	"hexslayer/internal/models"

	"github.com/google/uuid"

	h3light "github.com/ThingsIXFoundation/h3-light"
)

const (
	ZoneResolution   = 6
	EntityResolution = 12
	ZoneMonsterCap   = 300
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
	// Convert lat/lng to res-6 H3 cell
	zone := h3light.LatLonToCell(lat, lng, ZoneResolution)
	zoneStr := zone.String()

	// Count living monsters in this zone
	var aliveCount int64
	db.DB.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Count(&aliveCount)

	// Only spawn when alive count drops below 20% of cap, then fill back to cap
	threshold := ZoneMonsterCap / 5
	toSpawn := ZoneMonsterCap - int(aliveCount)
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
func randomChildCell(zone h3light.Cell) h3light.Cell {
	centerLat, centerLon := zone.LatLon()

	// Res-6 cell is ~3.6km edge. Scatter within ~4km radius (~0.036 degrees).
	const spread = 0.036
	for {
		lat := centerLat + (rand.Float64()*2-1)*spread
		lon := centerLon + (rand.Float64()*2-1)*spread
		child := h3light.LatLonToCell(lat, lon, EntityResolution)
		if child.Parent(ZoneResolution) == zone {
			return child
		}
	}
}

func spawnMonsters(zoneStr string, zone h3light.Cell, count int) error {
	// Get all monster types for random selection
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
		cell := randomChildCell(zone)

		monsters[i] = models.MapMonster{
			ID:            uuid.New().String(),
			H3Zone:        zoneStr,
			H3Index:       cell.String(),
			MonsterTypeID: mt.ID,
			CurrentHP:     mt.MaxHP,
			IsAlive:       true,
		}
	}

	// Batch insert
	return db.DB.CreateInBatches(monsters, 100).Error
}
