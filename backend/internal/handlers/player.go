package handlers

import (
	"net/http"

	"hexslayer/internal/services"

	"github.com/gin-gonic/gin"
)

type InitPlayerResponse struct {
	PlayerID     string `json:"playerId"`
	SessionToken string `json:"sessionToken"`
	Name         string `json:"name"`
}

func InitPlayer(c *gin.Context) {
	player, err := services.CreatePlayer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create player"})
		return
	}

	c.JSON(http.StatusCreated, InitPlayerResponse{
		PlayerID:     player.ID,
		SessionToken: player.SessionToken,
		Name:         player.Name,
	})
}
