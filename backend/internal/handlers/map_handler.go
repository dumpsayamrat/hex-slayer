package handlers

import (
	"net/http"
	"strconv"

	"hexslayer/internal/services"

	"github.com/gin-gonic/gin"
)

// ZoneResponse is the response for GET /api/map/zones.
type ZoneResponse struct {
	H3Zone   string                        `json:"h3_zone"`
	Monsters []services.ZoneMonsterResponse `json:"monsters"`
}

// GetZones godoc
// @Summary Get map zones
// @Description Compute zone from coordinate, ensure monsters spawned, return all monsters in zone
// @Tags map
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Security BearerAuth
// @Success 200 {object} ZoneResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/map/zones [get]
func GetZones(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng query params required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat value"})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng value"})
		return
	}

	zoneStr, monsters, err := services.GetOrCreateZoneMonsters(lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load zone"})
		return
	}

	c.JSON(http.StatusOK, ZoneResponse{
		H3Zone:   zoneStr,
		Monsters: monsters,
	})
}
