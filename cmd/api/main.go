package main

import (
	"log"

	"ecliptic.org/ticket-media/internal/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	app := fiber.New()

	app.Use(cors.New())

	mediaHandler := handler.NewMediaHandler()
	transcriptHandler := handler.NewTranscript()

	app.Post("/media/upload", mediaHandler.UploadMedia)
	app.Get("/media/:id", mediaHandler.GetMedia)
	app.Post("/transcript/upload", transcriptHandler.UploadTranscript)
	//app.Get("/transcript/")
	log.Fatal(app.Listen(":8080"))

	log.Printf("Starting server on :8080")
}
