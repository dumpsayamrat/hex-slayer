package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeployCharacterRequest struct {
	H3Zone string `json:"h3_zone" binding:"required"`
}

// DeployCharacter godoc
// @Summary Deploy a character to a zone
// @Description Deploy a new character to the specified H3 zone
// @Tags character
// @Accept json
// @Produce json
// @Param body body DeployCharacterRequest true "Deploy request"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/character/deploy [post]
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
