// handler/summarize.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type SummarizeRequest struct {
	Text string `json:"text" binding:"required"`
}

type SummarizeResponse struct {
	Summary string `json:"summary"`
}

func SummarizeHandler(c *gin.Context) {
	var req SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text field is required"})
		return
	}

	// Validate input text
	if strings.TrimSpace(req.Text) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text cannot be empty"})
		return
	}

	if len(strings.Split(req.Text, " ")) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text should be at least 10 words long"})
		return
	}

	hfToken := os.Getenv("HF_API_TOKEN")

	// If no API token is set, return a mock response for testing
	if hfToken == "" {
		// Create a mock summary for testing
		words := strings.Fields(req.Text)
		if len(words) > 50 {
			words = words[:50] // Take first 50 words
		}
		mockSummary := strings.Join(words, " ") + "... (mock summary - set HF_API_TOKEN for real summaries)"

		c.JSON(http.StatusOK, SummarizeResponse{Summary: mockSummary})
		return
	}

	client := resty.New().
		SetTimeout(30 * time.Second). // Increase timeout
		SetRetryCount(3).             // Add retries
		SetRetryWaitTime(2 * time.Second)

	// Use a summarization model from Hugging Face
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+hfToken).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"inputs": req.Text,
		}).
		Post("https://api-inference.huggingface.co/models/facebook/bart-large-cnn")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Hugging Face API: " + err.Error()})
		return
	}

	if resp.StatusCode() != http.StatusOK {
		// Handle specific Hugging Face API errors
		if resp.StatusCode() == 503 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Model is loading, please try again in a few seconds",
				"details": "The summarization model is currently loading. This usually takes 20-30 seconds on first request.",
			})
			return
		}

		c.JSON(resp.StatusCode(), gin.H{"error": fmt.Sprintf("Hugging Face API Error: %s", resp.String())})
		return
	}

	// Response is usually [{"summary_text":"..."}]
	var result []map[string]string
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Hugging Face response: " + err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response from Hugging Face API"})
		return
	}

	summary, exists := result[0]["summary_text"]
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format from Hugging Face API"})
		return
	}

	c.JSON(http.StatusOK, SummarizeResponse{Summary: summary})
}
