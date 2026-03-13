package models

import "time"

type MonsterType struct {
	ID              uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string  `gorm:"not null" json:"name"`
	BaseDamage      int     `gorm:"not null" json:"base_damage"`
	DamageAmp       float64 `gorm:"not null" json:"damage_amp"`
	DamageReduction float64 `gorm:"not null" json:"damage_reduction"`
	CritChance      float64 `gorm:"not null" json:"crit_chance"`
	CritMultiplier  float64 `gorm:"not null" json:"crit_multiplier"`
	MaxHP           int     `gorm:"not null" json:"max_hp"`
	Icon            string  `gorm:"not null" json:"icon"`
	CreatedAt       time.Time `json:"created_at"`
}
