package game

import (
	"math"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	h3light "github.com/ThingsIXFoundation/h3-light"
	"github.com/ziprecruiter/h3-go/pkg/h3"
)

// findNearestFreeMonster scans within DetectionRadius for the closest
// alive, unengaged monster. Returns nil if none found.
func findNearestFreeMonster(char *models.Character, monsters []*models.MapMonster, engaged map[string]bool) *models.MapMonster {
	charCell, err := h3.NewCellFromString(char.H3Index)
	if err != nil {
		return nil
	}

	// Build monster lookup by h3_index (only alive + not engaged)
	byCell := make(map[string][]*models.MapMonster)
	for _, m := range monsters {
		if !m.IsAlive || engaged[m.ID] {
			continue
		}
		byCell[m.H3Index] = append(byCell[m.H3Index], m)
	}

	// Scan outward ring by ring to find the nearest monster
	for k := 1; k <= config.DetectionRadius; k++ {
		cells, err := charCell.GridDisk(k)
		if err != nil {
			continue
		}
		// Check cells at exactly distance k (outer ring)
		for _, cell := range cells {
			dist, err := charCell.GridDistance(cell)
			if err != nil || dist != k {
				continue
			}
			if ms, ok := byCell[cell.String()]; ok && len(ms) > 0 {
				return ms[0]
			}
		}
	}

	return nil
}

// moveToward moves char one step toward the target monster's position.
// Uses h3-light for lat/lng, then converts back to H3.
// Returns the new H3 index and the remaining grid distance.
func moveToward(char *models.Character, target *models.MapMonster) (string, int) {
	charLight := h3light.MustCellFromString(char.H3Index)
	targetLight := h3light.MustCellFromString(target.H3Index)

	charLat, charLng := charLight.LatLon()
	targetLat, targetLng := targetLight.LatLon()

	// Compute bearing toward target
	dLng := (targetLng - charLng) * math.Pi / 180.0
	charLatRad := charLat * math.Pi / 180.0
	targetLatRad := targetLat * math.Pi / 180.0

	y := math.Sin(dLng) * math.Cos(targetLatRad)
	x := math.Cos(charLatRad)*math.Sin(targetLatRad) - math.Sin(charLatRad)*math.Cos(targetLatRad)*math.Cos(dLng)
	bearing := math.Atan2(y, x) * 180.0 / math.Pi

	// Update character bearing to face the target
	char.WanderBearing = normalizeBearing(bearing)

	// Move one step toward target
	bearingRad := bearing * math.Pi / 180.0
	newLat := charLat + stepDistanceDeg*math.Cos(bearingRad)
	newLng := charLng + stepDistanceDeg*math.Sin(bearingRad)/math.Cos(charLat*math.Pi/180.0)

	ll := h3.NewLatLng(newLat, newLng)
	newCell, err := h3.NewCellFromLatLng(ll, config.EntityResolution)
	if err != nil {
		return char.H3Index, 999
	}

	// Check still in zone
	parent, err := newCell.Parent(config.ZoneResolution)
	if err != nil || parent.String() != char.H3Zone {
		return char.H3Index, 999
	}

	// Calculate remaining distance to target
	targetCell, err := h3.NewCellFromString(target.H3Index)
	if err != nil {
		return newCell.String(), 999
	}
	dist, err := newCell.GridDistance(targetCell)
	if err != nil {
		dist = 999
	}

	return newCell.String(), dist
}
