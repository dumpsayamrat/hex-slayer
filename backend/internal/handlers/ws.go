package handlers

import (
	"log"
	"net/http"

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

	// TODO: validate session token against DB

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("websocket connected: token=%s", token)

	// Send initial connection confirmation
	conn.WriteJSON(gin.H{
		"type":    "connected",
		"message": "welcome to hexslayer",
	})

	// Read loop — stub for handling client messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		msgType, _ := msg["type"].(string)
		log.Printf("received ws message: type=%s", msgType)

		// TODO: handle subscribe_zone, deploy_character, ping
		switch msgType {
		case "ping":
			conn.WriteJSON(gin.H{"type": "pong"})
		default:
			conn.WriteJSON(gin.H{"type": "error", "message": "unknown message type"})
		}
	}
}
