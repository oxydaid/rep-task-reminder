package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"whatsapp-reminder/backend/internal/auth"
)

func AuthMiddleware(authSvc *auth.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		var tokenStr string
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenStr = c.Cookies("auth_token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Autentikasi diperlukan. Silakan masukkan PIN terlebih dahulu.",
			})
		}

		claims, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Sesi telah kedaluwarsa atau tidak valid. Silakan login kembali.",
			})
		}

		c.Locals("user", claims.Subject)
		return c.Next()
	}
}
