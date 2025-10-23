package util

import (
	"strings"
	"time"
)

// FormatDate formats a date string from YYYY-MM-DD to "Month Day, Year"
func FormatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse the date
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr // Return original if parsing fails
	}

	return t.Format("January 2, 2006")
}

// ParseDate parses a date string in various formats to YYYY-MM-DD
func ParseDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Try different formats
	formats := []string{
		"January 2, 2006", // "January 15, 2025"
		"2006-01-02",      // "2025-01-15"
		time.RFC3339,      // ISO with time
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}

	return dateStr // Return original if no format matches
}

// Capitalize capitalizes the first letter of each word
func Capitalize(s string) string {
	return strings.Title(strings.ToLower(s)) // TODO: Replace with golang.org/x/text/cases
}
