package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"whatsapp-reminder/backend/internal/api"
	"whatsapp-reminder/backend/internal/auth"
	"whatsapp-reminder/backend/internal/database"
	"whatsapp-reminder/backend/internal/reminder"
	"whatsapp-reminder/backend/internal/scheduler"
	"whatsapp-reminder/backend/internal/whatsapp"
)

func main() {
	// Try loading .env from current directory or parent directory
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "data/reminders.db")
	waDBPath := getEnv("WA_DB_PATH", "data/whatsapp.db")
	backupDir := getEnv("BACKUP_DIR", "data/backups")
	staticDir := getEnv("STATIC_DIR", "frontend/dist")
	defaultTZ := getEnv("DEFAULT_TIMEZONE", "Asia/Jakarta")

	// If staticDir doesn't exist relative to current dir, try ../frontend/dist
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		altStatic := filepath.Join("..", staticDir)
		if _, err := os.Stat(altStatic); err == nil {
			staticDir = altStatic
		}
	}

	log.Printf("[Init] Starting WhatsApp Reminder service on port %s...\n", port)

	// 1. Initialize SQLite Database
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("[Init] Fatal: Failed to initialize SQLite database: %v", err)
	}
	defer db.Close()
	log.Printf("[Init] Database connected at %s\n", dbPath)

	// 2. Initialize Auth Service
	authSvc, err := auth.NewAuthService(db)
	if err != nil {
		log.Fatalf("[Init] Fatal: Failed to initialize auth service: %v", err)
	}

	// 3. Initialize Reminder Service
	remSvc := reminder.NewService(db)

	// 4. Initialize WhatsApp Command Handler & Client
	cmdHandler := whatsapp.NewCommandHandler(remSvc, defaultTZ)
	waClient, err := whatsapp.NewClient(db, cmdHandler, waDBPath)
	if err != nil {
		log.Fatalf("[Init] Fatal: Failed to initialize WhatsApp client: %v", err)
	}

	// Start WhatsApp client (auto-reconnects if session already saved)
	if err := waClient.Start(); err != nil {
		log.Printf("[WhatsApp] Warning on startup connection: %v\n", err)
	} else {
		log.Printf("[WhatsApp] WhatsApp client initialized\n")
	}

	// 5. Initialize & Start Scheduler
	sched := scheduler.NewScheduler(remSvc, waClient, 15*time.Second)
	sched.Start()
	log.Println("[Scheduler] Reminder scheduler started")

	// 6. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName:      "WhatsApp Reminder API",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	handler := api.NewHandler(db, authSvc, remSvc, waClient, backupDir)
	api.SetupRouter(app, handler, staticDir)

	// 7. Start HTTP Server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("[Server] HTTP server listening on http://0.0.0.0:%s\n", port)
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			serverErrChan <- err
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("[Shutdown] Shutting down gracefully...")
	case err := <-serverErrChan:
		log.Fatalf("[Server] Server error: %v", err)
	}

	// Graceful shutdown sequence
	sched.Stop()
	log.Println("[Shutdown] Scheduler stopped")

	waClient.Disconnect()
	log.Println("[Shutdown] WhatsApp client disconnected")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("[Shutdown] Error during HTTP server shutdown: %v\n", err)
	}
	log.Println("[Shutdown] HTTP server stopped")

	log.Println("[Shutdown] WhatsApp Reminder stopped cleanly.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
