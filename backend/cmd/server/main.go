package main

import (
	"log"

	_ "hexslayer/docs"
	"hexslayer/internal/db"
	"hexslayer/internal/game"
	"hexslayer/internal/handlers"
	"hexslayer/internal/middleware"
	"hexslayer/internal/services"
	"hexslayer/internal/ws"

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
	// 1. Infrastructure
	database := db.Init()
	hub := ws.NewHub()

	// 2. Services
	playerSvc := services.NewPlayerService(database)
	zoneSvc := services.NewZoneService(database)
	charSvc := services.NewCharacterService(database)

	// 3. Game engine
	engine := game.NewEngine(database, hub)
	engine.Start()

	// 4. Handlers
	h := handlers.New(database, hub, engine, playerSvc, zoneSvc, charSvc)

	// 5. Router
	r := gin.Default()
	r.Use(middleware.RateLimit())

	r.GET("/api/health", h.Health)
	r.POST("/api/player/init", h.InitPlayer)

	auth := r.Group("/")
	auth.Use(middleware.SessionAuth(database))
	auth.GET("/api/map/zones", h.GetZones)
	auth.POST("/api/character/deploy", h.DeployCharacter)

	r.GET("/ws", h.WebSocketHandler)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("starting hexslayer server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
