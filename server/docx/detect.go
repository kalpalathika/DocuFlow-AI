package docx

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/nguyenthenguyen/docx"
)

var rxVar = regexp.MustCompile(`\{\{[a-zA-Z0-9_.]+\}\}`)

// DetectFields reads a .docx (bytes) and returns unique placeholders without {{ }}
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

	// Find all placeholders
	set := map[string]struct{}{}
	matches := rxVar.FindAllString(docText, -1)
	for _, v := range matches {
		// v is like {{client_name}} → strip braces
		key := v[2 : len(v)-2]
		set[key] = struct{}{}
	}

	// Convert to sorted slice
	fields := make([]string, 0, len(set))
	for k := range set {
		fields = append(fields, k)
	}
	sort.Strings(fields)

	return fields, nil
}
