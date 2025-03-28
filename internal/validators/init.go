package validators

import (
	"alumni_api/pkg/customtypes"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Init() {
	// Register the custom validation tag "encrypted"
	validate.RegisterCustomTypeFunc(ValidateValuer, customtypes.Encrypted[float32]{}, customtypes.Encrypted[int16]{}, customtypes.Encrypted[int32]{}, customtypes.Encrypted[string]{})
	validate.RegisterValidation("encrypted", validateEncrypted)
	validate.RegisterValidation("encrypted_dynamic", validateEncryptedGeneric)
	validate.RegisterValidation("customname", nameValidation)
	validate.RegisterValidation("cpe_generation", CPEGenerationValidation)
	validate.RegisterValidation("phone", ValidatePhone)
}

func ValidateValuer(field reflect.Value) interface{} {

	if field.Kind() == reflect.Struct && strings.HasPrefix(field.Type().Name(), "Encrypted[") {
		valueField := field.FieldByName("Value")

		if !valueField.IsValid() {
			return nil
		}

		return valueField.Interface()
	}

	return nil
}
