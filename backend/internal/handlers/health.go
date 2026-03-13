package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health godoc
// @Summary Health check
// @Description Returns server status
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"game":   "hexslayer",
	})
}
