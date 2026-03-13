package models

import "time"

type MapMonster struct {
	ID            string     `gorm:"primaryKey;type:text" json:"id"`
	H3Zone        string     `gorm:"not null;index" json:"h3_zone"`
	H3Index       string     `gorm:"not null" json:"h3_index"`
	MonsterTypeID uint       `gorm:"not null" json:"monster_type_id"`
	MonsterType   MonsterType `gorm:"foreignKey:MonsterTypeID" json:"monster_type"`
	CurrentHP     int        `gorm:"not null" json:"current_hp"`
	IsAlive       bool       `gorm:"not null;default:true" json:"is_alive"`
	RespawnAt     *time.Time `json:"respawn_at"`
	CreatedAt     time.Time  `json:"created_at"`
}
