package middlewares

import (
	"alumni_api/config"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

var redCache = cache.New(1*time.Minute, 2*time.Minute)

const windowDuration = time.Second

func REDWithQueueMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		route := c.Path()
		cfg, exists := config.REDRouteConfigs[route]
		if !exists {
			cfg = config.RouteREDConfig{
				MinThreshold: 80,
				MaxRequests:  100,
			}
		}

		ip := c.IP()
		cacheKey := route + ":" + ip

		countRaw, found := redCache.Get(cacheKey)
		count := 0
		if found {
			count = countRaw.(int)
		}
		count++
		redCache.Set(cacheKey, count, windowDuration)
		if count < cfg.MinThreshold {
			return c.Next()
		} else if count <= cfg.MaxRequests {
			prob := float64(count-cfg.MinThreshold) / float64(cfg.MaxRequests-cfg.MinThreshold)
			if rand.Float64() < prob {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"message": "Queue full, drop",
				})
			}
			return c.Next()
		} else {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "Queue full, hard drop",
			})
		}
	}
}
