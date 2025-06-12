/*
 * MarineNP JSON Utilities
 * Purpose: Custom JSON marshaling and utility functions for API responses
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides custom JSON marshaling logic and helpers to ensure
 * consistent API responses for the MarineNP project.
 */

package models

import (
	"encoding/json"
	"strings"
)

// CustomJSONMarshaler is a custom JSON marshaler that handles SQLite string escaping
type CustomJSONMarshaler struct {
	Value interface{}
}

// MarshalJSON implements the json.Marshaler interface
func (m CustomJSONMarshaler) MarshalJSON() ([]byte, error) {
	// First marshal the value to JSON
	data, err := json.Marshal(m.Value)
	if err != nil {
		return nil, err
	}

	// Convert to string and unescape SQLite strings
	str := string(data)
	str = strings.ReplaceAll(str, "\\\\", "\\")

	// Convert back to bytes
	return []byte(str), nil
}

// MarshalToJSON marshals a value to JSON with SQLite string unescaping
func MarshalToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(CustomJSONMarshaler{Value: v})
} 