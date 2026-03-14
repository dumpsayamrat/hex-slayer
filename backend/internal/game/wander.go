package game

import (
	"math"
	"math/rand"

	"hexslayer/internal/config"
	"hexslayer/internal/db"
	"hexslayer/internal/models"

	h3light "github.com/ThingsIXFoundation/h3-light"
	"github.com/ziprecruiter/h3-go/pkg/h3"
)

const (
	// ~50m per step — covers more ground while wandering
	stepDistanceDeg = 0.0005
)

// wander moves a character one step in its current bearing direction,
// with gradual random drift for smooth, human-like curves.
// Updates char.WanderBearing in place. Returns the new H3 index.
func wander(char *models.Character) string {
	// Get current position via h3-light
	lightCell := h3light.MustCellFromString(char.H3Index)
	lat, lng := lightCell.LatLon()

	// Drift the bearing gradually
	drift := (rand.Float64()*2 - 1) * config.WanderBearingDriftMax
	// Occasionally make a bigger turn
	if rand.Float64() < config.WanderBigTurnChance {
		drift = (rand.Float64()*2 - 1) * config.WanderBigTurnMax
	}
	char.WanderBearing = normalizeBearing(char.WanderBearing + drift)

	// Move one step in the bearing direction
	bearingRad := char.WanderBearing * math.Pi / 180.0
	newLat := lat + stepDistanceDeg*math.Cos(bearingRad)
	newLng := lng + stepDistanceDeg*math.Sin(bearingRad)/math.Cos(lat*math.Pi/180.0)

	// Convert to H3 cell
	ll := h3.NewLatLng(newLat, newLng)
	newCell, err := h3.NewCellFromLatLng(ll, config.EntityResolution)
	if err != nil {
		return char.H3Index
	}

	// Check still within zone
	parent, err := newCell.Parent(config.ZoneResolution)
	if err != nil || parent.String() != char.H3Zone {
		// Hit zone boundary — turn around (reverse + some random)
		char.WanderBearing = normalizeBearing(char.WanderBearing + 150 + rand.Float64()*60)
		return char.H3Index
	}

	return newCell.String()
}

// wanderAndEmit performs a wander step, persists the move to DB,
// and returns events to broadcast. Used after kills so the character
// walks away before scanning for the next fight.
func wanderAndEmit(char *models.Character) []map[string]interface{} {
	newIndex := wander(char)
	if newIndex == char.H3Index {
		return nil
	}
	char.H3Index = newIndex
	db.DB.Model(char).Updates(map[string]interface{}{
		"h3_index":          newIndex,
		"wander_bearing":    char.WanderBearing,
		"target_monster_id": nil,
	})
	return []map[string]interface{}{
		{
			"type":         "char_move",
			"character_id": char.ID,
			"h3_index":     newIndex,
		},
	}
}

// normalizeBearing keeps bearing in [0, 360)
func normalizeBearing(b float64) float64 {
	b = math.Mod(b, 360)
	if b < 0 {
		b += 360
	}
	return b
}
