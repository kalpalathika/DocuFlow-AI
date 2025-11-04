package utils

import "strings"

// InferFieldType determines the input type based on the field name
// Returns: "text", "number", or "date"
func InferFieldType(fieldName string) string {
	lowerField := strings.ToLower(fieldName)

	// Date patterns
	datePatterns := []string{
		"date", "dob", "birth", "deadline",
		"expiry", "expiration", "anniversary",
	}
	for _, pattern := range datePatterns {
		if strings.Contains(lowerField, pattern) {
			return "date"
		}
	}

	// Number patterns
	numberPatterns := []string{
		"age", "count", "number", "amount",
		"quantity", "price", "total", "sum",
		"year", "months", "days", "hours",
	}
	for _, pattern := range numberPatterns {
		if strings.Contains(lowerField, pattern) {
			return "number"
		}
	}

	// Default to text
	return "text"
}
