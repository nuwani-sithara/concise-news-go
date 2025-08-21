// main.go
package main

import (
	"concise-news-go/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	r := gin.Default()

	// Configure CORS middleware - allow all origins for development
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // Allow all origins for development
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve the HTML file
	r.GET("/", func(c *gin.Context) {
		c.File("./index.html")
	})

	// Add a simple test endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check if API token is configured
		hfToken := os.Getenv("HF_API_TOKEN")
		status := "ok"
		if hfToken == "" {
			status = "ok_no_token"
		}
		
		c.JSON(200, gin.H{"status": status})
	})

	r.POST("/summarize", handler.SummarizeHandler)

	// Use a different port to avoid conflict
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // Changed from 8080 to 8090
	}
	
	log.Printf("Server starting on port %s", port)
	log.Printf("Hugging Face API token configured: %t", os.Getenv("HF_API_TOKEN") != "")
	
	r.Run(":" + port)
}