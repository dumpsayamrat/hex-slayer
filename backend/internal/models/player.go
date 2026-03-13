package models

import "time"

type Player struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	SessionToken string    `gorm:"uniqueIndex;not null" json:"-"`
	Name         string    `gorm:"not null;default:Adventurer" json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}
