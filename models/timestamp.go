/*
 * MarineNP Timestamp Utilities
 * Purpose: Custom time types and utilities for database models
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides custom time types and helpers for handling timestamps
 * in the MarineNP database models, ensuring compatibility with SQLite and JSON.
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// SQLiteTime is a custom type to handle SQLite timestamps
type SQLiteTime time.Time

// Value implements the driver.Valuer interface
func (t SQLiteTime) Value() (driver.Value, error) {
	return time.Time(t).Unix(), nil
}

// Scan implements the sql.Scanner interface
func (t *SQLiteTime) Scan(value interface{}) error {
	if value == nil {
		*t = SQLiteTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case int64:
		// Handle potential invalid timestamps
		if v < 0 || v > 253402300799 { // Max valid timestamp (9999-12-31 23:59:59)
			*t = SQLiteTime(time.Time{})
			return nil
		}
		*t = SQLiteTime(time.Unix(v, 0))
		return nil
	case time.Time:
		*t = SQLiteTime(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into SQLiteTime", value)
	}
}

// Time returns the time.Time value
func (t SQLiteTime) Time() time.Time {
	return time.Time(t)
}

// MarshalJSON implements the json.Marshaler interface
func (t SQLiteTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(tt.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *SQLiteTime) UnmarshalJSON(data []byte) error {
	var tm time.Time
	if err := json.Unmarshal(data, &tm); err != nil {
		return err
	}
	*t = SQLiteTime(tm)
	return nil
} 