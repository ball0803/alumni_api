package middlewares

import (
	"alumni_api/internal/auth"
	"alumni_api/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func JWTMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return controllers.HandleFail(c, fiber.StatusUnauthorized, "Missing Token", logger, nil)
		}

		tokenString := authHeader[len("Bearer "):]
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			return controllers.HandleFail(c, fiber.StatusUnauthorized, "Invalid or expired token", logger, nil)
		}

		// Store claims in the context
		c.Locals("claims", claims)
		return c.Next()
	}
}

// Authenticated WebSocket Upgrade Middleware
func WebSocketMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from headers
		token := c.Get("Sec-WebSocket-Protocol") // WebSocket can't send Authorization header
		if token == "" {
			token = c.Query("token") // Fallback to query params
		}

		// Validate JWT
		claims, err := auth.ParseJWT(token)
		if err != nil {
			return controllers.HandleFail(c, fiber.StatusUnauthorized, "Invalid or expired token", logger, nil)
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
