package main

import (
	"log"

	"hexslayer/internal/db"
	"hexslayer/internal/game"
	"hexslayer/internal/handlers"
	"hexslayer/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database (migrate + seed)
	db.Init()

	// Start game engine (stub)
	engine := game.NewEngine()
	engine.Start()

	r := gin.Default()

	// CORS — allow frontend dev server
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Rate limiting middleware
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/api/health", handlers.Health)

	// API routes
	r.POST("/api/player/init", handlers.InitPlayer)
	r.GET("/api/map/zones", handlers.GetZones)
	r.POST("/api/character/deploy", handlers.DeployCharacter)

	// WebSocket
	r.GET("/ws", handlers.WebSocketHandler)

	log.Println("starting hexslayer server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
