package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetZones godoc
// @Summary Get map zones
// @Description Compute zones from coordinate, ensure monsters spawned, return monsters in zone
// @Tags map
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/map/zones [get]
func GetZones(c *gin.Context) {
	// TODO: compute zones via h3.GridDisk(bangkokCenter, 2) and return with monster counts
	c.JSON(http.StatusOK, gin.H{
		"message": "zones endpoint — not yet implemented",
		"zones":   []interface{}{},
	})
}
