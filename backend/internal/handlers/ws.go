package handlers

import (
	"log"
	"net/http"

	"hexslayer/internal/models"
	"hexslayer/internal/util"
	"hexslayer/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: restrict to allowed origins in production
		return true
	},
}

func (h *Handler) WebSocketHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Validate session token against DB
	var player models.Player
	if err := h.DB.Where("session_token = ?", token).First(&player).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session token"})
		return
	}

	rawConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	conn := ws.NewConn(rawConn)
	defer func() {
		h.Hub.UnsubscribeAll(conn)
		conn.Close()
	}()

	log.Printf("websocket connected: player=%s", player.ID)

	conn.SendJSON(gin.H{
		"type":    "connected",
		"message": "welcome to hexslayer",
	})

	for {
		var msg map[string]interface{}
		err := conn.Raw.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		msgType, _ := msg["type"].(string)
		log.Printf("ws message: player=%s type=%s", player.ID, msgType)

		switch msgType {
		case "ping":
			conn.SendJSON(gin.H{"type": "pong"})

		case "subscribe_zone":
			zone, _ := msg["h3_zone"].(string)
			if zone == "" {
				conn.SendJSON(gin.H{"type": "error", "message": "h3_zone required"})
				continue
			}
			if err := util.ValidateZone(zone); err != nil {
				conn.SendJSON(gin.H{"type": "error", "message": err.Error()})
				continue
			}
			h.Hub.Subscribe("zone:"+zone, conn)
			conn.SendJSON(gin.H{"type": "subscribed", "h3_zone": zone})
			h.sendZoneSnapshot(conn, zone)

		case "unsubscribe_zone":
			zone, _ := msg["h3_zone"].(string)
			if zone == "" {
				conn.SendJSON(gin.H{"type": "error", "message": "h3_zone required"})
				continue
			}
			h.Hub.Unsubscribe("zone:"+zone, conn)
			conn.SendJSON(gin.H{"type": "unsubscribed", "h3_zone": zone})

		default:
			conn.SendJSON(gin.H{"type": "error", "message": "unknown message type"})
		}
	}
}

// sendZoneSnapshot sends characters and engaged monsters to a single connection.
// FE already has full monster list from GET /api/map/zones — snapshot only sends
// monsters currently in combat so FE can sync their HP.
func (h *Handler) sendZoneSnapshot(conn *ws.Conn, zone string) {
	// Load alive characters
	var characters []models.Character
	h.DB.Where("h3_zone = ? AND is_alive = true", zone).Find(&characters)

	// Load engagements for these characters
	charIDs := make([]string, len(characters))
	for i, c := range characters {
		charIDs[i] = c.ID
	}
	engagementByChar := make(map[string]string)
	var engagedMonsterIDs []string
	if len(charIDs) > 0 {
		var engagements []models.CharacterEngagement
		h.DB.Where("character_id IN ?", charIDs).Find(&engagements)
		for _, e := range engagements {
			engagementByChar[e.CharacterID] = e.MonsterID
			engagedMonsterIDs = append(engagedMonsterIDs, e.MonsterID)
		}
	}

	charData := make([]gin.H, len(characters))
	for i, c := range characters {
		entry := gin.H{
			"id":        c.ID,
			"name":      c.Name,
			"hp":        c.HP,
			"max_hp":    c.MaxHP,
			"player_id": c.PlayerID,
			"h3_index":  c.H3Index,
		}
		if monsterID, ok := engagementByChar[c.ID]; ok {
			entry["fighting_monster_id"] = monsterID
		}
		charData[i] = entry
	}

	// Load only engaged monsters (to sync current HP)
	var monsterData []gin.H
	if len(engagedMonsterIDs) > 0 {
		var monsters []models.MapMonster
		h.DB.Where("id IN ?", engagedMonsterIDs).Find(&monsters)
		monsterData = make([]gin.H, len(monsters))
		for i, m := range monsters {
			monsterData[i] = gin.H{
				"id":         m.ID,
				"current_hp": m.CurrentHP,
			}
		}
	}

	conn.SendJSON(gin.H{
		"type":       "zone_snapshot",
		"zone":       zone,
		"characters": charData,
		"monsters":   monsterData,
	})
}
