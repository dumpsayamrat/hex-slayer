package game

import (
	"math"
	"math/rand"

	"hexslayer/internal/models"
)

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
