package handlers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

const uploadDir = "/var/www/html/uploads/temp"

func Upload(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		// Define the file path
		filePath := filepath.Join(uploadDir, file.Filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
		}

		successMessage := "Upload Sucessfully"
		return HandleSuccess(
			c,
			fiber.StatusOK,
			successMessage,
			c.JSON(fiber.Map{"url": "/uploads/" + file.Filename}),
			logger)
	}
}
