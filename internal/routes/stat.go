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
	stat.Get("/activity", controllers.GetActivityStat(driver, logger))

	statWithAuth := group.Group("/stat")
	statWithAuth.Use(middlewares.JWTMiddleware(logger))

	statWithAuth.Get("/post", controllers.GetPostStat(driver, logger))
	statWithAuth.Get("/registry", controllers.GetRegistryStat(driver, logger))
	statWithAuth.Post("/generation", controllers.GetGenerationSTStat(driver, logger))
	statWithAuth.Get("/salary", controllers.GetUserSalary(driver, logger))
	statWithAuth.Get("/job", controllers.GetUserJob(driver, logger))
}
