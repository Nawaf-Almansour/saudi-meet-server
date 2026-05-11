package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/Nawaf-Almansour/diwan-server/internal/config"
	"github.com/Nawaf-Almansour/diwan-server/internal/db"
	"github.com/Nawaf-Almansour/diwan-server/internal/handlers"
	"github.com/Nawaf-Almansour/diwan-server/internal/middleware"
	"github.com/Nawaf-Almansour/diwan-server/internal/ws"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.Connect(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Handlers
	authHandler := handlers.NewAuthHandler(pool, cfg.Server.JWTSecret)
	workspaceHandler := handlers.NewWorkspaceHandler(pool)
	channelHandler := handlers.NewChannelHandler(pool)
	messageHandler := handlers.NewMessageHandler(pool, hub)
	meetingHandler := handlers.NewMeetingHandler(pool, hub, cfg.PlugNmeet.Host, cfg.PlugNmeet.APIKey, cfg.PlugNmeet.APISecret, cfg.PlugNmeet.ClientHost)
	wsHandler := handlers.NewWSHandler(hub, cfg.Server.JWTSecret)

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Diwan API",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// WebSocket (before auth middleware)
	app.Use("/api/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/api/ws", websocket.New(wsHandler.Handle))

	// Public routes
	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(cfg.Server.JWTSecret))
	protected.Get("/auth/me", authHandler.Me)

	// Workspaces
	protected.Post("/workspaces", workspaceHandler.Create)
	protected.Get("/workspaces", workspaceHandler.List)
	protected.Get("/workspaces/discover", workspaceHandler.ListAll)
	protected.Post("/workspaces/join", workspaceHandler.Join)
	protected.Get("/workspaces/:id", workspaceHandler.Get)

	// Channels
	protected.Post("/workspaces/:workspaceId/channels", channelHandler.Create)
	protected.Get("/workspaces/:workspaceId/channels", channelHandler.List)
	protected.Post("/channels/:id/join", channelHandler.Join)

	// Messages
	protected.Post("/channels/:id/messages", messageHandler.Send)
	protected.Get("/channels/:id/messages", messageHandler.List)

	// Meetings
	protected.Post("/channels/:id/meetings/start", meetingHandler.StartMeeting)
	protected.Post("/channels/:id/meetings/join", meetingHandler.JoinMeeting)

	log.Printf("Diwan server starting on port %d", cfg.Server.Port)
	if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
