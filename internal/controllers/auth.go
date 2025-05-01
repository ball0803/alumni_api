package controllers

import (
	"alumni_api/internal/auth"
	"alumni_api/internal/models"
	"alumni_api/internal/repositories"
	"alumni_api/internal/services"
	"alumni_api/internal/validators"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func VerifyToken(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		ret := map[string]interface{}{
			"user_id":   claim.UserID,
			"user_role": claim.Role,
		}

		successMessage := "Verify Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, ret, logger)
	}
}

func Login(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.LoginRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		ok, err := services.UsernameVerify(c.Context(), driver, req.Username, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, err)
		}
		if !ok {
			return HandleError(c, fiber.StatusUnauthorized, "User Is Not Verify", logger, err)
		}

		user, err := repositories.Login(c.Context(), driver, req.Username, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		err = auth.CheckPasswordHash(req.Password, user.Password)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, "invalid password", logger, err)
		}

		token, err := auth.GenerateJWT(user.UserID, user.Role, int(user.AdmitYear))
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    token,
			HTTPOnly: true,
			// TODO: Turn back to strict when frontend in production
			// Secure:   true, // Enable in production (HTTPS only)
			// SameSite: "Strict",
			Secure:   false,
			SameSite: "None",
			Path:     "/",
			MaxAge:   24 * 60 * 60,
		})

		ret := map[string]interface{}{
			"token":     token,
			"user_id":   user.UserID,
			"user_role": user.Role,
		}

		successMessage := "Login Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, ret, logger)
	}
}

func Logout(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    "",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "None",
			Path:     "/",
			MaxAge:   -1,
		})
		successMessage := "Logout Succesfully"

		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func RegistryUser(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.RegistryRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		user, err := repositories.RegistryUser(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Registry Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func RegistryAlumnus(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.RegistryOneTimeRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		claim, err := auth.ParseOTRJWT(req.Token)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		user, err := repositories.RegistryAlumnus(c.Context(), driver, req, claim.Email, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}
		user_id, _ := user["user_id"]

		token, err := auth.GenerateJWT(user_id.(string), "alumnus", 0)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    token,
			HTTPOnly: true,
			// TODO: Turn back to strict when frontend in production
			// Secure:   true, // Enable in production (HTTPS only)
			// SameSite: "Strict",
			Secure:   false,
			SameSite: "None",
			Path:     "/",
			MaxAge:   24 * 60 * 60,
		})

		successMessage := "Registry Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func VerifyAccount(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		verifyToken := c.Query("token")
		// client := config.GetEnv("CLIENT", "")

		claim, err := auth.ParseVerification(verifyToken)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		exists, err := services.UserExist(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", claim.UserID), logger, nil)
		}

		data, err := repositories.VerifyAccount(c.Context(), driver, claim.UserID, claim.VerificationToken, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		Username, _ := data["username"]

		user, err := repositories.Login(c.Context(), driver, Username.(string), logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		JWT, err := auth.GenerateJWT(user.UserID, user.Role, int(user.AdmitYear))
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    JWT,
			HTTPOnly: true,
			// TODO: Turn back to strict when frontend in production
			// Secure:   true, // Enable in production (HTTPS only)
			// SameSite: "Strict",
			Secure:   false,
			SameSite: "None",
			Path:     "/",
			MaxAge:   24 * 60 * 60,
		})

		ret := map[string]interface{}{
			"token":     JWT,
			"user_id":   user.UserID,
			"user_role": user.Role,
		}

		successMessage := "Login Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, ret, logger)
		// return c.Redirect(fmt.Sprintf("%s/registry", client), fiber.StatusFound)
	}
}

func RequestChangePassword(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.EmailRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		data, err := repositories.RequestChangePassword(c.Context(), driver, req.Email, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Request Reset Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

func ChangePassword(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.ResetPassword

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		claim, err := auth.ParseVerification(req.ResetJWT)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		exists, err := services.UserExist(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", claim.UserID), logger, nil)
		}

		err = repositories.ChangePassword(c.Context(), driver, claim.UserID, req.Password, claim.VerificationToken, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    "",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "None",
			Path:     "/",
			MaxAge:   -1,
		})

		successMessage := "Change Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func RequestChangeEmail(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.EmailRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		exists, err := services.EmailExist(c.Context(), driver, req.Email, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("Email %s Already Exist", req.Email), logger, nil)
		}

		exists, err = services.UserExist(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", claim.UserID), logger, nil)
		}

		data, err := repositories.RequestChangeMail(c.Context(), driver, claim.UserID, req.Email, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Request Reset Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

func VerifyEmail(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Query("token")

		claim, err := auth.ParseVerification(token)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		exists, err := services.UserExist(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", claim.UserID), logger, nil)
		}

		err = repositories.VerifyEmail(c.Context(), driver, claim.UserID, claim.Email, claim.VerificationToken, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Change Email Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func RequestAlumniOneTimeRegistry(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.EmailRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		data, err := repositories.RequestAlumniOneTimeRegistry(c.Context(), driver, req.Email, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "If this email match in the database the one time request will be send to your email"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

func RequestAlumnusRole(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		exists, err := services.UserExist(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", claim.UserID), logger, nil)
		}

		err = repositories.RequestAlumnusRole(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Get Request Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func ConfirmAlumnusRole(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		request_id := c.Params("request_id")

		if err := validators.UUID(request_id); err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err := repositories.ConfirmAlumnusRole(c.Context(), driver, request_id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Request Email Checkup Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func GetAllRequest(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		data, err := repositories.GetAllRequest(c.Context(), driver, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Get Request Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}
