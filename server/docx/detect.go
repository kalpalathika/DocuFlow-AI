package docx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// ErrGeminiQuotaExhausted is returned when Gemini API quota is exhausted
var ErrGeminiQuotaExhausted = errors.New("gemini_quota_exhausted")

// isGeminiQuotaError checks if an error response indicates Gemini quota exhaustion
func isGeminiQuotaError(statusCode int, body []byte) bool {
	// Check for 429 status code (rate limit/quota exceeded)
	if statusCode == 429 {
		return true
	}

	// Check for common quota-related error messages in response body
	bodyStr := strings.ToLower(string(body))
	quotaKeywords := []string{
		"quota",
		"resource_exhausted",
		"rate limit",
		"rate_limit",
		"quota exceeded",
		"quota_exceeded",
	}

	for _, keyword := range quotaKeywords {
		if strings.Contains(bodyStr, keyword) {
			return true
		}
	}

	// Check for specific Gemini error structure
	var geminiError struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &geminiError); err == nil {
		errorMsg := strings.ToLower(geminiError.Error.Message)
		errorStatus := strings.ToLower(geminiError.Error.Status)
		for _, keyword := range quotaKeywords {
			if strings.Contains(errorMsg, keyword) || strings.Contains(errorStatus, keyword) {
				return true
			}
		}
		// Check for RESOURCE_EXHAUSTED status
		if errorStatus == "resource_exhausted" {
			return true
		}
	}

	return false
}

// OpenAI API structures for field detection
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

// DetectFields reads a .docx (bytes) and returns unique placeholders detected by AI
func DetectFields(docBytes []byte) ([]string, error) {
	// Write bytes to temp file (nguyenthenguyen/docx needs a file path)
	tmpFile, err := os.CreateTemp("", "docx-*.docx")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, bytes.NewReader(docBytes)); err != nil {
		return nil, err
	}
	tmpFile.Close() // Close before reading

	// Read docx file
	doc, err := docx.ReadDocxFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	// Get all text content
	docText := doc.Editable().GetContent()

	// Use AI to detect placeholders
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	fields, err := detectFieldsWithAI(docText, apiKey)
	if err != nil {
		return nil, fmt.Errorf("AI field detection failed: %w", err)
	}

	return fields, nil
}

// detectFieldsWithAI uses Gemini to intelligently detect dynamic placeholders
func detectFieldsWithAI(docText, apiKey string) ([]string, error) {
	prompt := buildDetectionPrompt(docText)

	// Prepare Gemini request
	type geminiContent struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}

	type geminiRequest struct {
		Contents []geminiContent `json:"contents"`
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: "You are an expert at analyzing legal documents and identifying dynamic placeholders that need to be filled in. You can distinguish between placeholders (like [Company Name], {{client_name}}, $[__________]) and static template text (like [Section 1(d)], [1]). Always respond with valid JSON only.\n\n" + prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request to Gemini
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if isGeminiQuotaError(resp.StatusCode, body) {
			return nil, ErrGeminiQuotaExhausted
		}
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse Gemini response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Parse the JSON array of field names from AI
	content := geminiResp.Candidates[0].Content.Parts[0].Text

	// Strip markdown code blocks if present (Gemini often wraps JSON in ```json ... ```)
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		// Remove opening code fence
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			lines = lines[1:] // Remove first line (```json or ```)
		}
		// Remove closing code fence
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		content = strings.Join(lines, "\n")
		content = strings.TrimSpace(content)
	}

	var fieldList []string
	if err := json.Unmarshal([]byte(content), &fieldList); err != nil {
		return nil, fmt.Errorf("failed to parse AI-detected fields: %w", err)
	}

	// Normalize field names to lowercase with underscores
	normalized := make([]string, 0, len(fieldList))
	for _, field := range fieldList {
		// Convert to lowercase and replace spaces with underscores
		normalized = append(normalized, normalizeFieldName(field))
	}

	// Remove duplicates and sort
	set := map[string]struct{}{}
	for _, field := range normalized {
		set[field] = struct{}{}
	}

	fields := make([]string, 0, len(set))
	for k := range set {
		fields = append(fields, k)
	}
	sort.Strings(fields)

	return fields, nil
}

// normalizeFieldName converts field names to consistent format
func normalizeFieldName(field string) string {
	// Remove common placeholder markers
	field = strings.Trim(field, "[]{}()$")
	field = strings.TrimSpace(field)

	// Convert to lowercase
	field = strings.ToLower(field)

	// Replace spaces with underscores
	field = strings.ReplaceAll(field, " ", "_")

	// Remove special characters except underscores
	var result strings.Builder
	for _, char := range field {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// buildDetectionPrompt creates the prompt for AI field detection
func buildDetectionPrompt(docText string) string {
	// Truncate document text if too long (to stay within token limits)
	maxLength := 10000
	if len(docText) > maxLength {
		docText = docText[:maxLength] + "... [truncated]"
	}

	return fmt.Sprintf(`Analyze the following legal document text and identify all DYNAMIC PLACEHOLDERS that need to be filled in with user data.

INCLUDE placeholders like:
- [Company Name], [Investor Name], [Date]
- {{client_name}}, {{contract_amount}}
- $[_____________] or $[__________] when they represent fields to be filled (look at nearby text for context)
- Any text that looks like a variable to be filled in

EXCLUDE:
- Section references like [Section 1(d)], [1], [a], [i]
- Footnote markers like [1], [2]
- Static text in brackets
- Legal citation references
- Page numbers

Important: For underscore blanks like $[__________], look at the surrounding text to determine what field they represent. For example:
- If you see "$[_____________] (the "Purchase Amount")", identify it as "Purchase Amount"
- If you see "by [Investor Name]", identify it as "Investor Name"

For underscore blanks, infer the field name from the context around them.

Document text:
%s

Return ONLY a JSON array of the dynamic placeholder field names you found (use descriptive names from context).
For example:
["Company Name", "Investor Name", "Date of Safe", "Purchase Amount", "Valuation Cap"]

Do not include any explanation, just the JSON array.`, docText)
}
