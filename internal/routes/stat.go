package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func StatRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	stat := group.Group("/stat")
	stat.Use(middlewares.JWTMiddleware(logger))

	stat.Get("/post", controllers.GetPostStat(driver, logger))
	stat.Get("/registry", controllers.GetRegistryStat(driver, logger))
	stat.Get("/generation", controllers.GetGenerationSTStat(driver, logger))
	stat.Get("/job", controllers.GetUserJob(driver, logger))
}
