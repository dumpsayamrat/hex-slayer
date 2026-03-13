package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeployCharacterRequest struct {
	H3Zone string `json:"h3_zone" binding:"required"`
}

func DeployCharacter(c *gin.Context) {
	var req DeployCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: validate session, check character cap, create character in zone
	c.JSON(http.StatusOK, gin.H{
		"message": "deploy character endpoint — not yet implemented",
	})
}
