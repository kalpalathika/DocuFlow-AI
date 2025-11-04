package models

import "time"

// Session represents a document filling session
type Session struct {
	ID          string            `json:"id"`
	OriginalDoc []byte            `json:"-"` // Raw DOCX bytes (not sent to client)
	Fields      []string          `json:"fields"`
	FieldTypes  map[string]string `json:"fieldTypes"` // field -> type (text, number, date)
	Answers     map[string]string `json:"answers"`
	Questions   map[string]string `json:"questions"` // AI-phrased questions (field -> question)
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// UploadResponse is returned after a successful document upload
type UploadResponse struct {
	SessionID string   `json:"sessionId"`
	Fields    []string `json:"fields"`
	Message   string   `json:"message"`
}

// QuestionResponse is returned when requesting the next question
type QuestionResponse struct {
	Field       string `json:"field"`
	FieldType   string `json:"fieldType"`   // Type: text, number, or date
	Question    string `json:"question"`
	IsAIPhrased bool   `json:"isAIPhrased"` // True if AI-generated, false if fallback
	Progress    int    `json:"progress"`    // Number of answered fields
	Total       int    `json:"total"`       // Total number of fields
	Done        bool   `json:"done"`        // True if all questions answered
}

// AnswerRequest is the request body for submitting answers
type AnswerRequest struct {
	Field  string `json:"field" binding:"required"`
	Answer string `json:"answer" binding:"required"`
}

// GenerateQuestionsResponse is returned after AI question generation
type GenerateQuestionsResponse struct {
	Questions map[string]string `json:"questions"` // field -> AI-phrased question
	Count     int               `json:"count"`
	Message   string            `json:"message"`
}

// SessionStatusResponse returns the current session status
type SessionStatusResponse struct {
	SessionID   string            `json:"sessionId"`
	Fields      []string          `json:"fields"`
	Answers     map[string]string `json:"answers"`
	Questions   map[string]string `json:"questions"`
	Progress    int               `json:"progress"`
	Total       int               `json:"total"`
	IsCompleted bool              `json:"isCompleted"`
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
