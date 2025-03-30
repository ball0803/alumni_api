package repositories

import (
	"alumni_api/internal/auth"
	"alumni_api/internal/models"
	"alumni_api/internal/utils"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func Login(ctx context.Context, driver neo4j.DriverWithContext, username string, logger *zap.Logger) (models.LoginResponse, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {username: $username})
    RETURN u.user_id AS user_id, u.user_password AS user_password, u.role AS role,
      u.admit_year AS admit_year
  `
	params := map[string]interface{}{
		"username": username,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusInternalServerError, "Error querying user")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("username", username))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}

	var res models.LoginResponse
	if err := utils.MapToStruct(record.AsMap(), &res); err != nil {
		logger.Error("Error decoding user properties", zap.Error(err))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusInternalServerError, "Error decoding user properties")
	}

	return res, nil
}

func RegistryAlumnus(ctx context.Context, driver neo4j.DriverWithContext, user models.ReqistryRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Start a transaction
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to start transaction", zap.Error(err))
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Defer a function to handle transaction rollback in case of failure
	defer func() {
		if err != nil {
			logger.Info("Rolling back transaction due to error", zap.Error(err))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	// Check if the username already exists
	checkQuery := `
        MATCH (u:UserProfile {username: $username})
        RETURN u.user_id AS user_id, u.is_verify AS is_verify, u.verification_token AS verification_token
    `
	checkParams := map[string]interface{}{"username": user.Username}
	checkResult, err := tx.Run(ctx, checkQuery, checkParams)
	if err != nil {
		logger.Error("Failed to check username uniqueness", zap.Error(err))
		return nil, fmt.Errorf("error checking username uniqueness: %w", err)
	}

	// Hash the password
	hashedPass, err := auth.HashPassword(user.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Generate a verification token
	token := auth.GenerateVerificationToken()

	// If the username exists
	if !checkResult.Next(ctx) {
		logger.Error("Alumni User don't exist")
		return nil, fmt.Errorf("Alumni User don't exist")
	}

	record := checkResult.Record()
	userID, _ := record.Get("user_id")
	isVerify, _ := record.Get("is_verify")
	jwtToken, err := auth.GenerateVerificationJWT(userID.(string), token)
	if err != nil {
		logger.Error("Failed to create verify jwt", zap.Error(err))
		return nil, fmt.Errorf("failed to create verify jwt: %w", err)
	}

	// If the user is already verified, return an error
	if isVerify.(bool) {
		logger.Warn("Username already exists and is verified", zap.String("username", user.Username))
		return nil, fmt.Errorf("username already exists and is verified")
	}

	// If the user is not verified, allow claiming the account
	logger.Info("Username exists but is not verified. Allowing claim.", zap.String("username", user.Username))

	// Update the existing user with the new password and verification token
	updateQuery := `
            MATCH (u:UserProfile {username: $username})
            SET u.user_password = $password,
                u.verification_token = $token
        `
	updateParams := map[string]interface{}{
		"username": user.Username,
		"password": hashedPass,
		"token":    token,
	}

	_, err = tx.Run(ctx, updateQuery, updateParams)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	// Send verification email
	if err = utils.SendVerificationEmail(user.Username, jwtToken); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
		// Rollback the transaction if email sending fails
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return nil, fmt.Errorf("error sending verification email: %w", err)
	}

	// Commit the transaction after sending the email
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	ret := map[string]interface{}{
		"user_id": userID,
	}
	return ret, nil
}

func RegistryUser(ctx context.Context, driver neo4j.DriverWithContext, user models.ReqistryRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Start a transaction
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to start transaction", zap.Error(err))
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Defer a function to handle transaction rollback in case of failure
	defer func() {
		if err != nil {
			logger.Info("Rolling back transaction due to error", zap.Error(err))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	// Check if the username already exists
	checkQuery := `
        MATCH (u:UserProfile {username: $username})
        RETURN u.user_id AS user_id, u.is_verify AS is_verify, u.verification_token AS verification_token
    `
	checkParams := map[string]interface{}{"username": user.Username}
	checkResult, err := tx.Run(ctx, checkQuery, checkParams)
	if err != nil {
		logger.Error("Failed to check username uniqueness", zap.Error(err))
		return nil, fmt.Errorf("error checking username uniqueness: %w", err)
	}

	// Hash the password
	hashedPass, err := auth.HashPassword(user.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Generate a verification token
	token := auth.GenerateVerificationToken()

	if checkResult.Next(ctx) {
		record := checkResult.Record()
		isVerify, _ := record.Get("is_verify")

		// If the user is already verified, return an error
		if isVerify.(bool) {
			logger.Warn("Username already exists and is verified", zap.String("username", user.Username))
			return nil, fmt.Errorf(")username already exists and is verified")
		}
	}

	// If the username does not exist, create a new user
	userID := uuid.New().String()

	jwtToken, err := auth.GenerateVerificationJWT(userID, token)
	if err != nil {
		logger.Error("Failed to create verify jwt", zap.Error(err))
		return nil, fmt.Errorf("failed to create verify jwt: %w", err)
	}

	createQuery := `
        CREATE (u:UserProfile {
            user_id: $userID,
            username: $username,
            user_password: $password,
            is_verify: false,
            verification_token: $token,
            role: $role
        })
        RETURN u.user_id AS user_id
    `
	createParams := map[string]interface{}{
		"userID":   userID,
		"username": user.Username,
		"password": hashedPass,
		"token":    token,
		"role":     "user",
	}

	createResult, err := tx.Run(ctx, createQuery, createParams)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	createRecord, err := createResult.Single(ctx)
	if err != nil {
		logger.Error("Failed to fetch created user ID", zap.Error(err))
		return nil, fmt.Errorf("error fetching created user ID: %w", err)
	}

	createdUserID, _ := createRecord.Get("user_id")

	// Commit the transaction before sending the email
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	// Send verification email
	if err = utils.SendVerificationEmail(user.Username, jwtToken); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
		// Rollback the transaction if email sending fails
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return nil, fmt.Errorf("error sending verification email: %w", err)
	}

	ret := map[string]interface{}{
		"user_id": createdUserID,
	}
	return ret, nil
}

func VerifyAccount(ctx context.Context, driver neo4j.DriverWithContext, user_id, token string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Query to fetch user details
	query := `
        MATCH (u:UserProfile {user_id: $user_id})
        RETURN u.is_verify AS is_verify, u.verification_token AS verification_token
    `
	params := map[string]interface{}{
		"user_id": user_id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return fmt.Errorf("error querying user: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("user_id", user_id))
		return fmt.Errorf("user not found: %w", err)
	}

	user := record.AsMap()

	// Check if the user is already verified
	if isVerify, ok := user["is_verify"].(bool); ok && isVerify {
		logger.Warn("User already verified", zap.String("user_id", user_id))
		return fmt.Errorf("user already verified")
	}

	// Check if the verification token matches
	if verificationToken, ok := user["verification_token"].(string); !ok || verificationToken != token {
		logger.Warn("Incorrect verification token", zap.String("user_id", user_id))
		return fmt.Errorf("incorrect verification token")
	}

	// Update the user to mark them as verified and clear the token
	updateQuery := `
        MATCH (u:UserProfile {user_id: $user_id})
        REMOVE u.verification_token
        SET u.is_verify = true
    `
	updateParams := map[string]interface{}{
		"user_id": user_id,
	}

	_, err = session.Run(ctx, updateQuery, updateParams)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func RequestChangePassword(ctx context.Context, driver neo4j.DriverWithContext, user_id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Start a transaction
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to start transaction", zap.Error(err))
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Defer a function to handle transaction rollback in case of failure
	defer func() {
		if err != nil {
			logger.Info("Rolling back transaction due to error", zap.Error(err))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	token := auth.GenerateVerificationToken()

	jwtToken, err := auth.GenerateVerificationJWT(user_id, token)
	if err != nil {
		logger.Error("Failed to create verify jwt", zap.Error(err))
		return fmt.Errorf("failed to create verify jwt: %w", err)
	}

	query := `
        MATCH (u:UserProfile {
            user_id: $userID
        })
        SET u.reset_password_token = $token
        RETURN u.username AS username
    `
	params := map[string]interface{}{
		"userID": user_id,
		"token":  token,
	}

	result, err := tx.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("error updating user: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("user_id", user_id))
		return fmt.Errorf("user not found: %w", err)
	}
	username, ok := record.Get("username")
	if !ok {
		logger.Warn("Username not found", zap.String("user_id", user_id))
		return fmt.Errorf("Username not found: %w", err)
	}

	// Send verification email
	if err = utils.SendResetMail(username.(string), jwtToken); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
		// Rollback the transaction if email sending fails
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return fmt.Errorf("error sending verification email: %w", err)
	}

	// Commit the transaction after sending the email
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func ChangePassword(ctx context.Context, driver neo4j.DriverWithContext, user_id, password, token string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Query to fetch user details
	query := `
        MATCH (u:UserProfile {user_id: $user_id})
        RETURN u.reset_password_token AS reset_password_token
    `
	params := map[string]interface{}{
		"user_id": user_id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return fmt.Errorf("error querying user: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("user_id", user_id))
		return fmt.Errorf("user not found: %w", err)
	}

	user := record.AsMap()

	// Check if the verification token matches
	if resetToken, ok := user["reset_password_token"].(string); !ok || resetToken != token {
		logger.Warn("Incorrect Reset token", zap.String("user_id", user_id))
		return fmt.Errorf("incorrect Reset token")
	}

	hashedPass, err := auth.HashPassword(password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update the user to mark them as verified and clear the token
	updateQuery := `
        MATCH (u:UserProfile {user_id: $user_id})
        REMOVE u.reset_password_token
        SET u.user_password = $user_password
    `
	updateParams := map[string]interface{}{
		"user_id":       user_id,
		"user_password": hashedPass,
	}

	_, err = session.Run(ctx, updateQuery, updateParams)
	if err != nil {
		logger.Error("Failed to update user password", zap.Error(err))
		return fmt.Errorf("error updating user password: %w", err)
	}

	return nil
}

func RequestChangeMail(ctx context.Context, driver neo4j.DriverWithContext, user_id, email string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Start a transaction
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to start transaction", zap.Error(err))
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Defer a function to handle transaction rollback in case of failure
	defer func() {
		if err != nil {
			logger.Info("Rolling back transaction due to error", zap.Error(err))
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	token := auth.GenerateVerificationToken()

	jwtToken, err := auth.GenerateVerificationJWT(user_id, token)
	if err != nil {
		logger.Error("Failed to create verify jwt", zap.Error(err))
		return fmt.Errorf("failed to create verify jwt: %w", err)
	}

	query := `
        MATCH (u:UserProfile {
            user_id: $userID
        })
        SET u.change_email_token = $token
    `
	params := map[string]interface{}{
		"userID": user_id,
		"token":  token,
	}

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("error updating user: %w", err)
	}

	// Send verification email
	if err = utils.SendVerificationChangeEmail(email, jwtToken); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
		// Rollback the transaction if email sending fails
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", zap.Error(rollbackErr))
		}
		return fmt.Errorf("error sending verification email: %w", err)
	}

	// Commit the transaction after sending the email
	if err = tx.Commit(ctx); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func VerifyEmail(ctx context.Context, driver neo4j.DriverWithContext, user_id, token string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Query to fetch user details
	query := `
        MATCH (u:UserProfile {user_id: $user_id})
        RETURN u.change_email_token AS change_email_token
    `
	params := map[string]interface{}{
		"user_id": user_id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return fmt.Errorf("error querying user: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("user_id", user_id))
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if the verification token matches
	if mailToken, ok := record.Get("change_email_token"); !ok || mailToken.(string) != token {
		logger.Warn("Incorrect verification token", zap.String("user_id", user_id))
		return fmt.Errorf("incorrect verification token")
	}

	// Update the user to mark them as verified and clear the token
	updateQuery := `
        MATCH (u:UserProfile {user_id: $user_id})
        REMOVE u.change_email_token
    `
	updateParams := map[string]interface{}{
		"user_id": user_id,
	}

	_, err = session.Run(ctx, updateQuery, updateParams)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func CheckExistAlumni(ctx context.Context, driver neo4j.DriverWithContext, email string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {email: $email})
    RETURN COUNT(u) > 0 AS userExist
  `
	params := map[string]interface{}{
		"email": email,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error querying user")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("email", email))
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}

	userExists, ok := record.Get("userExists")
	if !ok {
		logger.Warn("User not found")
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error Using the Query")
	}

	ret := map[string]interface{}{"isUserExist": userExists.(bool)}

	return ret, nil
}

func ConfirmAlumnusRole(ctx context.Context, driver neo4j.DriverWithContext, request_id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile)-[:HAS_REQUEST]->(r:Request {request_id: $request_id})
    SET u.role = "alumnus"
    DETACH DELETE r
  `
	params := map[string]interface{}{
		"request_id": request_id,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return fmt.Errorf("error querying user: %w", err)
	}

	return nil
}

func RequestAlumnusRole(ctx context.Context, driver neo4j.DriverWithContext, user_id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	request_id := uuid.New().String()

	query := `
    MATCH (u:UserProfile {user_id: $user_id})
    MERGE (r:Request {
      type: "role_request",
      timestamp: timestamp()
    })<-[:HAS_REQUEST]-(u)
    ON CREATE SET
      r.request_id = $request_id
  `
	params := map[string]interface{}{
		"user_id":    user_id,
		"request_id": request_id,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return fmt.Errorf("error querying user: %w", err)
	}

	return nil
}

func GetAllRequest(ctx context.Context, driver neo4j.DriverWithContext, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (user:UserProfile)-[:HAS_REQUEST]->(request:Request)
    OPTIONAL MATCH (user)-[workRel:HAS_WORK_WITH]->(company:Company)
    OPTIONAL MATCH (user)-->(studentType:StudentType)<--(field:Field)<--(department:Department)<--(faculty:Faculty)

    // First collect all companies for each user
    WITH user, request, faculty, department, field, studentType,
        collect(
          CASE WHEN company IS NOT NULL THEN {
            company: company.name,
            address: company.address,
            position: workRel.position
          } ELSE null END
        ) AS companies

    // Then build the final result structure
    RETURN {
      user: {
        user_id: user.user_id,
        username: user.username,
        gender: user.gender,
        dob: toString(user.dob),
        name: user.first_name + " " + user.last_name,
        name_eng: user.first_name_eng + " " + user.last_name_eng,
        profile_picture: user.profile_picture,
        role: user.role,
        student_info: {
          faculty: faculty.name,
          department: department.name,
          field: field.name,
          student_type: studentType.name,
          education_level: user.education_level,
          admit_year: user.admit_year,
          graduate_year: user.graduate_year,
          gpax: user.gpax
        },
        companies: companies,
        contact_info: {
          email: user.email,
          github: user.github,
          linkedin: user.linkdin,
          facebook: user.facebook,
          phone: user.phone
        }
      },
      request: {
        type: request.type,
        request_id: request.request_id,
        timestamp: request.timestamp
      }
    } AS result
  `

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error querying user: %s", err))
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect query results", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error retrieving data")
	}

	var requests []map[string]interface{}

	for _, record := range records {
		request, _ := record.Get("result")
		utils.CleanNullValues(request)
		requests = append(requests, request.(map[string]interface{}))
	}

	return requests, nil
}
