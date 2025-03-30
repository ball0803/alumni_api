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

	auth.Post("/registry/user", controllers.RegistryUser(driver, logger))
	auth.Post("/registry/alumnus", controllers.RegistryAlumnus(driver, logger))
	auth.Post("/login", controllers.Login(driver, logger))
	auth.Get("/verify-account", controllers.VerifyAccount(driver, logger))
	auth.Get("/check_alumni_email", controllers.CheckAlumniExist(driver, logger))

	authWithAuth := group.Group("/auth")
	authWithAuth.Use(middlewares.JWTMiddleware(logger))

	authWithAuth.Get("/request/password_reset", controllers.RequestChangePassword(driver, logger))
	authWithAuth.Post("/request/password_reset/confirm", controllers.ChangePassword(driver, logger))
	authWithAuth.Get("/request/email_change", controllers.RequestChangeEmail(driver, logger))
	authWithAuth.Post("/request/email_change/confirm", controllers.VerifyEmail(driver, logger))

	authWithAuth.Get("/request", controllers.GetAllRequest(driver, logger))
	authWithAuth.Post("/request/role", controllers.RequestAlumnusRole(driver, logger))
	authWithAuth.Post("/request/:request_id/confirm", controllers.ConfirmAlumnusRole(driver, logger))
}
