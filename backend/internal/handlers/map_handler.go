package handlers

import (
	"net/http"
	"strconv"

	"hexslayer/internal/db"
	"hexslayer/internal/models"
	"hexslayer/internal/services"

	"github.com/gin-gonic/gin"
)

type ZoneCharacterResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	PlayerID string `json:"player_id"`
	H3Index  string `json:"h3_index"`
}

// ZoneResponse is the response for GET /api/map/zones.
type ZoneResponse struct {
	H3Zone     string                        `json:"h3_zone"`
	Monsters   []services.ZoneMonsterResponse `json:"monsters"`
	Characters []ZoneCharacterResponse        `json:"characters"`
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

	// Load alive characters in this zone
	var chars []models.Character
	db.DB.Where("h3_zone = ? AND is_alive = true", zoneStr).Find(&chars)
	charData := make([]ZoneCharacterResponse, len(chars))
	for i, c := range chars {
		charData[i] = ZoneCharacterResponse{
			ID:       c.ID,
			Name:     c.Name,
			HP:       c.HP,
			MaxHP:    c.MaxHP,
			PlayerID: c.PlayerID,
			H3Index:  c.H3Index,
		}
	}

	c.JSON(http.StatusOK, ZoneResponse{
		H3Zone:     zoneStr,
		Monsters:   monsters,
		Characters: charData,
	})
}
