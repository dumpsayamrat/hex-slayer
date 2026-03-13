package middleware

import "github.com/gin-gonic/gin"

// RateLimit is a placeholder middleware for rate limiting.
// TODO: implement per-session rate limiting (1 deploy per 3s, WS message throttling)
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
