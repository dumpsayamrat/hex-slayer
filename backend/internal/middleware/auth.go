package middleware

import (
	"net/http"
	"strings"

	"hexslayer/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SessionAuth validates the Bearer token from the Authorization header,
// looks up the player, and sets it in the Gin context.
func SessionAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		var player models.Player
		if err := db.Where("session_token = ?", token).First(&player).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session token"})
			return
		}

		c.Set("player", &player)
		c.Next()
	}
}

// GetPlayer retrieves the authenticated player from the Gin context.
func GetPlayer(c *gin.Context) *models.Player {
	p, exists := c.Get("player")
	if !exists {
		return nil
	}
	return p.(*models.Player)
}
