package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetZones(c *gin.Context) {
	// TODO: compute zones via h3.GridDisk(bangkokCenter, 2) and return with monster counts
	c.JSON(http.StatusOK, gin.H{
		"message": "zones endpoint — not yet implemented",
		"zones":   []interface{}{},
	})
}
