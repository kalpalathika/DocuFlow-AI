package docx

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/nguyenthenguyen/docx"
)

// FillDocument replaces placeholders with answers in the document
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

	// Replace each placeholder with its answer
	for field, answer := range answers {
		placeholder := "{{" + field + "}}"
		editable.Replace(placeholder, answer, -1) // -1 means replace all occurrences
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
