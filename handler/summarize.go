package handler

import (
    "net/http"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/go-resty/resty/v2"
    "fmt"
    "encoding/json"
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

    hfToken := os.Getenv("HF_API_TOKEN")
    if hfToken == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Hugging Face API token not configured"})
        return
    }

    client := resty.New()

    // Use a summarization model from Hugging Face — e.g., facebook/bart-large-cnn
    resp, err := client.R().
        SetHeader("Authorization", "Bearer "+hfToken).
        SetHeader("Content-Type", "application/json").
        SetBody(map[string]string{
            "inputs": req.Text,
        }).
        Post("https://api-inference.huggingface.co/models/facebook/bart-large-cnn")

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if resp.StatusCode() != http.StatusOK {
        c.JSON(resp.StatusCode(), gin.H{"error": fmt.Sprintf("API Error: %s", resp.String())})
        return
    }

    // Response is usually [{"summary_text":"..."}]
    var result []map[string]string
    err = json.Unmarshal(resp.Body(), &result)
    if err != nil || len(result) == 0 {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Hugging Face response"})
        return
    }

    summary := result[0]["summary_text"]

    c.JSON(http.StatusOK, SummarizeResponse{Summary: summary})
}
