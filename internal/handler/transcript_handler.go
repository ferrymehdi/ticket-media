package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

const (
	TranscriptUploadDir = "./transcripts"
)

type transcriptHander struct {
}

func NewTranscript() *transcriptHander {
	return &transcriptHander{}
}

func (h *transcriptHander) UploadTranscript(c *fiber.Ctx) error {
	file, err := c.FormFile("transcript")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No transcript file provided",
		})
	}

	if err := os.MkdirAll(TranscriptUploadDir, 0755); err != nil {
		log.Printf("Error creating transcript directory: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create upload directory",
		})
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening uploaded file: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create unique ID based on file content and timestamp
	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		log.Printf("Error calculating file hash: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process file",
		})
	}

	// Reset file pointer to beginning
	src, err = file.Open()
	if err != nil {
		log.Printf("Error reopening file: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process file",
		})
	}
	defer src.Close()

	// Generate ID
	fileHash := hex.EncodeToString(hasher.Sum(nil))
	fileID := fileHash[:12] // Use first 12 chars of hash as ID

	// Create filename with ID
	filename := fileID + filepath.Ext(file.Filename)
	filePath := filepath.Join(TranscriptUploadDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating destination file: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err = io.Copy(dst, src); err != nil {
		log.Printf("Error saving file: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Return success with file ID
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"id":       fileID,
		"filename": filename,
	})

}
