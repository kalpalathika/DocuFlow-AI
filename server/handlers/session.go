package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/session"
)

// HandleGetSession returns the current session status
func HandleGetSession(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		sess, err := store.Get(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "session_not_found",
				Message: "Session not found. Please upload a document first.",
			})
			return
		}

		// Count answered fields
		answeredCount := len(sess.Answers)

		c.JSON(http.StatusOK, models.SessionStatusResponse{
			SessionID:   sess.ID,
			Fields:      sess.Fields,
			Answers:     sess.Answers,
			Questions:   sess.Questions,
			Progress:    answeredCount,
			Total:       len(sess.Fields),
			IsCompleted: answeredCount == len(sess.Fields),
		})
	}
}

// HandleSubmitAnswers handles submitting an answer for a field
func HandleSubmitAnswers(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var req models.AnswerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid request body. Required: field, answer",
			})
			return
		}

		// Check if session exists
		sess, err := store.Get(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "session_not_found",
				Message: "Session not found.",
			})
			return
		}

		// Validate field exists
		fieldExists := false
		for _, f := range sess.Fields {
			if f == req.Field {
				fieldExists = true
				break
			}
		}

		if !fieldExists {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_field",
				Message: "Field '" + req.Field + "' does not exist in this document.",
			})
			return
		}

		// Update session with answer
		err = store.Update(sessionID, func(s *models.Session) {
			s.Answers[req.Field] = req.Answer
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "update_failed",
				Message: "Failed to save answer.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Answer saved successfully.",
			"field":    req.Field,
			"progress": len(sess.Answers) + 1,
			"total":    len(sess.Fields),
		})
	}
}

// HandleGetNextQuestion returns the next unanswered question
func HandleGetNextQuestion(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		sess, err := store.Get(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "session_not_found",
				Message: "Session not found.",
			})
			return
		}

		// Find first unanswered field
		for _, field := range sess.Fields {
			if _, answered := sess.Answers[field]; !answered {
				// Check if we have an AI-phrased question
				question, hasAIQuestion := sess.Questions[field]
				if !hasAIQuestion {
					// Generate a simple humanized question as fallback
					question = humanizeFieldName(field)
				}

				c.JSON(http.StatusOK, models.QuestionResponse{
					Field:       field,
					Question:    question,
					IsAIPhrased: hasAIQuestion,
					Progress:    len(sess.Answers),
					Total:       len(sess.Fields),
					Done:        false,
				})
				return
			}
		}

		// All questions answered
		c.JSON(http.StatusOK, models.QuestionResponse{
			Done:     true,
			Progress: len(sess.Answers),
			Total:    len(sess.Fields),
		})
	}
}

// humanizeFieldName converts snake_case to human-readable question
func humanizeFieldName(field string) string {
	// Replace underscores with spaces
	words := strings.Split(field, "_")
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return "What is the " + strings.Join(words, " ") + "?"
}
