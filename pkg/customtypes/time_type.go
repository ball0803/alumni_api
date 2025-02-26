package customtypes

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"time"
)

// CustomTime wraps time.Time for custom JSON parsing
type CustomTime struct {
	time.Time `mapstruct:"Time,squash"`
}

// Custom format expected: "YYYY-MM-DD"
const customTimeFormat = "2006-01-02"

// DecodeCustomTimeHook is a custom hook that handles CustomTime decoding
func DecodeCustomTimeHook(f reflect.Type) mapstructure.DecodeHookFunc {
	return func(from, to interface{}) error {
		// Check if we're decoding into a CustomTime
		if toCustomTime, ok := to.(*CustomTime); ok {
			if str, ok := from.(string); ok {
				// Parse the string into CustomTime
				t, err := time.Parse(customTimeFormat, str)
				if err != nil {
					return fmt.Errorf("invalid date format for CustomTime, expected YYYY-MM-DD: %v", err)
				}
				toCustomTime.Time = t
				return nil
			}
			return fmt.Errorf("invalid type for CustomTime, expected string, got %T", from)
		}
		return nil
	}
}

// UnmarshalJSON parses JSON date into time.Time
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Remove quotes from JSON string
	str := string(b)
	if len(str) < 2 {
		return errors.New("invalid date format")
	}
	str = str[1 : len(str)-1] // Trim quotes

	// Parse the date into time.Time
	t, err := time.Parse(customTimeFormat, str)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %v", err)
	}
	ct.Time = t
	return nil
}

// // MarshalJSON converts time.Time back to string format
// func (ct CustomTime) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(ct.Time.Format(customTimeFormat))
// }

// MarshalJSON converts CustomTime to a string representation
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	// Format time in RFC3339
	return []byte(fmt.Sprintf("\"%s\"", ct.Format(customTimeFormat))), nil
}
