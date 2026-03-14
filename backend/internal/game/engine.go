package game

import (
	"log"
	"sync"
	"time"

	"hexslayer/internal/config"
	"hexslayer/internal/models"
	"hexslayer/internal/ws"

	"gorm.io/gorm"
)

// Engine manages the game tick loop — one goroutine per active zone.
type Engine struct {
	db     *gorm.DB
	hub    *ws.Hub
	active map[string]chan struct{}
	mu     sync.Mutex
}

func NewEngine(db *gorm.DB, hub *ws.Hub) *Engine {
	return &Engine{
		db:     db,
		hub:    hub,
		active: make(map[string]chan struct{}),
	}
}

func (e *Engine) Start() {
	e.hub.OnSubscribe = e.onSubscribe
	log.Println("game engine: ready (zone loops start on subscriber)")
}

func (e *Engine) onSubscribe(topic string) {
	zone := topicToZone(topic)
	if zone == "" {
		return
	}
	e.EnsureZoneLoop(zone)
}

// EnsureZoneLoop starts a zone tick loop if one isn't already running.
func (e *Engine) EnsureZoneLoop(zone string) {
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
	e.db.Where("h3_zone = ? AND is_alive = true", zone).Find(&characters)
	if len(characters) == 0 {
		return false
	}

	// Load alive monsters in zone with their type
	var monsters []models.MapMonster
	e.db.Preload("MonsterType").Where("h3_zone = ? AND is_alive = true", zone).Find(&monsters)

	// Build monster lookups
	monsterByID := make(map[string]*models.MapMonster, len(monsters))
	monsterPtrs := make([]*models.MapMonster, len(monsters))
	for i := range monsters {
		monsterByID[monsters[i].ID] = &monsters[i]
		monsterPtrs[i] = &monsters[i]
	}

	// Load existing engagements
	charIDs := make([]string, len(characters))
	for i, c := range characters {
		charIDs[i] = c.ID
	}
	var engagements []models.CharacterEngagement
	e.db.Where("character_id IN ?", charIDs).Find(&engagements)

	engagementByChar := make(map[string]*models.CharacterEngagement, len(engagements))
	for i := range engagements {
		engagementByChar[engagements[i].CharacterID] = &engagements[i]
	}

	// Shared state for this tick
	engaged := make(map[string]bool)

	// Process each character
	var allEvents []map[string]interface{}
	for i := range characters {
		ct := &charTick{
			db:               e.db,
			char:             &characters[i],
			monsterByID:      monsterByID,
			monsterPtrs:      monsterPtrs,
			engaged:          engaged,
			engagementByChar: engagementByChar,
		}
		ct.process()
		allEvents = append(allEvents, ct.events...)
	}

	// Broadcast all events
	for _, evt := range allEvents {
		e.hub.Broadcast(topic, evt)
	}

	return true
}

func topicToZone(topic string) string {
	const prefix = "zone:"
	if len(topic) > len(prefix) && topic[:len(prefix)] == prefix {
		return topic[len(prefix):]
	}
	return ""
}
