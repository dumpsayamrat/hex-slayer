package util

import (
	"fmt"

	"hexslayer/internal/config"

	h3 "github.com/uber/h3-go/v4"
)

// ValidateZone checks that the string is a valid H3 cell at ZoneResolution.
func ValidateZone(zone string) error {
	cell := h3.CellFromString(zone)
	if !h3.IsValidIndex(cell) {
		return fmt.Errorf("invalid h3_zone: %s", zone)
	}
	if cell.Resolution() != config.ZoneResolution {
		return fmt.Errorf("h3_zone must be resolution %d, got %d", config.ZoneResolution, cell.Resolution())
	}
	return nil
}
