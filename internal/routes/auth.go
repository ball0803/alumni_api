package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AuthRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	auth := group.Group("/auth")

	auth.Post("/registry", controllers.Registry(driver, logger))
	auth.Post("/login", controllers.Login(driver, logger))
	auth.Get("/verify-account", controllers.VerifyAccount(driver, logger))

	authWithAuth := group.Group("/auth")
	authWithAuth.Use(middlewares.JWTMiddleware(logger))

	authWithAuth.Get("/request_reset_password", controllers.RequestChangePassword(driver, logger))
	authWithAuth.Post("/change_password", controllers.ChangePassword(driver, logger))
	authWithAuth.Get("/request_change_email", controllers.RequestChangeEmail(driver, logger))
	authWithAuth.Post("/verify-email", controllers.VerifyEmail(driver, logger))
}
