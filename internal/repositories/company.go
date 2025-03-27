package repositories

import (
	"alumni_api/internal/models"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AddUserCompany(ctx context.Context, driver neo4j.DriverWithContext, id string, companies models.UserRequestCompany, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Begin the transaction for adding or connecting companies
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	// Query to create or connect the company
	query := `
    MERGE (a:Company {name: $name})
    ON CREATE SET a.company_id = $companyID
    WITH a
    MATCH (u:UserProfile {user_id: $userID})
    MERGE (u)-[r:HAS_WORK_WITH]->(a)
    SET r.position = $position,
        r.salary_min = $salary_min,
        r.salary_max = $salary_max,
        r.created_timestamp = timestamp()
  `

	for _, company := range companies.Companies {
		// Generate a new companyID only if the company is being created
		params := map[string]interface{}{
			"companyID":  uuid.New().String(),
			"userID":     id,
			"name":       company.Company,
			"position":   company.Position.Raw,
			"salary_min": company.SalaryMin.Raw,
			"salary_max": company.SalaryMax.Raw,
		}
		_, err = tx.Run(ctx, query, params)
		if err != nil {
			logger.Error("Failed to create or connect user company info", zap.Error(err))
			_ = tx.Rollback(ctx)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect user company info")
		}
	}

	query = `
    MATCH (u:UserProfile {user_id: $userID})-[r:HAS_WORK_WITH]->(c:Company)
    RETURN
      c.name AS company,
      r.position AS position,
      r.salary_min AS salary_min,
      r.salary_max AS salary_max
  `

	params := map[string]interface{}{
		"userID": id,
	}

	result, err := tx.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create or connect user company info", zap.Error(err))
		_ = tx.Rollback(ctx)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect user company info"+err.Error())
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to collect results")
	}

	var user_companies []map[string]interface{}

	// Iterate over the records and prepare the results
	for _, record := range records {
		user_companies = append(user_companies, record.AsMap())
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return user_companies, nil
}

func UpdateUserCompany(ctx context.Context, driver neo4j.DriverWithContext, userID, companyID string, company models.UserCompanyUpdateRequest, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if both User and Company exist
	checkExistenceQuery := `
		MATCH (u:UserProfile {user_id: $userID})
		MATCH (c:Company {company_id: $companyID})
		RETURN u, c LIMIT 1
	`
	existsResult, err := session.Run(ctx, checkExistenceQuery, map[string]interface{}{
		"userID":    userID,
		"companyID": companyID,
	})
	if err != nil {
		logger.Error("Failed to check user or company existence", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user or company existence")
	}
	if !existsResult.Next(ctx) {
		logger.Warn("User or Company not found", zap.String("userID", userID), zap.String("companyID", companyID))
		return fiber.NewError(fiber.StatusNotFound, "User or Company not found")
	}

	// Update or create the relationship with position
	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			updateQuery := `
			MATCH (a:Company {company_id: $companyID})<-[r:HAS_WORK_WITH]-(u:UserProfile {user_id: $userID})
			SET r.position = $position,
				r.updated_timestamp = timestamp()
		`
			_, err := tx.Run(ctx, updateQuery, map[string]interface{}{
				"companyID": companyID,
				"userID":    userID,
				"position":  company.Position,
			})
			if err != nil {
				logger.Error("Failed to update user company info", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update user company info")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}

	return nil
}

func DeleteUserCompany(ctx context.Context, driver neo4j.DriverWithContext, userID, companyID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if both User and Company exist
	checkExistenceQuery := `
		MATCH (u:UserProfile {user_id: $userID})
		MATCH (c:Company {company_id: $companyID})
		RETURN u, c LIMIT 1
	`
	existsResult, err := session.Run(ctx, checkExistenceQuery, map[string]interface{}{
		"userID":    userID,
		"companyID": companyID,
	})
	if err != nil {
		logger.Error("Failed to check user or company existence", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user or company existence")
	}
	if !existsResult.Next(ctx) {
		logger.Warn("User or Company not found", zap.String("userID", userID), zap.String("companyID", companyID))
		return fiber.NewError(fiber.StatusNotFound, "User or Company not found")
	}

	// Update or create the relationship with position
	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			updateQuery := `
			MATCH (a:Company {company_id: $companyID})<-[r:HAS_WORK_WITH]-(u:UserProfile {user_id: $userID})
      DELETE r
		`
			_, err := tx.Run(ctx, updateQuery, map[string]interface{}{
				"companyID": companyID,
				"userID":    userID,
			})
			if err != nil {
				logger.Error("Failed to removed user company info", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to removed user company info")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}

	return nil
}

func FindCompanyAssociate(ctx context.Context, driver neo4j.DriverWithContext, company models.Company, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile)-->(c:Company {name: $name})
    RETURN
      u.user_id AS user_id,
      u.first_name + ' ' + u.last_name AS fullname,
      u.first_name_eng + ' ' + u.last_name_eng AS fullname_eng
  `

	params := map[string]interface{}{
		"name": company.Company,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	var associate []map[string]interface{}

	for _, record := range records {
		associate = append(associate, record.AsMap())
	}

	return associate, nil
}
