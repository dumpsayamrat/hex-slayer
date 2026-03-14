package game

import (
	"time"

	"hexslayer/internal/config"
	"hexslayer/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// charTick processes one character through a single game tick.
// It holds references to shared tick state (monster index, engaged set)
// and collects events to be broadcast after all characters are processed.
type charTick struct {
	db     *gorm.DB
	char   *models.Character
	events []map[string]interface{}

	// Shared tick state (read/write across all charTicks in one tickZone call)
	monsterByID      map[string]*models.MapMonster
	monsterPtrs      []*models.MapMonster
	engaged          map[string]bool
	engagementByChar map[string]*models.CharacterEngagement
}

// process runs the state machine for one character.
func (ct *charTick) process() {
	if !ct.char.IsAlive {
		return
	}

	eng := ct.engagementByChar[ct.char.ID]

	// STATE: COMBAT — has active engagement with a living monster
	if eng != nil {
		ct.processCombat(eng)
		return
	}

	// STATE: HUNTING — has a target, walking toward it
	if ct.char.TargetMonsterID != nil {
		if ct.processHunting() {
			return
		}
		// Target was invalid, fall through to scanning
	}

	// STATE: SCANNING — look for nearby monsters
	if ct.processScanning() {
		return
	}

	// STATE: WANDERING — no monsters nearby
	ct.processWandering()
}

func (ct *charTick) processCombat(eng *models.CharacterEngagement) {
	monster := ct.monsterByID[eng.MonsterID]

	// Monster dead or gone — disengage and walk away
	if monster == nil || !monster.IsAlive {
		ct.disengage(ct.char)
		ct.char.TargetMonsterID = nil
		ct.events = append(ct.events, ct.wanderStep()...)
		return
	}

	// Fight
	ct.engaged[monster.ID] = true
	logs := doCombat(ct.db, ct.char, monster)
	ct.events = append(ct.events, logs...)

	ct.handleCombatOutcome(monster)
}

func (ct *charTick) processHunting() bool {
	target := ct.monsterByID[*ct.char.TargetMonsterID]

	// Target gone or stolen by another character
	if target == nil || !target.IsAlive || ct.engaged[target.ID] {
		ct.char.TargetMonsterID = nil
		ct.db.Model(ct.char).Update("target_monster_id", nil)
		return false
	}

	// Move one step toward target
	newIndex, dist := moveToward(ct.char, target)
	if newIndex != ct.char.H3Index {
		ct.char.H3Index = newIndex
		ct.db.Model(ct.char).Updates(map[string]interface{}{
			"h3_index":       newIndex,
			"wander_bearing": ct.char.WanderBearing,
		})
		ct.events = append(ct.events, map[string]interface{}{
			"type":         "char_move",
			"character_id": ct.char.ID,
			"h3_index":     newIndex,
		})
	}

	// Not close enough yet — keep walking
	if dist > config.GridDiskRadius {
		return true
	}

	// Close enough — engage!
	ct.engaged[target.ID] = true
	newEng := models.CharacterEngagement{
		ID:          uuid.New().String(),
		CharacterID: ct.char.ID,
		MonsterID:   target.ID,
		EngagedAt:   time.Now(),
	}
	ct.db.Create(&newEng)
	ct.engagementByChar[ct.char.ID] = &newEng

	ct.events = append(ct.events, map[string]interface{}{
		"type":         "combat_engage",
		"character_id": ct.char.ID,
		"monster_id":   target.ID,
	})

	// First strike
	logs := doCombat(ct.db, ct.char, target)
	ct.events = append(ct.events, logs...)

	ct.handleCombatOutcome(target)
	return true
}

func (ct *charTick) processScanning() bool {
	target := findNearestFreeMonster(ct.char, ct.monsterPtrs, ct.engaged)
	if target == nil {
		return false
	}

	ct.char.TargetMonsterID = &target.ID
	ct.db.Model(ct.char).Update("target_monster_id", target.ID)
	return true
}

func (ct *charTick) processWandering() {
	newIndex := wander(ct.char)
	if newIndex == ct.char.H3Index {
		return
	}
	ct.char.H3Index = newIndex
	ct.db.Model(ct.char).Updates(map[string]interface{}{
		"h3_index":       newIndex,
		"wander_bearing": ct.char.WanderBearing,
	})
	ct.events = append(ct.events, map[string]interface{}{
		"type":         "char_move",
		"character_id": ct.char.ID,
		"h3_index":     newIndex,
	})
}

// handleCombatOutcome checks if the monster or character died after combat
// and emits the appropriate events. Called from both processCombat and processHunting.
func (ct *charTick) handleCombatOutcome(monster *models.MapMonster) {
	if !monster.IsAlive {
		ct.char.Kills++
		ct.db.Model(ct.char).Update("kills", ct.char.Kills)
		ct.events = append(ct.events, map[string]interface{}{
			"type":       "monster_died",
			"monster_id": monster.ID,
			"killed_by":  ct.char.Name,
		})
		ct.disengage(ct.char)
		ct.char.TargetMonsterID = nil
		ct.events = append(ct.events, ct.wanderStep()...)
	}

	if !ct.char.IsAlive {
		ct.events = append(ct.events, map[string]interface{}{
			"type":         "character_died",
			"character_id": ct.char.ID,
			"killed_by":    monster.MonsterType.Name,
		})
		ct.disengage(ct.char)
	}
}

// disengage removes the character's engagement from DB and the local map.
func (ct *charTick) disengage(char *models.Character) {
	ct.db.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
	delete(ct.engagementByChar, char.ID)
}

// wanderStep performs one wander step and returns events.
// Used after kills so the character walks away before scanning for the next fight.
func (ct *charTick) wanderStep() []map[string]interface{} {
	newIndex := wander(ct.char)
	if newIndex == ct.char.H3Index {
		return nil
	}
	ct.char.H3Index = newIndex
	ct.db.Model(ct.char).Updates(map[string]interface{}{
		"h3_index":          newIndex,
		"wander_bearing":    ct.char.WanderBearing,
		"target_monster_id": nil,
	})
	return []map[string]interface{}{
		{
			"type":         "char_move",
			"character_id": ct.char.ID,
			"h3_index":     newIndex,
		},
	}
}
