package repositories

import (
	"alumni_api/internal/models"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AddStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, college_info models.CollegeInfo, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	query := `
    MERGE (faculty:Faculty {name: $faculty})
    MERGE (faculty)-[:HAS_DEPARTMENT]->(department:Department {name: $department})
    MERGE (department)-[:HAS_FIELD]->(field:Field {name: $field})
    MERGE (field)-[:HAS_STUDENT_TYPE]->(studentType:StudentType {name: $studentType})

    WITH field, studentType
    MATCH (u:UserProfile {user_id: $userID})
    MERGE (u)-[:BELONGS_TO_FIELD]->(field)
    MERGE (u)-[:BELONGS_TO_STUDENT_TYPE]->(studentType)
  `

	params := map[string]interface{}{
		"userID":      id,
		"faculty":     college_info.Faculty,
		"department":  college_info.Department,
		"field":       college_info.Field,
		"studentType": college_info.StudentType,
	}

	_, err2 := session.Run(ctx, query, params)
	if err2 != nil {
		logger.Error("Failed to create or connect student info", zap.Error(err2))
		return fiber.NewError(http.StatusInternalServerError, "Failed to create or connect student info")
	}

	return nil

}

func UpdateStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, college_info models.CollegeInfo, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	query := `
    MATCH (u:UserProfile {user_id: $userID})-[r: BELONGS_TO_FIELD|BELONGS_TO_STUDENT_TYPE]->()
    DELETE r

    MERGE (f:Faculty {name: $faculty})
    MERGE (f)-[:HAS_DEPARTMENT]->(d:Department {name: $department})
    MERGE (d)-[:HAS_FIELD]->(fld:Field {name: $field})
    MERGE (fld)-[:HAS_STUDENT_TYPE]->(st:StudentType {name: $studentType})

    MERGE (u)-[:BELONGS_TO_FIELD]->(fld)
    MERGE (u)-[:BELONGS_TO_STUDENT_TYPE]->(st)
  `

	params := map[string]interface{}{
		"userID":      id,
		"faculty":     college_info.Faculty,
		"department":  college_info.Department,
		"field":       college_info.Field,
		"studentType": college_info.StudentType,
	}

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to create or connect student info", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect student info")
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return nil
}

func DeleteStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	query := `
    MATCH (u:UserProfile {user_id: $userID})-[r: BELONGS_TO_FIELD|BELONGS_TO_STUDENT_TYPE]->()
    DELETE r
  `

	params := map[string]interface{}{
		"userID": id,
	}

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to remove student info", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect student info")
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return nil
}
