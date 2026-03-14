package dto

// ZoneMonsterResponse is the lean monster data sent to the frontend.
type ZoneMonsterResponse struct {
	ID        string `json:"id"`
	H3Index   string `json:"h3_index"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	CurrentHP int    `json:"current_hp"`
	MaxHP     int    `json:"max_hp"`
	IsAlive   bool   `json:"is_alive"`
}
