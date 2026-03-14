package main

import (
	"log"

	_ "hexslayer/docs"
	"hexslayer/internal/db"
	"hexslayer/internal/game"
	"hexslayer/internal/handlers"
	"hexslayer/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title HexSlayer API
// @version 1.0
// @description Idle geo-based monster hunting game API
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Initialize database (migrate + seed)
	db.Init()

	// Start game engine (stub)
	engine := game.NewEngine()
	engine.Start()

	r := gin.Default()

	// Rate limiting middleware
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/api/health", handlers.Health)

	// Public routes
	r.POST("/api/player/init", handlers.InitPlayer)

	// Protected routes (require Bearer token)
	auth := r.Group("/")
	auth.Use(middleware.SessionAuth())
	auth.GET("/api/map/zones", handlers.GetZones)
	auth.POST("/api/character/deploy", handlers.DeployCharacter)

	// WebSocket (token validated in handler via query param)
	r.GET("/ws", handlers.WebSocketHandler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("starting hexslayer server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
