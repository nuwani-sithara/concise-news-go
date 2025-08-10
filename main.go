package main

import (
    "concise-news-go/handler" // fixed import path
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "log"
    "os"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found")
    }

    r := gin.Default()

    r.POST("/summarize", handler.SummarizeHandler)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}