package game

import (
	"log"
	"sync"
	"time"

	"hexslayer/internal/config"
	"hexslayer/internal/db"
	"hexslayer/internal/models"
	"hexslayer/internal/ws"

	"github.com/google/uuid"
)

// Engine manages the game tick loop — one goroutine per active zone.
type Engine struct {
	active map[string]chan struct{}
	mu     sync.Mutex
}

func NewEngine() *Engine {
	return &Engine{
		active: make(map[string]chan struct{}),
	}
}

func (e *Engine) Start() {
	ws.Hub.OnFirstSubscribe = e.onFirstSubscribe
	log.Println("game engine: ready (zone loops start on first subscriber)")
}

func (e *Engine) onFirstSubscribe(topic string) {
	zone := topicToZone(topic)
	if zone == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, running := e.active[zone]; running {
		return
	}

	stop := make(chan struct{})
	e.active[zone] = stop
	go e.runZoneLoop(zone, stop)
	log.Printf("game engine: started zone loop %s", zone)
}

func (e *Engine) runZoneLoop(zone string, stop chan struct{}) {
	ticker := time.NewTicker(config.TickIntervalSeconds * time.Second)
	defer ticker.Stop()

	maxTicks := (config.ZoneMaxDurationMins * 60) / config.TickIntervalSeconds
	tickCount := 0

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			tickCount++
			if tickCount > maxTicks {
				e.mu.Lock()
				delete(e.active, zone)
				e.mu.Unlock()
				log.Printf("game engine: stopped zone loop %s (hit %d min limit)", zone, config.ZoneMaxDurationMins)
				return
			}
			if !e.tickZone(zone) {
				e.mu.Lock()
				delete(e.active, zone)
				e.mu.Unlock()
				log.Printf("game engine: stopped zone loop %s (no alive characters)", zone)
				return
			}
		}
	}
}

// tickZone processes one tick. Returns false if no alive characters.
func (e *Engine) tickZone(zone string) bool {
	topic := "zone:" + zone

	// Load alive characters in zone
	var characters []models.Character
	db.DB.Where("h3_zone = ? AND is_alive = true", zone).Find(&characters)

	if len(characters) == 0 {
		return false
	}

	// Load alive monsters in zone with their type
	var monsters []models.MapMonster
	db.DB.Preload("MonsterType").Where("h3_zone = ? AND is_alive = true", zone).Find(&monsters)

	// Build monster lookup by ID
	monsterByID := make(map[string]*models.MapMonster, len(monsters))
	monsterPtrs := make([]*models.MapMonster, len(monsters))
	for i := range monsters {
		monsterByID[monsters[i].ID] = &monsters[i]
		monsterPtrs[i] = &monsters[i]
	}

	// Load existing engagements for characters in this zone
	charIDs := make([]string, len(characters))
	for i, c := range characters {
		charIDs[i] = c.ID
	}
	var engagements []models.CharacterEngagement
	db.DB.Where("character_id IN ?", charIDs).Find(&engagements)

	engagementByChar := make(map[string]*models.CharacterEngagement, len(engagements))
	for i := range engagements {
		engagementByChar[engagements[i].CharacterID] = &engagements[i]
	}

	// Track which monsters are engaged this tick (prevents double-targeting)
	engaged := make(map[string]bool)

	// Collect events to broadcast
	var events []map[string]interface{}

	for i := range characters {
		char := &characters[i]
		if !char.IsAlive {
			continue
		}

		eng := engagementByChar[char.ID]

		// === STATE: COMBAT (has active engagement) ===
		if eng != nil {
			monster := monsterByID[eng.MonsterID]
			if monster != nil && monster.IsAlive {
				engaged[monster.ID] = true
				logs := doCombat(char, monster)
				events = append(events, logs...)

				if !monster.IsAlive {
					char.Kills++
					db.DB.Model(char).Update("kills", char.Kills)
					events = append(events, map[string]interface{}{
						"type":       "monster_died",
						"monster_id": monster.ID,
						"killed_by":  char.Name,
					})
					db.DB.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
					delete(engagementByChar, char.ID)
					char.TargetMonsterID = nil
					// Walk away after kill before next fight
					events = append(events, wanderAndEmit(char)...)
				}
				if !char.IsAlive {
					events = append(events, map[string]interface{}{
						"type":         "character_died",
						"character_id": char.ID,
						"killed_by":    monster.MonsterType.Name,
					})
					db.DB.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
					delete(engagementByChar, char.ID)
				}
				continue
			}

			// Monster dead or gone — walk away
			db.DB.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
			delete(engagementByChar, char.ID)
			char.TargetMonsterID = nil
			events = append(events, wanderAndEmit(char)...)
		}

		// === STATE: HUNTING (has target, walking toward it) ===
		if char.TargetMonsterID != nil {
			target := monsterByID[*char.TargetMonsterID]
			if target == nil || !target.IsAlive || engaged[target.ID] {
				// Target gone — clear and fall through to scan/wander
				char.TargetMonsterID = nil
				db.DB.Model(char).Update("target_monster_id", nil)
			} else {
				// Move one step toward target
				newIndex, dist := moveToward(char, target)
				if newIndex != char.H3Index {
					char.H3Index = newIndex
					db.DB.Model(char).Updates(map[string]interface{}{
						"h3_index":       newIndex,
						"wander_bearing": char.WanderBearing,
					})
					events = append(events, map[string]interface{}{
						"type":         "char_move",
						"character_id": char.ID,
						"h3_index":     newIndex,
					})
				}

				// Close enough to engage (within 2 cells)
				if dist <= config.GridDiskRadius {
					engaged[target.ID] = true
					newEng := models.CharacterEngagement{
						ID:          uuid.New().String(),
						CharacterID: char.ID,
						MonsterID:   target.ID,
						EngagedAt:   time.Now(),
					}
					db.DB.Create(&newEng)
					engagementByChar[char.ID] = &newEng

					events = append(events, map[string]interface{}{
						"type":         "combat_engage",
						"character_id": char.ID,
						"monster_id":   target.ID,
					})

					logs := doCombat(char, target)
					events = append(events, logs...)

					if !target.IsAlive {
						char.Kills++
						db.DB.Model(char).Update("kills", char.Kills)
						events = append(events, map[string]interface{}{
							"type":       "monster_died",
							"monster_id": target.ID,
							"killed_by":  char.Name,
						})
						db.DB.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
						delete(engagementByChar, char.ID)
						char.TargetMonsterID = nil
						// Walk away after kill before next fight
						events = append(events, wanderAndEmit(char)...)
					}
					if !char.IsAlive {
						events = append(events, map[string]interface{}{
							"type":         "character_died",
							"character_id": char.ID,
							"killed_by":    target.MonsterType.Name,
						})
						db.DB.Delete(&models.CharacterEngagement{}, "character_id = ?", char.ID)
						delete(engagementByChar, char.ID)
					}
				}
				continue
			}
		}

		// === STATE: SCANNING (look for nearby monsters) ===
		target := findNearestFreeMonster(char, monsterPtrs, engaged)
		if target != nil {
			// Spotted a monster — start hunting
			char.TargetMonsterID = &target.ID
			db.DB.Model(char).Update("target_monster_id", target.ID)
			// Don't move yet, will start walking next tick
			continue
		}

		// === STATE: WANDERING (no monsters nearby) ===
		newIndex := wander(char)
		if newIndex != char.H3Index {
			char.H3Index = newIndex
			db.DB.Model(char).Updates(map[string]interface{}{
				"h3_index":       newIndex,
				"wander_bearing": char.WanderBearing,
			})
			events = append(events, map[string]interface{}{
				"type":         "char_move",
				"character_id": char.ID,
				"h3_index":     newIndex,
			})
		}
	}

	// Broadcast all events
	for _, evt := range events {
		ws.Hub.Broadcast(topic, evt)
	}

	return true
}

// doCombat runs one round: character hits monster, monster hits character.
// Updates HP in DB. Returns combat log events.
func doCombat(char *models.Character, monster *models.MapMonster) []map[string]interface{} {
	var logs []map[string]interface{}

	charStats := combatantFromCharacter(char)
	monStats := combatantFromMonster(monster)

	// Character attacks monster
	hit := attack(charStats, monStats)
	monster.CurrentHP -= hit.Damage
	if monster.CurrentHP <= 0 {
		monster.CurrentHP = 0
		monster.IsAlive = false
		db.DB.Model(monster).Updates(map[string]interface{}{
			"current_hp": 0,
			"is_alive":   false,
		})
	} else {
		db.DB.Model(monster).Update("current_hp", monster.CurrentHP)
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
			db.DB.Model(char).Updates(map[string]interface{}{
				"hp":       0,
				"is_alive": false,
				"died_at":  now,
			})
		} else {
			db.DB.Model(char).Update("hp", char.HP)
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

func topicToZone(topic string) string {
	const prefix = "zone:"
	if len(topic) > len(prefix) && topic[:len(prefix)] == prefix {
		return topic[len(prefix):]
	}
	return ""
}
