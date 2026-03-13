package game

import "log"

// Engine manages the game tick loop — one goroutine per zone.
// TODO: implement Start(), runZoneLoop(), tickZone()
type Engine struct {
	// zones will hold the list of H3 res-6 zone indexes
	Zones []string
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Start() {
	log.Println("game engine initialized (stub — no tick loops running)")
	// TODO: compute zones via h3.GridDisk, launch goroutine per zone
}
