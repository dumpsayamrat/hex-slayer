package services

import (
	"hexslayer/internal/db"
	"hexslayer/internal/models"

	"github.com/google/uuid"
)

func CreatePlayer() (*models.Player, error) {
	player := models.Player{
		ID:           uuid.New().String(),
		SessionToken: uuid.New().String(),
	}

	if err := db.DB.Create(&player).Error; err != nil {
		return nil, err
	}

	return &player, nil
}
