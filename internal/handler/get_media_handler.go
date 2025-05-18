package handler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validate ID
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No attachment ID provided",
		})
	}

	// Prevent directory traversal attacks
	if strings.Contains(id, "..") || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid attachment ID",
		})
	}

	filePath := filepath.Join(UploadDir, id)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Attachment not found",
		})
	}

	// Get content type based on file extension
	extension := filepath.Ext(id)
	contentType := GetContentType(extension)

	// Set cache headers
	c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	c.Set("Content-Type", contentType)

	// Serve the file
	return c.SendFile(filePath)
}

// GetContentType returns the content type based on file extension
func GetContentType(extension string) string {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
