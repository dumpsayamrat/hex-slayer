package util

import (
	"fmt"

	"hexslayer/internal/config"

	"github.com/ziprecruiter/h3-go/pkg/h3"
)

// ValidateZone checks that the string is a valid H3 cell at ZoneResolution.
func ValidateZone(zone string) error {
	cell, err := h3.NewCellFromString(zone)
	if err != nil {
		return fmt.Errorf("invalid h3_zone: %s", zone)
	}
	if cell.Resolution() != config.ZoneResolution {
		return fmt.Errorf("h3_zone must be resolution %d, got %d", config.ZoneResolution, cell.Resolution())
	}
	return nil
}
