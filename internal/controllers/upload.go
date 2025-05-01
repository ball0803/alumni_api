package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

const uploadDir = "./uploads"

func Upload(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		// Validate size (e.g., limit to 5MB)
		if file.Size > 5*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File too large"})
		}

		// Optional: validate MIME type (basic check)
		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type"})
		}

		// Ensure upload directory exists
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		// Use a safe, unique filename
		extension := filepath.Ext(file.Filename)
		uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
		filePath := filepath.Join(uploadDir, uniqueName)

		// Save the file
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
		}

		ret := map[string]interface{}{
			"url": "/uploads/" + uniqueName,
		}

		// Return public URL
		return HandleSuccess(c, fiber.StatusOK, "Upload successfully", ret, logger)
	}
}
