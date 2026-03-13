package db

import (
	"log"

	"hexslayer/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("hexslayer.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// SQLite WAL mode for better concurrency
	sqlDB.Exec("PRAGMA journal_mode=WAL")
	sqlDB.Exec("PRAGMA synchronous=NORMAL")

	migrate()
	seed()
}

func migrate() {
	err := DB.AutoMigrate(
		&models.MonsterType{},
		&models.MapMonster{},
		&models.Player{},
		&models.Character{},
		&models.CharacterEngagement{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("database migrated successfully")
}

func seed() {
	var count int64
	DB.Model(&models.MonsterType{}).Count(&count)
	if count > 0 {
		log.Println("monster_types already seeded, skipping")
		return
	}

	monsterTypes := []models.MonsterType{
		{Name: "Slime", BaseDamage: 15, DamageAmp: 1.0, DamageReduction: 0.05, CritChance: 0.05, CritMultiplier: 1.3, MaxHP: 80, Icon: "slime.png"},
		{Name: "Goblin", BaseDamage: 25, DamageAmp: 1.1, DamageReduction: 0.08, CritChance: 0.10, CritMultiplier: 1.5, MaxHP: 120, Icon: "goblin.png"},
		{Name: "Orc", BaseDamage: 35, DamageAmp: 1.2, DamageReduction: 0.15, CritChance: 0.12, CritMultiplier: 1.6, MaxHP: 200, Icon: "orc.png"},
		{Name: "Troll", BaseDamage: 45, DamageAmp: 1.3, DamageReduction: 0.20, CritChance: 0.15, CritMultiplier: 1.8, MaxHP: 300, Icon: "troll.png"},
		{Name: "Dragon", BaseDamage: 60, DamageAmp: 1.5, DamageReduction: 0.25, CritChance: 0.20, CritMultiplier: 2.0, MaxHP: 500, Icon: "dragon.png"},
	}

	if err := DB.Create(&monsterTypes).Error; err != nil {
		log.Fatalf("failed to seed monster types: %v", err)
	}
	log.Println("seeded 5 monster types")

	// TODO: Seed map_monsters — compute H3 zones via GridDisk and spawn monsters
	// This requires h3-go and will be implemented in the game engine
	log.Println("TODO: seed initial map monsters when game engine is implemented")
}
