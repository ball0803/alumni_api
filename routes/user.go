package routes

import (
	"alumni_api/handlers"
	"alumni_api/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UserRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	user := group.Group("/users")

	user.Use(middlewares.JWTMiddleware(logger))

	user.Get("/:id", handlers.GetUserByID(driver, logger))

	user.Get("", handlers.FindUserByFilter(driver, logger))

	user.Post("", handlers.CreateUser(driver, logger))

	user.Put("/:id", handlers.UpdateUserByID(driver, logger))

	user.Delete("/:id", handlers.DeleteUserByID(driver, logger))

	user.Post("/:id/companies", handlers.AddUserCompany(driver, logger))

	user.Put("/:user_id/companies/:company_id", handlers.UpdateUserCompany(driver, logger))

	user.Delete("/:user_id/companies/:company_id", handlers.DeleteUserCompany(driver, logger))

	user.Post("/:id/student_info", handlers.AddStudentInfo(driver, logger))

	user.Put("/:id/student_info", handlers.UpdateStudentInfo(driver, logger))

	user.Delete("/:id/student_info", handlers.DeleteStudentInfo(driver, logger))

	user.Get("/:id/friends", handlers.GetUserFriendByID(driver, logger))

	user.Post("/:id/friends", handlers.AddFriend(driver, logger))

	user.Delete("/:id/friends", handlers.Unfriend(driver, logger))
}
