package handlers

import (
	"net/http"

	"hexslayer/internal/game"
	"hexslayer/internal/middleware"
	"hexslayer/internal/services"
	"hexslayer/internal/ws"

	"github.com/gin-gonic/gin"
)

type DeployCharacterRequest struct {
	H3Zone string `json:"h3_zone" binding:"required"`
}

type DeployCharacterResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	H3Zone          string  `json:"h3_zone"`
	H3Index         string  `json:"h3_index"`
	HP              int     `json:"hp"`
	MaxHP           int     `json:"max_hp"`
	BaseDamage      int     `json:"base_damage"`
	DamageAmp       float64 `json:"damage_amp"`
	DamageReduction float64 `json:"damage_reduction"`
	CritChance      float64 `json:"crit_chance"`
	CritMultiplier  float64 `json:"crit_multiplier"`
}

// DeployCharacter godoc
// @Summary Deploy a character to a zone
// @Description Deploy a new character with randomized stats to the specified H3 zone. Max 2 alive characters per player.
// @Tags character
// @Accept json
// @Produce json
// @Param body body DeployCharacterRequest true "Deploy request"
// @Security BearerAuth
// @Success 201 {object} DeployCharacterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/character/deploy [post]
func DeployCharacter(c *gin.Context) {
	var req DeployCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := middleware.GetPlayer(c)
	if player == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	char, err := services.DeployCharacter(player.ID, req.H3Zone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the zone tick loop is running
	game.GameEngine.EnsureZoneLoop(char.H3Zone)

	// Broadcast to all zone subscribers
	ws.Hub.Broadcast("zone:"+char.H3Zone, map[string]interface{}{
		"type":      "char_deployed",
		"id":        char.ID,
		"name":      char.Name,
		"player_id": char.PlayerID,
		"h3_zone":   char.H3Zone,
		"h3_index":  char.H3Index,
		"hp":        char.HP,
		"max_hp":    char.MaxHP,
	})

	c.JSON(http.StatusCreated, DeployCharacterResponse{
		ID:              char.ID,
		Name:            char.Name,
		H3Zone:          char.H3Zone,
		H3Index:         char.H3Index,
		HP:              char.HP,
		MaxHP:           char.MaxHP,
		BaseDamage:      char.BaseDamage,
		DamageAmp:       char.DamageAmp,
		DamageReduction: char.DamageReduction,
		CritChance:      char.CritChance,
		CritMultiplier:  char.CritMultiplier,
	})
}
