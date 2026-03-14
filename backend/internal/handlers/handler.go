package handlers

import (
	"errors"
	"log"
	"net/http"

	"hexslayer/internal/apperr"
	"hexslayer/internal/game"
	"hexslayer/internal/services"
	"hexslayer/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB         *gorm.DB
	Hub        *ws.Hub
	Engine     *game.Engine
	Players    *services.PlayerService
	Zones      *services.ZoneService
	Characters *services.CharacterService
}

func New(db *gorm.DB, hub *ws.Hub, engine *game.Engine, ps *services.PlayerService, zs *services.ZoneService, cs *services.CharacterService) *Handler {
	return &Handler{
		DB:         db,
		Hub:        hub,
		Engine:     engine,
		Players:    ps,
		Zones:      zs,
		Characters: cs,
	}
}

// respondError inspects the error type and writes the appropriate HTTP response.
// Business errors (Validation, NotFound) return their message to the client.
// System errors return a generic message and log the real error server-side.
func (h *Handler) respondError(c *gin.Context, err error) {
	var ve *apperr.Validation
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
		return
	}

	var nf *apperr.NotFound
	if errors.As(err, &nf) {
		c.JSON(http.StatusNotFound, gin.H{"error": nf.Message})
		return
	}

	// System error — log details, return generic message
	log.Printf("internal error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
