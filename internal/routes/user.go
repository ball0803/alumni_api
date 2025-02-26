package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UserRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {

	// Public routes
	user := group.Group("/users")
	user.Get("/search", controllers.FindUserByFilter(driver, logger))
	user.Get("/fulltext_search", controllers.NameFullTextSearch(driver, logger))
	user.Get("/company_associate", controllers.FindCompanyAssociate(driver, logger))

	// Authenticated routes
	userWithAuth := group.Group("/users")
	userWithAuth.Use(middlewares.JWTMiddleware(logger))

	// User endpoints
	userWithAuth.Post("/", controllers.CreateProfile(driver, logger))
	userWithAuth.Get("/:id", controllers.GetUserByID(driver, logger))
	userWithAuth.Put("/:id", controllers.UpdateUserByID(driver, logger))
	userWithAuth.Delete("/:id", controllers.DeleteUserByID(driver, logger))

	// Companies endpoints
	userWithAuth.Post("/:id/companies", controllers.AddUserCompany(driver, logger))
	userWithAuth.Put("/:user_id/companies/:company_id", controllers.UpdateUserCompany(driver, logger))
	userWithAuth.Delete("/:user_id/companies/:company_id", controllers.DeleteUserCompany(driver, logger))

	// Student info endpoints
	userWithAuth.Post("/:id/student_info", controllers.AddStudentInfo(driver, logger))
	userWithAuth.Put("/:id/student_info", controllers.UpdateStudentInfo(driver, logger))
	userWithAuth.Delete("/:id/student_info", controllers.DeleteStudentInfo(driver, logger))

	// Friends endpoints
	userWithAuth.Get("/:id/friends", controllers.GetUserFriendByID(driver, logger))
	userWithAuth.Post("/:id/friends", controllers.AddFriend(driver, logger))
	userWithAuth.Delete("/:id/friends", controllers.Unfriend(driver, logger))

	userWithAuth.Get("/:user_id/foaf/:other_id", controllers.GetFOAF(driver, logger))
}
