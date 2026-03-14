package game

import (
	"math"
	"math/rand"
	"time"

	"hexslayer/internal/models"

	"gorm.io/gorm"
)

// doCombat runs one round: character hits monster, monster hits character.
// Updates HP in DB. Returns combat log events.
func doCombat(db *gorm.DB, char *models.Character, monster *models.MapMonster) []map[string]interface{} {
	var logs []map[string]interface{}

	charStats := combatantFromCharacter(char)
	monStats := combatantFromMonster(monster)

	// Character attacks monster
	hit := attack(charStats, monStats)
	monster.CurrentHP -= hit.Damage
	if monster.CurrentHP <= 0 {
		monster.CurrentHP = 0
		monster.IsAlive = false
		db.Model(monster).Updates(map[string]interface{}{
			"current_hp": 0,
			"is_alive":   false,
		})
	} else {
		db.Model(monster).Update("current_hp", monster.CurrentHP)
	}
	logs = append(logs, map[string]interface{}{
		"type":         "combat_log",
		"attacker":     char.Name,
		"attacker_id":  char.ID,
		"defender":     monster.MonsterType.Name,
		"defender_id":  monster.ID,
		"damage":       hit.Damage,
		"is_crit":      hit.IsCrit,
		"character_id": char.ID,
		"character_hp": char.HP,
		"monster_id":   monster.ID,
		"monster_hp":   monster.CurrentHP,
	})

	// Monster attacks character (only if both still alive)
	if monster.IsAlive && char.IsAlive {
		hit2 := attack(monStats, charStats)
		char.HP -= hit2.Damage
		if char.HP <= 0 {
			char.HP = 0
			char.IsAlive = false
			now := time.Now()
			db.Model(char).Updates(map[string]interface{}{
				"hp":       0,
				"is_alive": false,
				"died_at":  now,
			})
		} else {
			db.Model(char).Update("hp", char.HP)
		}
		logs = append(logs, map[string]interface{}{
			"type":         "combat_log",
			"attacker":     monster.MonsterType.Name,
			"attacker_id":  monster.ID,
			"defender":     char.Name,
			"defender_id":  char.ID,
			"damage":       hit2.Damage,
			"is_crit":      hit2.IsCrit,
			"character_id": char.ID,
			"character_hp": char.HP,
			"monster_id":   monster.ID,
			"monster_hp":   monster.CurrentHP,
		})
	}

	return logs
}

// Combatant is a shared interface for anything with combat stats.
type Combatant struct {
	BaseDamage      int
	DamageAmp       float64
	DamageReduction float64
	CritChance      float64
	CritMultiplier  float64
}

// CombatResult holds the outcome of one attack.
type CombatResult struct {
	Damage int
	IsCrit bool
}

func combatantFromCharacter(c *models.Character) Combatant {
	return Combatant{
		BaseDamage:      c.BaseDamage,
		DamageAmp:       c.DamageAmp,
		DamageReduction: c.DamageReduction,
		CritChance:      c.CritChance,
		CritMultiplier:  c.CritMultiplier,
	}
}

func combatantFromMonster(m *models.MapMonster) Combatant {
	return Combatant{
		BaseDamage:      m.MonsterType.BaseDamage,
		DamageAmp:       m.MonsterType.DamageAmp,
		DamageReduction: m.MonsterType.DamageReduction,
		CritChance:      m.MonsterType.CritChance,
		CritMultiplier:  m.MonsterType.CritMultiplier,
	}
}

// attack calculates damage from attacker to defender.
func attack(attacker Combatant, defender Combatant) CombatResult {
	rawDamage := float64(attacker.BaseDamage) * attacker.DamageAmp
	isCrit := rand.Float64() < attacker.CritChance
	if isCrit {
		rawDamage *= attacker.CritMultiplier
	}
	finalDamage := rawDamage * (1.0 - defender.DamageReduction)
	return CombatResult{
		Damage: int(math.Round(finalDamage)),
		IsCrit: isCrit,
	}
}
