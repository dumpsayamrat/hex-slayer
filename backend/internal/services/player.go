package services

import (
	"hexslayer/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerService struct {
	db *gorm.DB
}

func NewPlayerService(db *gorm.DB) *PlayerService {
	return &PlayerService{db: db}
}

func (s *PlayerService) Create() (*models.Player, error) {
	player := models.Player{
		ID:           uuid.New().String(),
		SessionToken: uuid.New().String(),
	}

	if err := s.db.Create(&player).Error; err != nil {
		return nil, err
	}

	return &player, nil
}
