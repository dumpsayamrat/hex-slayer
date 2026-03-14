package game

import (
	"log"
	"sync"
	"time"

	"hexslayer/internal/config"
	"hexslayer/internal/db"
	"hexslayer/internal/models"
	"hexslayer/internal/ws"
)

// Engine manages the game tick loop — one goroutine per active zone.
// A zone goroutine starts when the first client subscribes.
// It keeps running as long as there are alive characters in the zone,
// even if all clients disconnect.
type Engine struct {
	active map[string]chan struct{} // stop channel per running zone goroutine
	mu     sync.Mutex
}

func NewEngine() *Engine {
	return &Engine{
		active: make(map[string]chan struct{}),
	}
}

// Start registers the subscribe hook on the hub.
func (e *Engine) Start() {
	ws.Hub.OnFirstSubscribe = e.onFirstSubscribe
	log.Println("game engine: ready (zone loops start on first subscriber)")
}

// onFirstSubscribe is called by the hub when a topic goes from 0→1 subscribers.
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

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if !e.tickZone(zone) {
				// No alive characters — stop this zone loop
				e.mu.Lock()
				delete(e.active, zone)
				e.mu.Unlock()
				log.Printf("game engine: stopped zone loop %s (no alive characters)", zone)
				return
			}
		}
	}
}

// tickZone processes one tick for a zone. Returns false if the zone
// has no alive characters and should stop.
func (e *Engine) tickZone(zone string) bool {
	// Check if any alive characters exist in this zone
	var charCount int64
	db.DB.Model(&models.Character{}).
		Where("h3_zone = ? AND is_alive = true", zone).
		Count(&charCount)

	if charCount == 0 {
		return false
	}

	topic := "zone:" + zone

	// Mock tick payload
	ws.Hub.Broadcast(topic, map[string]interface{}{
		"type": "tick_update",
		"zone": zone,
		"mock": true,
		"ts":   time.Now().UnixMilli(),
	})

	return true
}

// topicToZone extracts the zone hex from a "zone:<hex>" topic string.
func topicToZone(topic string) string {
	const prefix = "zone:"
	if len(topic) > len(prefix) && topic[:len(prefix)] == prefix {
		return topic[len(prefix):]
	}
	return ""
}
