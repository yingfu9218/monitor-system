package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/monitor-system/internal/server/config"
	"github.com/monitor-system/internal/server/database"
	"github.com/monitor-system/internal/server/handler"
	"github.com/monitor-system/internal/server/middleware"
)

func main() {
	configPath := flag.String("config", "./configs/server-config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start background tasks
	go startBackgroundTasks(db, cfg)

	// Setup HTTP server
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	h := handler.New(db)

	// Frontend API (requires API Key)
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.Auth.APIKey))
	{
		api.POST("/auth/verify", h.VerifyAuth)
		api.GET("/servers", h.GetServers)
		api.GET("/servers/:id", h.GetServerDetail)
		api.GET("/servers/:id/history", h.GetHistory)
		api.GET("/servers/:id/disks", h.GetDisks)
		api.GET("/servers/:id/processes", h.GetProcesses)
		api.GET("/servers/:id/network", h.GetNetwork)
	}

	// Agent API (requires Agent Key)
	agent := r.Group("/api/v1/agent")
	agent.Use(middleware.AgentAuthMiddleware(cfg.Auth.AgentKey))
	{
		agent.POST("/report", h.AgentReport)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startBackgroundTasks(db *database.DB, cfg *config.Config) {
	// Update server status every 10 seconds
	statusTicker := time.NewTicker(10 * time.Second)
	go func() {
		for range statusTicker.C {
			if err := db.UpdateServerStatus(); err != nil {
				log.Printf("Failed to update server status: %v", err)
			}
		}
	}()

	// Cleanup old data daily
	cleanupTicker := time.NewTicker(time.Duration(cfg.Data.CleanupInterval) * time.Hour)
	go func() {
		for range cleanupTicker.C {
			if err := db.CleanupOldData(cfg.Data.RetentionDays); err != nil {
				log.Printf("Failed to cleanup old data: %v", err)
			} else {
				log.Printf("Cleaned up data older than %d days", cfg.Data.RetentionDays)
			}
		}
	}()
}
