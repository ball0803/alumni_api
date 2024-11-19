package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"time"
)

func RequestLogger(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		message, _ := c.Locals("message").(string)
		body := c.Body()

		logger.Info("Request completed",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.IP()),
			zap.String("message", message),
			zap.ByteString("body", body),
		)

		return err
	}
}
