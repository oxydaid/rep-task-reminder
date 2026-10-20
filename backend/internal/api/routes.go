package api

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(app *fiber.App, h *Handler, staticDir string) {
	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: false,
	}))

	// Health check
	app.Get("/health", h.Health)

	// Public Auth routes
	authGroup := app.Group("/api/auth")
	authGroup.Get("/status", h.GetAuthStatus)
	authGroup.Post("/setup", h.SetupPIN)
	authGroup.Post("/login", h.Login)

	// Protected routes
	protected := app.Group("/api", AuthMiddleware(h.authSvc))

	// Auth actions
	protected.Post("/auth/logout", h.Logout)
	protected.Post("/auth/change-pin", h.ChangePIN)

	// Dashboard & Stats
	protected.Get("/stats", h.GetStats)

	// Reminders
	protected.Get("/reminders/upcoming", h.GetUpcomingReminder)
	protected.Get("/reminders", h.ListReminders)
	protected.Get("/reminders/:id", h.GetReminder)
	protected.Post("/reminders", h.CreateReminder)
	protected.Put("/reminders/:id", h.UpdateReminder)
	protected.Delete("/reminders/:id", h.DeleteReminder)
	protected.Patch("/reminders/:id/toggle", h.ToggleReminder)

	// Delivery Logs
	protected.Get("/logs", h.ListLogs)

	// WhatsApp connection
	protected.Get("/whatsapp/status", h.GetWhatsAppStatus)
	protected.Post("/whatsapp/connect", h.ConnectWhatsApp)
	protected.Post("/whatsapp/disconnect", h.DisconnectWhatsApp)
	protected.Post("/whatsapp/logout", h.LogoutWhatsApp)
	protected.Post("/whatsapp/test", h.TestSendWhatsApp)

	// Settings & Backup
	protected.Get("/settings", h.GetSettings)
	protected.Post("/settings", h.UpdateSettings)
	protected.Post("/backup", h.TriggerBackup)

	// Static SPA fallback for production
	if staticDir != "" {
		if fi, err := os.Stat(staticDir); err == nil && fi.IsDir() {
			app.Static("/", staticDir)

			// Single Page Application rewrite for client-side routing
			app.Get("/*", func(c *fiber.Ctx) error {
				path := c.Path()
				// Don't intercept API calls
				if len(path) >= 4 && path[:4] == "/api" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Endpoint tidak ditemukan"})
				}
				return c.SendFile(filepath.Join(staticDir, "index.html"))
			})
		}
	}
}
