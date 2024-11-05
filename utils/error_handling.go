package utils

import (
	"fmt"
)

// WrapError to add context to errors
func WrapError(err error, context string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", context, err)
	}
	return nil
}
