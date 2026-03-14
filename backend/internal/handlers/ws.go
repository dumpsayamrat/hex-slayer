package handlers

import (
	"log"
	"net/http"

	"hexslayer/internal/db"
	"hexslayer/internal/models"
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

func WebSocketHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Validate session token against DB
	var player models.Player
	if err := db.DB.Where("session_token = ?", token).First(&player).Error; err != nil {
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
		ws.Hub.UnsubscribeAll(conn)
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
			ws.Hub.Subscribe("zone:"+zone, conn)
			conn.SendJSON(gin.H{"type": "subscribed", "h3_zone": zone})

		case "unsubscribe_zone":
			zone, _ := msg["h3_zone"].(string)
			if zone == "" {
				conn.SendJSON(gin.H{"type": "error", "message": "h3_zone required"})
				continue
			}
			ws.Hub.Unsubscribe("zone:"+zone, conn)
			conn.SendJSON(gin.H{"type": "unsubscribed", "h3_zone": zone})

		default:
			conn.SendJSON(gin.H{"type": "error", "message": "unknown message type"})
		}
	}
}
