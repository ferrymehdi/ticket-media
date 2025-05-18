package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxFileSize  = 2 * 1024 * 1024 // 2MB
	UploadDir    = "./uploads"
	MediaBaseURL = "http://localhost:8080/media" // Change as needed
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

func GetContentTypeFromBuffer(buffer []byte) string {
	contentType := http.DetectContentType(buffer)
	contentTypeParts := strings.Split(contentType, ";")
	return contentTypeParts[0]
}

type MediaHandler struct {
}

func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

func (h *MediaHandler) UploadMedia(c *fiber.Ctx) error {

	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}

	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Check file size
	if file.Size > MaxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File size exceeds the 2MB limit",
		})
	}

	// Get file type
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}
	defer src.Close()

	hasher := sha256.New()

	// Use a buffer to read and hash the file in chunks
	buffer := make([]byte, 32*1024) // 32KB chunks

	if _, err := io.CopyBuffer(hasher, src, buffer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process file",
		})
	}

	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Reset the file pointer to the beginning
	src.Seek(0, io.SeekStart)

	// Read a buffer to determine the content type
	contentTypeBuffer := make([]byte, 512)
	_, err = src.Read(contentTypeBuffer)
	if err != nil && err != io.EOF {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}

	// Reset the file pointer to the beginning again
	src.Seek(0, io.SeekStart)

	// Get content type
	contentType := GetContentTypeFromBuffer(contentTypeBuffer)

	// Check if the file type is allowed
	if !AllowedMimeTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File type not allowed. Only images are accepted.",
		})
	}

	// Get file extension from original filename
	extension := filepath.Ext(file.Filename)
	if extension == "" {
		// If no extension, derive it from MIME type
		switch contentType {
		case "image/jpeg":
			extension = ".jpg"
		case "image/png":
			extension = ".png"
		case "image/gif":
			extension = ".gif"
		case "image/webp":
			extension = ".webp"
		case "image/svg+xml":
			extension = ".svg"
		default:
			extension = ""
		}
	}
	// Create a filename with the hash and original extension
	filename := fileHash + extension
	filePath := filepath.Join(UploadDir, filename)

	// Check if file already exists (deduplication)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, save it
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save file",
			})
		}
	} else {
		// File already exists, no need to save again
		log.Printf("File with hash %s already exists, skipping save", fileHash)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"id":      filename,
		"url":     fmt.Sprintf("%s/%s", MediaBaseURL, filename),
		"hash":    fileHash,
	})
}
