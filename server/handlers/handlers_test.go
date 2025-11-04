package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/session"
)

// setupTestRouter creates a test router with the same routes as the main application
func setupTestRouter() (*gin.Engine, *session.Store) {
	gin.SetMode(gin.TestMode)
	store := session.NewStore()
	
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	
	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// API routes
	api := r.Group("/api")
	{
		api.POST("/upload", HandleUpload(store))
		api.GET("/session/:id", HandleGetSession(store))
		api.POST("/session/:id/answers", HandleSubmitAnswers(store))
		api.GET("/session/:id/next", HandleGetNextQuestion(store))
		api.POST("/session/:id/generate", HandleGenerateDocument(store))
	}
	
	return r, store
}

// createTestDocx creates minimal test DOCX bytes for testing
// Note: This is a minimal valid DOCX structure - actual file content doesn't matter
// for session workflow tests since we're not calling DetectFields
func createTestDocx() ([]byte, error) {
	// For integration tests, we just need some bytes to pass to store.Create()
	// We're not actually parsing the DOCX in the workflow test
	// So we can use a simple placeholder - in a real scenario, this would be actual DOCX bytes
	// For now, return a minimal byte array that satisfies the type requirement
	return []byte("mock docx bytes"), nil
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter()
	
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestUploadInvalidFileType tests upload endpoint with invalid file type
func TestUploadInvalidFileType(t *testing.T) {
	router, _ := setupTestRouter()
	
	// Create a multipart form with a text file instead of DOCX
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("document", "test.txt")
	require.NoError(t, err)
	part.Write([]byte("This is not a DOCX file"))
	writer.Close()
	
	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid_file_type", response.Error)
	assert.Contains(t, response.Message, ".docx")
}

// TestSessionWorkflow tests the complete session workflow
func TestSessionWorkflow(t *testing.T) {
	router, store := setupTestRouter()
	
	// Create a session directly (bypassing upload to avoid Gemini API dependency)
	testDocx, err := createTestDocx()
	require.NoError(t, err)
	
	testFields := []string{"test_field", "another_field"}
	sess, err := store.Create(testDocx, testFields)
	require.NoError(t, err)
	sessionID := sess.ID
	
	// Test 1: Get session
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/session/%s", sessionID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var sessionResponse models.SessionStatusResponse
	err = json.Unmarshal(w.Body.Bytes(), &sessionResponse)
	require.NoError(t, err)
	assert.Equal(t, sessionID, sessionResponse.SessionID)
	assert.Equal(t, 2, len(sessionResponse.Fields))
	assert.Equal(t, 0, sessionResponse.Progress)
	assert.Equal(t, 2, sessionResponse.Total)
	
	// Test 2: Get next question
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/session/%s/next", sessionID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var questionResponse models.QuestionResponse
	err = json.Unmarshal(w.Body.Bytes(), &questionResponse)
	require.NoError(t, err)
	assert.False(t, questionResponse.Done)
	assert.Equal(t, "test_field", questionResponse.Field)
	assert.Equal(t, 0, questionResponse.Progress)
	assert.Equal(t, 2, questionResponse.Total)
	
	// Test 3: Submit answer
	answerReq := models.AnswerRequest{
		Field:  "test_field",
		Answer: "test answer",
	}
	body, err := json.Marshal(answerReq)
	require.NoError(t, err)
	
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/session/%s/answers", sessionID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var answerResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &answerResponse)
	require.NoError(t, err)
	assert.Equal(t, "Answer saved successfully.", answerResponse["message"])
	assert.Equal(t, "test_field", answerResponse["field"])
	
	// Test 4: Get next question again (should return second field)
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/session/%s/next", sessionID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &questionResponse)
	require.NoError(t, err)
	assert.False(t, questionResponse.Done)
	assert.Equal(t, "another_field", questionResponse.Field)
	assert.Equal(t, 1, questionResponse.Progress)
	
	// Test 5: Submit second answer
	answerReq = models.AnswerRequest{
		Field:  "another_field",
		Answer: "another answer",
	}
	body, err = json.Marshal(answerReq)
	require.NoError(t, err)
	
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/session/%s/answers", sessionID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test 6: Get next question (should be done)
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/session/%s/next", sessionID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &questionResponse)
	require.NoError(t, err)
	assert.True(t, questionResponse.Done)
	assert.Equal(t, 2, questionResponse.Progress)
	assert.Equal(t, 2, questionResponse.Total)
}
