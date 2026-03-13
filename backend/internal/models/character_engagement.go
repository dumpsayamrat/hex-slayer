package models

import "time"

type CharacterEngagement struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	CharacterID string    `gorm:"uniqueIndex;not null" json:"character_id"`
	Character   Character `gorm:"foreignKey:CharacterID" json:"-"`
	MonsterID   string    `gorm:"not null;index" json:"monster_id"`
	Monster     MapMonster `gorm:"foreignKey:MonsterID" json:"-"`
	EngagedAt   time.Time `json:"engaged_at"`
}
