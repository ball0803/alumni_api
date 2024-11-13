package routes

import (
	"alumni_api/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, driver neo4j.DriverWithContext, logger *zap.Logger) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/users/:id", handlers.GetUserByID(driver, logger))

	app.Get("/users", handlers.FindUserByFilter(driver, logger))

	app.Post("/users", handlers.CreateUser(driver, logger))

	app.Put("/users/:id", handlers.UpdateUserByID(driver, logger))

	app.Delete("/users/:id", handlers.DeleteUserByID(driver, logger))

	app.Post("/users/:id/companies", handlers.AddUserCompany(driver, logger))

	app.Put("/users/:user_id/companies/:company_id", handlers.UpdateUserCompany(driver, logger))

	app.Delete("/users/:user_id/companies/:company_id", handlers.DeleteUserCompany(driver, logger))

	app.Post("/users/:id/student_info", handlers.AddStudentInfo(driver, logger))

	app.Put("/users/:id/student_info", handlers.UpdateStudentInfo(driver, logger))

	app.Delete("/users/:id/student_info", handlers.DeleteStudentInfo(driver, logger))

	app.Get("/users/:id/friends", handlers.GetUserFriendByID(driver, logger))

	app.Post("/users/:id/friends", handlers.AddFriend(driver, logger))

	app.Delete("/users/:id/friends", handlers.Unfriend(driver, logger))
}
