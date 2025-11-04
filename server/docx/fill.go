package docx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// FillDocument replaces placeholders with answers in the document using AI-powered smart replacement
func FillDocument(docBytes []byte, answers map[string]string) ([]byte, error) {
	// Write bytes to temp file (nguyenthenguyen/docx needs a file path)
	tmpFile, err := os.CreateTemp("", "docx-*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, bytes.NewReader(docBytes)); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close() // Close before reading

	// Read docx file
	doc, err := docx.ReadDocxFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read docx: %w", err)
	}
	defer doc.Close()

	// Get editable document
	editable := doc.Editable()

	// Get document content for smart replacement
	docText := editable.GetContent()

	// Use AI to create smart placeholder mappings
	placeholderMap, err := createSmartPlaceholderMap(docText, answers)
	if err != nil {
		// Fallback to simple replacement if AI fails
		fmt.Printf("AI replacement failed, using simple replacement: %v\n", err)
		placeholderMap = createSimplePlaceholderMap(answers)
	}

	// Replace placeholders using the mapping
	for placeholder, answer := range placeholderMap {
		editable.Replace(placeholder, answer, -1)
	}

	// Write the modified document to a new temp file
	outputFile, err := os.CreateTemp("", "docx-filled-*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	outputFileName := outputFile.Name()
	outputFile.Close()

	// Save the edited document
	if err := editable.WriteToFile(outputFileName); err != nil {
		return nil, fmt.Errorf("failed to write filled document: %w", err)
	}

	// Read the filled document back as bytes
	filledBytes, err := os.ReadFile(outputFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read filled document: %w", err)
	}

	return filledBytes, nil
}

// createSimplePlaceholderMap creates basic placeholder variations for each field
func createSimplePlaceholderMap(answers map[string]string) map[string]string {
	placeholders := make(map[string]string)

	for field, answer := range answers {
		// Format 1: {{field_name}}
		placeholders["{{"+field+"}}"] = answer

		// Format 2: [field_name] (lowercase with underscores)
		placeholders["["+field+"]"] = answer

		// Format 3: [Field Name] (title case with spaces)
		titleCase := strings.ReplaceAll(field, "_", " ")
		titleCase = strings.Title(titleCase)
		placeholders["["+titleCase+"]"] = answer

		// Format 4: [lowercase] without underscores
		lowercase := strings.ReplaceAll(field, "_", "")
		placeholders["["+lowercase+"]"] = answer

		// Format 5: $[_____] - all underscore blanks (will replace ALL of them with this value)
		// This is risky, so we'll skip it in simple mode
	}

	return placeholders
}

// createSmartPlaceholderMap uses AI to map field names to exact placeholder strings in the document
func createSmartPlaceholderMap(docText string, answers map[string]string) (map[string]string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	// Build list of fields to find
	fields := make([]string, 0, len(answers))
	for field := range answers {
		fields = append(fields, field)
	}

	// Use AI to find exact placeholders
	mapping, err := findPlaceholdersWithAI(docText, fields, apiKey)
	if err != nil {
		return nil, err
	}

	// Create final mapping with answers
	result := make(map[string]string)
	for field, placeholder := range mapping {
		if answer, ok := answers[field]; ok {
			result[placeholder] = answer
		}
	}

	return result, nil
}

// findPlaceholdersWithAI uses AI to find the exact placeholder text for each field
func findPlaceholdersWithAI(docText string, fields []string, apiKey string) (map[string]string, error) {
	// Truncate if needed
	maxLength := 10000
	if len(docText) > maxLength {
		docText = docText[:maxLength] + "... [truncated]"
	}

	fieldsJSON, _ := json.Marshal(fields)
	prompt := fmt.Sprintf(`Given this document text and a list of field names, find the EXACT placeholder text in the document that should be replaced for each field.

Fields to find: %s

Document text:
%s

For each field, identify the exact placeholder text as it appears in the document. This could be:
- [Field Name] format
- {{field_name}} format
- $[___________] (underscore blanks)
- Any other placeholder format

Return a JSON object mapping each field name to its exact placeholder text. For example:
{
  "company_name": "[COMPANY]",
  "investor_name": "[Investor Name]",
  "purchase_amount": "$[_____________]"
}

Important: Return the EXACT text as it appears in the document, including brackets, dollar signs, underscores, etc.

Do not include any explanation, just the JSON object.`, fieldsJSON, docText)

	// Make Gemini request
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
					{Text: "You are an expert at analyzing documents and finding placeholders. Always respond with valid JSON only.\n\n" + prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if isGeminiQuotaError(resp.StatusCode, body) {
			return nil, ErrGeminiQuotaExhausted
		}
		return nil, fmt.Errorf("Gemini API error: %s", string(body))
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
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Get the content and strip markdown code blocks if present
	content := geminiResp.Candidates[0].Content.Parts[0].Text
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

	// Parse the mapping
	var mapping map[string]string
	if err := json.Unmarshal([]byte(content), &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping: %w", err)
	}

	return mapping, nil
}
