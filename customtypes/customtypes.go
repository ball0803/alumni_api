package customtypes

import (
	"alumni_api/encrypt"
	"encoding/json"
	"fmt"
	"reflect"
)

type Encrypted[T any] struct {
	Raw   []byte
	Value T
}

func (e *Encrypted[T]) UnmarshalJSON(data []byte) error {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("invalid value for type %T: %v", e.Value, err)
	}
	e.Value = value
	return nil
}

func (e Encrypted[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Value)
}

func (e *Encrypted[T]) Encrypt() error {
	if reflect.ValueOf(e.Value).IsZero() {
		return fmt.Errorf("Value field is not set")
	}
	// Encrypt the Value field with a header
	raw, err := encrypt.AESEncryptWithHeader(reflect.ValueOf(e))
	if err != nil {
		return err
	}

	// Set the Raw field with the encrypted data
	e.Raw = raw.Interface().([]byte)
	return nil
}

func (e *Encrypted[T]) Decrypt() error {
	// Ensure the Raw field is set
	if len(e.Raw) == 0 {
		return fmt.Errorf("Raw field is not set")
	}
	// Decrypt the Raw field and get the original value
	decrypted, err := encrypt.AESDecryptWithHeader(reflect.ValueOf(e))
	if err != nil {
		return err
	}

	// Use a type assertion to ensure type safety for the Value field
	if value, ok := decrypted.Interface().(T); ok {
		e.Value = value
	} else {
		return fmt.Errorf("type mismatch: expected %T, got %T", e.Value, decrypted.Interface())
	}
	return nil
}
