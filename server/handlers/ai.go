package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/session"
)

// OpenAI API structures
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

// HandleGenerateQuestions generates natural questions for all fields using OpenAI
func HandleGenerateQuestions(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		// Get session
		sess, err := store.Get(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "session_not_found",
				Message: "Session not found.",
			})
			return
		}

		// Check for OpenAI API key
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "api_key_missing",
				Message: "OpenAI API key not configured.",
			})
			return
		}

		// Generate questions for all fields
		questions, err := generateQuestionsWithAI(sess.Fields, apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "ai_generation_failed",
				Message: "Failed to generate questions with AI: " + err.Error(),
			})
			return
		}

		// Update session with AI-generated questions
		err = store.Update(sessionID, func(s *models.Session) {
			for field, question := range questions {
				s.Questions[field] = question
			}
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "update_failed",
				Message: "Failed to save AI-generated questions.",
			})
			return
		}

		c.JSON(http.StatusOK, models.GenerateQuestionsResponse{
			Count:     len(questions),
			Questions: questions,
			Message:   "AI questions generated successfully.",
		})
	}
}

// generateQuestionsWithAI calls OpenAI API to generate natural questions
func generateQuestionsWithAI(fields []string, apiKey string) (map[string]string, error) {
	// Build prompt
	prompt := buildPrompt(fields)

	// Prepare OpenAI request
	reqBody := openAIRequest{
		Model: "gpt-4",
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are a helpful legal assistant that converts technical field names into natural, conversational questions. Always respond with valid JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON content from AI
	content := openAIResp.Choices[0].Message.Content
	var questions map[string]string
	if err := json.Unmarshal([]byte(content), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse AI-generated questions: %w", err)
	}

	return questions, nil
}

// buildPrompt creates the prompt for OpenAI
func buildPrompt(fields []string) string {
	fieldList := strings.Join(fields, "\n- ")

	return fmt.Sprintf(`I have a legal document with the following placeholder fields:
- %s

Please convert each field name into a natural, conversational question that I can ask a client.
The questions should be friendly, professional, and easy to understand.

Return ONLY a JSON object where keys are the field names and values are the questions.
Example format: {"field_name": "What is the field name you'd like to use?"}

Do not include any explanation, just the JSON object.`, fieldList)
}
