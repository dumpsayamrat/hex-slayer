package game

import (
	"math"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	h3 "github.com/uber/h3-go/v4"
)

// findNearestFreeMonster scans within DetectionRadius for the closest
// alive, unengaged monster. Returns nil if none found.
func findNearestFreeMonster(char *models.Character, monsters []*models.MapMonster, engaged map[string]bool) *models.MapMonster {
	charCell := h3.CellFromString(char.H3Index)
	if !h3.IsValidIndex(charCell) {
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
		ring, err := h3.GridRing(charCell, k)
		if err != nil {
			continue
		}
		for _, cell := range ring {
			if ms, ok := byCell[cell.String()]; ok && len(ms) > 0 {
				return ms[0]
			}
		}
	}

	return nil
}

// moveToward moves char one step toward the target monster's position.
// Returns the new H3 index and the remaining grid distance.
func moveToward(char *models.Character, target *models.MapMonster) (string, int) {
	charCell := h3.CellFromString(char.H3Index)
	targetCell := h3.CellFromString(target.H3Index)

	charLL, err := h3.CellToLatLng(charCell)
	if err != nil {
		return char.H3Index, 999
	}
	targetLL, err := h3.CellToLatLng(targetCell)
	if err != nil {
		return char.H3Index, 999
	}

	// Compute bearing toward target
	dLng := (targetLL.Lng - charLL.Lng) * math.Pi / 180.0
	charLatRad := charLL.Lat * math.Pi / 180.0
	targetLatRad := targetLL.Lat * math.Pi / 180.0

	y := math.Sin(dLng) * math.Cos(targetLatRad)
	x := math.Cos(charLatRad)*math.Sin(targetLatRad) - math.Sin(charLatRad)*math.Cos(targetLatRad)*math.Cos(dLng)
	bearing := math.Atan2(y, x) * 180.0 / math.Pi

	// Update character bearing to face the target
	char.WanderBearing = normalizeBearing(bearing)

	// Move one step toward target
	bearingRad := bearing * math.Pi / 180.0
	newLat := charLL.Lat + stepDistanceDeg*math.Cos(bearingRad)
	newLng := charLL.Lng + stepDistanceDeg*math.Sin(bearingRad)/math.Cos(charLL.Lat*math.Pi/180.0)

	newCell, err := h3.LatLngToCell(h3.NewLatLng(newLat, newLng), config.EntityResolution)
	if err != nil {
		return char.H3Index, 999
	}

	// Check still in zone
	parent, err := newCell.Parent(config.ZoneResolution)
	if err != nil || parent.String() != char.H3Zone {
		return char.H3Index, 999
	}

	// Calculate remaining distance to target
	dist, err := h3.GridDistance(newCell, targetCell)
	if err != nil {
		dist = 999
	}

	return newCell.String(), dist
}
