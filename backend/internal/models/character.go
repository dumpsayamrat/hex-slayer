package models

import "time"

type Character struct {
	ID              string     `gorm:"primaryKey;type:text" json:"id"`
	PlayerID        string     `gorm:"not null;index" json:"player_id"`
	Player          Player     `gorm:"foreignKey:PlayerID" json:"-"`
	Name            string     `gorm:"not null" json:"name"`
	H3Zone          string     `gorm:"not null;index" json:"h3_zone"`
	H3Index         string     `gorm:"not null" json:"h3_index"`
	HP              int        `gorm:"not null" json:"hp"`
	MaxHP           int        `gorm:"not null" json:"max_hp"`
	BaseDamage      int        `gorm:"not null" json:"base_damage"`
	DamageAmp       float64    `gorm:"not null" json:"damage_amp"`
	DamageReduction float64    `gorm:"not null" json:"damage_reduction"`
	CritChance      float64    `gorm:"not null" json:"crit_chance"`
	CritMultiplier  float64    `gorm:"not null" json:"crit_multiplier"`
	Kills           int        `gorm:"not null;default:0" json:"kills"`
	IsAlive         bool       `gorm:"not null;default:true" json:"is_alive"`
	WanderBearing   float64    `gorm:"not null;default:0" json:"wander_bearing"`
	TargetMonsterID *string    `gorm:"index" json:"target_monster_id"`
	DeployedAt      time.Time  `json:"deployed_at"`
	DiedAt          *time.Time `json:"died_at"`
}
