package handlers

import (
	"hexslayer/internal/game"
	"hexslayer/internal/services"
	"hexslayer/internal/ws"

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
