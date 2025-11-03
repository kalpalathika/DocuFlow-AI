package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/you/lexsy-mvp/server/docx"
	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/session"
)

// HandleGenerateDocument generates the filled document for download
func HandleGenerateDocument(store *session.Store) gin.HandlerFunc {
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

		// Check if all fields have been answered
		unansweredFields := []string{}
		for _, field := range sess.Fields {
			if _, answered := sess.Answers[field]; !answered {
				unansweredFields = append(unansweredFields, field)
			}
		}

		if len(unansweredFields) > 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "incomplete_answers",
				Message: fmt.Sprintf("Not all fields have been answered. Missing: %v", unansweredFields),
			})
			return
		}

		// Fill the document with answers
		filledDoc, err := docx.FillDocument(sess.OriginalDoc, sess.Answers)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "document_generation_failed",
				Message: "Failed to generate document: " + err.Error(),
			})
			return
		}

		// Return the document as a downloadable file
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		c.Header("Content-Disposition", "attachment; filename=filled_document.docx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", filledDoc)
	}
}
