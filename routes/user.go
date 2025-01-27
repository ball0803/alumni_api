package routes

import (
	"alumni_api/handlers"
	"alumni_api/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UserRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {

	// Public routes
	user := group.Group("/users")
	user.Post("", handlers.CreateUser(driver, logger))

	// Authenticated routes
	userWithAuth := group.Group("/users")
	userWithAuth.Use(middlewares.JWTMiddleware(logger))

	// User endpoints
	userWithAuth.Get("/:id", handlers.GetUserByID(driver, logger))
	userWithAuth.Get("", handlers.FindUserByFilter(driver, logger))
	userWithAuth.Put("/:id", handlers.UpdateUserByID(driver, logger))
	userWithAuth.Delete("/:id", handlers.DeleteUserByID(driver, logger))

	// Companies endpoints
	userWithAuth.Post("/:id/companies", handlers.AddUserCompany(driver, logger))
	userWithAuth.Put("/:user_id/companies/:company_id", handlers.UpdateUserCompany(driver, logger))
	userWithAuth.Delete("/:user_id/companies/:company_id", handlers.DeleteUserCompany(driver, logger))

	// Student info endpoints
	userWithAuth.Post("/:id/student_info", handlers.AddStudentInfo(driver, logger))
	userWithAuth.Put("/:id/student_info", handlers.UpdateStudentInfo(driver, logger))
	userWithAuth.Delete("/:id/student_info", handlers.DeleteStudentInfo(driver, logger))

	// Friends endpoints
	userWithAuth.Get("/:id/friends", handlers.GetUserFriendByID(driver, logger))
	userWithAuth.Post("/:id/friends", handlers.AddFriend(driver, logger))
	userWithAuth.Delete("/:id/friends", handlers.Unfriend(driver, logger))
}
