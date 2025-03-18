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

func Login(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.LoginRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		user, err := repositories.Login(c.Context(), driver, req.Username, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		ok, err := services.UserVerify(c.Context(), driver, user.UserID, logger)
		fmt.Println(ok)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, err)
		}
		if !ok {
			return HandleError(c, fiber.StatusUnauthorized, "User Is Not Verify", logger, err)
		}

		err = auth.CheckPasswordHash(req.Password, user.Password)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, "invalid password", logger, err)
		}

		token, err := auth.GenerateJWT(user.UserID, user.Role, int(user.AdmitYear))
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		ret := map[string]interface{}{
			"token":     token,
			"user_id":   user.UserID,
			"user_role": user.Role,
		}

		successMessage := "Login Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, ret, logger)
	}
}

func Registry(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.ReqistryRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		user, err := repositories.Registry(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Registry Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func VerifyAccount(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
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

		err = repositories.VerifyAccount(c.Context(), driver, claim.UserID, claim.VerificationToken, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		return c.Redirect("http://localhost:3000/v1", fiber.StatusFound)
	}
}

func RequestChangePassword(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
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

		err = repositories.RequestChangePassword(c.Context(), driver, claim.UserID, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Request Reset Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
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

		successMessage := "Change Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func RequestChangeEmail(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.ChangeEmail

		if err := validators.Request(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

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

		err = repositories.RequestChangeMail(c.Context(), driver, claim.UserID, req.Email, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		successMessage := "Request Reset Password Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
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

		err = repositories.VerifyEmail(c.Context(), driver, claim.UserID, claim.VerificationToken, logger)
		if err != nil {
			return HandleError(c, fiber.StatusUnauthorized, err.Error(), logger, nil)
		}

		return c.Redirect("http://localhost:3000/v1", fiber.StatusFound)
	}
}
