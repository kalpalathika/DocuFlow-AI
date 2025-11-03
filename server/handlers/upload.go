package handlers

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/you/lexsy-mvp/server/docx"
	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/session"
)

// HandleUpload processes document upload and creates a new session
func HandleUpload(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get file from multipart form (try common field names)
		var file *multipart.FileHeader
		var err error

		// Try "document" first
		file, err = c.FormFile("document")
		if err != nil {
			// Try "file" as fallback
			file, err = c.FormFile("file")
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "missing_file",
				Message: "No document file uploaded. Please upload a .docx file with field name 'document' or 'file'.",
			})
			return
		}

		// Validate file type by extension (more reliable than Content-Type)
		if !strings.HasSuffix(strings.ToLower(file.Filename), ".docx") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_file_type",
				Message: "Only .docx files are supported.",
			})
			return
		}

		// Open and read file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "file_read_error",
				Message: "Failed to read uploaded file.",
			})
			return
		}
		defer src.Close()

		// Read file bytes
		docBytes, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "file_read_error",
				Message: "Failed to read file contents.",
			})
			return
		}

		// Detect placeholders in document
		fields, err := docx.DetectFields(docBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "field_detection_error",
				Message: "Failed to detect fields in document. Error: " + err.Error(),
			})
			return
		}

		// Check if any fields were found
		if len(fields) == 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "no_fields_found",
				Message: "No placeholders found in document. Use {{field_name}} format for placeholders.",
			})
			return
		}

		// Create session
		sess, err := store.Create(docBytes, fields)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "session_creation_error",
				Message: "Failed to create session.",
			})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, models.UploadResponse{
			SessionID: sess.ID,
			Fields:    fields,
			Message:   "Document uploaded successfully.",
		})
	}
}
