package middlewares

import (
	"alumni_api/auth"
	"alumni_api/handlers"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func JWTMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return handlers.HandleFail(c, fiber.StatusUnauthorized, "Missing Token", logger, nil)
		}

		tokenString := authHeader[len("Bearer "):]
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			return handlers.HandleFail(c, fiber.StatusUnauthorized, "Invalid or expired token", logger, nil)
		}

		// Store claims in the context
		c.Locals("claims", claims)
		return c.Next()
	}
}
