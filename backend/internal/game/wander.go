package game

import (
	"math/rand"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	"github.com/ziprecruiter/h3-go/pkg/h3"
)

// wander moves a character to a random res-12 cell ~5–15 cells away,
// staying within the same zone.
func wander(char *models.Character) string {
	charCell, err := h3.NewCellFromString(char.H3Index)
	if err != nil {
		return char.H3Index
	}

	// Pick random k between 5 and 15
	k := 5 + rand.Intn(11)

	cells, err := charCell.GridDisk(k)
	if err != nil || len(cells) == 0 {
		return char.H3Index
	}

	// Filter to cells that are still within the same zone
	var valid []h3.Cell
	for _, c := range cells {
		parent, err := c.Parent(config.ZoneResolution)
		if err != nil {
			continue
		}
		if parent.String() == char.H3Zone {
			valid = append(valid, c)
		}
	}

	if len(valid) == 0 {
		return char.H3Index
	}

	return valid[rand.Intn(len(valid))].String()
}
