package services

import (
	"log"
	"math/rand"

	"hexslayer/internal/apperr"
	"hexslayer/internal/config"
	"hexslayer/internal/dto"
	"hexslayer/internal/models"

	"github.com/google/uuid"
	h3 "github.com/uber/h3-go/v4"
	"gorm.io/gorm"
)

type ZoneService struct {
	db *gorm.DB
}

func NewZoneService(db *gorm.DB) *ZoneService {
	return &ZoneService{db: db}
}

// GetOrCreateMonsters computes the res-6 zone from lat/lng,
// ensures monsters are spawned up to cap, and returns all monsters in the zone.
func (s *ZoneService) GetOrCreateMonsters(lat, lng float64) (string, []dto.ZoneMonsterResponse, error) {
	ll := h3.NewLatLng(lat, lng)
	zone, err := h3.LatLngToCell(ll, config.ZoneResolution)
	if err != nil {
		return "", nil, apperr.NewValidation("invalid coordinates: lat=%f lng=%f", lat, lng)
	}
	zoneStr := zone.String()

	// Count living monsters in this zone
	var aliveCount int64
	s.db.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Count(&aliveCount)

	// Only spawn when alive count drops below 20% of cap, then fill back to cap
	threshold := config.ZoneMonsterCap / 5
	toSpawn := config.ZoneMonsterCap - int(aliveCount)

	log.Printf("zone %s: aliveCount=%d threshold=%d toSpawn=%d", zoneStr, aliveCount, threshold, toSpawn)

	if int(aliveCount) < threshold {
		if err := s.spawnMonsters(zoneStr, zone, toSpawn); err != nil {
			return "", nil, err
		}
	}

	// Fetch all alive monsters in zone with their type
	var monsters []models.MapMonster
	if err := s.db.Preload("MonsterType").
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Find(&monsters).Error; err != nil {
		return "", nil, err
	}

	log.Printf("zone %s: returning %d monsters", zoneStr, len(monsters))

	// Map to lean response
	result := make([]dto.ZoneMonsterResponse, len(monsters))
	for i, m := range monsters {
		result[i] = dto.ZoneMonsterResponse{
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

// randomChildCell picks a random res-12 cell within a zone, biased toward the center.
// 70% chance to pick from the inner half, 30% from the outer half.
func randomChildCell(zone h3.Cell) string {
	inner, outer := getAvailableCells(zone)
	if len(inner) == 0 && len(outer) == 0 {
		return zone.String()
	}

	useInner := len(inner) > 0 && (len(outer) == 0 || rand.Float64() < 0.7)
	if useInner {
		return inner[rand.Intn(len(inner))].String()
	}
	return outer[rand.Intn(len(outer))].String()
}

// getAvailableCells returns all res-12 children of a zone, split into
// inner (center half) and outer (edge half), excluding cells that already
// have alive monsters.
func getAvailableCells(zone h3.Cell) (inner, outer []h3.Cell) {
	// Get all res-12 children of this res-6 zone
	allChildren, err := h3.UncompactCells([]h3.Cell{zone}, config.EntityResolution)
	if err != nil || len(allChildren) == 0 {
		return nil, nil
	}

	// Get zone center at res-12 for distance calculation
	centerLL, err := h3.CellToLatLng(zone)
	if err != nil {
		return nil, nil
	}
	centerCell, err := h3.LatLngToCell(centerLL, config.EntityResolution)
	if err != nil {
		return nil, nil
	}

	// Find max distance to determine midpoint
	maxDist := 0
	type cellDist struct {
		cell h3.Cell
		dist int
	}
	cells := make([]cellDist, 0, len(allChildren))
	for _, c := range allChildren {
		d, err := h3.GridDistance(centerCell, c)
		if err != nil {
			continue
		}
		cells = append(cells, cellDist{cell: c, dist: d})
		if d > maxDist {
			maxDist = d
		}
	}

	mid := maxDist / 2
	for _, cd := range cells {
		if cd.dist <= mid {
			inner = append(inner, cd.cell)
		} else {
			outer = append(outer, cd.cell)
		}
	}
	return inner, outer
}

func (s *ZoneService) spawnMonsters(zoneStr string, zone h3.Cell, count int) error {
	var monsterTypes []models.MonsterType
	if err := s.db.Find(&monsterTypes).Error; err != nil {
		return err
	}
	if len(monsterTypes) == 0 {
		return nil
	}

	inner, outer := getAvailableCells(zone)
	if len(inner) == 0 && len(outer) == 0 {
		return nil
	}

	// Load occupied cells (alive monsters) to filter them out
	var occupied []string
	s.db.Model(&models.MapMonster{}).
		Where("h3_zone = ? AND is_alive = true", zoneStr).
		Pluck("h3_index", &occupied)
	occupiedSet := make(map[string]bool, len(occupied))
	for _, idx := range occupied {
		occupiedSet[idx] = true
	}

	// Filter out occupied cells
	filterAvailable := func(cells []h3.Cell) []h3.Cell {
		available := make([]h3.Cell, 0, len(cells))
		for _, c := range cells {
			if !occupiedSet[c.String()] {
				available = append(available, c)
			}
		}
		return available
	}
	innerAvail := filterAvailable(inner)
	outerAvail := filterAvailable(outer)

	if len(innerAvail) == 0 && len(outerAvail) == 0 {
		return nil
	}

	// Shuffle both pools for random picking without replacement
	rand.Shuffle(len(innerAvail), func(i, j int) { innerAvail[i], innerAvail[j] = innerAvail[j], innerAvail[i] })
	rand.Shuffle(len(outerAvail), func(i, j int) { outerAvail[i], outerAvail[j] = outerAvail[j], outerAvail[i] })

	innerIdx, outerIdx := 0, 0
	monsters := make([]models.MapMonster, 0, count)
	for i := 0; i < count; i++ {
		mt := monsterTypes[rand.Intn(len(monsterTypes))]

		// 50/50 inner vs outer, fallback if one is exhausted
		var cellStr string
		canInner := innerIdx < len(innerAvail)
		canOuter := outerIdx < len(outerAvail)
		if !canInner && !canOuter {
			break
		}

		useInner := canInner && (!canOuter || rand.Float64() < 0.5)
		if useInner {
			cellStr = innerAvail[innerIdx].String()
			innerIdx++
		} else {
			cellStr = outerAvail[outerIdx].String()
			outerIdx++
		}

		monsters = append(monsters, models.MapMonster{
			ID:            uuid.New().String(),
			H3Zone:        zoneStr,
			H3Index:       cellStr,
			MonsterTypeID: mt.ID,
			CurrentHP:     mt.MaxHP,
			IsAlive:       true,
		})
	}

	return s.db.CreateInBatches(monsters, 100).Error
}
