/*
 * MarineNP SQLite Utilities
 * Purpose: Helper functions for SQLite database operations
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides utility functions for working with SQLite databases
 * in the MarineNP project, including migrations and connection helpers.
 */

package utils

import (
	"strings"
)

// UnescapeSQLiteString removes the extra backslashes that SQLite adds to escaped characters
func UnescapeSQLiteString(s string) string {
	// Replace double backslashes with single backslashes
	return strings.ReplaceAll(s, "\\\\", "\\")
}

// EscapeSQLiteString adds the necessary escaping for SQLite string literals
func EscapeSQLiteString(s string) string {
	// Replace single backslashes with double backslashes
	return strings.ReplaceAll(s, "\\", "\\\\")
}

// UnescapeSQLiteStrings handles a slice of strings
func UnescapeSQLiteStrings(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = UnescapeSQLiteString(s)
	}
	return result
}

// UnescapeSQLiteMap handles a map of strings
func UnescapeSQLiteMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = UnescapeSQLiteString(v)
	}
	return result
}

// UnescapeSQLiteInterface handles interface{} that might contain strings
func UnescapeSQLiteInterface(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return UnescapeSQLiteString(val)
	case []string:
		return UnescapeSQLiteStrings(val)
	case map[string]string:
		return UnescapeSQLiteMap(val)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = UnescapeSQLiteInterface(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = UnescapeSQLiteInterface(v)
		}
		return result
	default:
		return v
	}
} 