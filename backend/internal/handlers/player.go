package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type InitPlayerResponse struct {
	PlayerID     string `json:"playerId"`
	SessionToken string `json:"sessionToken"`
	Name         string `json:"name"`
}

// InitPlayer godoc
// @Summary Create a new player
// @Description Creates a new player with a session token. No request body needed.
// @Tags player
// @Produce json
// @Success 201 {object} InitPlayerResponse
// @Failure 500 {object} map[string]string
// @Router /api/player/init [post]
func (h *Handler) InitPlayer(c *gin.Context) {
	player, err := h.Players.Create()
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, InitPlayerResponse{
		PlayerID:     player.ID,
		SessionToken: player.SessionToken,
		Name:         player.Name,
	})
}
