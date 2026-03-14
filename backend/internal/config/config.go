package config

const (
	ZoneMonsterCap      = 300
	MaxCharactersAlive  = 2
	TickIntervalSeconds = 2
	KRingSearchMax      = 5
	BangkokLat          = 13.7563
	BangkokLng          = 100.5018
	ZoneResolution      = 6
	EntityResolution    = 12
	GridDiskRadius      = 2
	ZoneMaxDurationMins = 30 // hard stop per zone loop for demo

	// Wander settings
	WanderBearingDriftMax = 30.0 // max degrees drift per tick
	WanderBigTurnChance   = 0.05 // 5% chance of a big turn per tick
	WanderBigTurnMax      = 90.0 // max degrees for a big turn

	// Hunting settings
	DetectionRadius = 25 // H3 cells radius to scan for monsters
)
