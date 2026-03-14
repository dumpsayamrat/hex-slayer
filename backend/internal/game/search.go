package game

import (
	"math/rand"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	"github.com/ziprecruiter/h3-go/pkg/h3"
)

// findRandomFreeMonster does a GridDisk search from the character's position
// and picks a random alive monster within range that isn't already engaged.
func findRandomFreeMonster(char *models.Character, monsters []*models.MapMonster, engaged map[string]bool) *models.MapMonster {
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

	// GridDisk at k=KRingSearchMax
	cells, err := charCell.GridDisk(config.KRingSearchMax)
	if err != nil {
		return nil
	}

	// Collect all candidates within range
	var candidates []*models.MapMonster
	for _, cell := range cells {
		if ms, ok := byCell[cell.String()]; ok {
			candidates = append(candidates, ms...)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	return candidates[rand.Intn(len(candidates))]
}
