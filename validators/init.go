package validators

import (
	"alumni_api/customtypes"
	"github.com/go-playground/validator/v10"
)

func Init() {
	// Register the custom validation tag "encrypted"
	validate.RegisterValidation("encrypted", validateEncrypted)
}

func validateEncrypted(fl validator.FieldLevel) bool {
	encrypted, ok := fl.Field().Interface().(customtypes.Encrypted[any])
	if !ok {
		return false
	}

	// Validate the Value or other criteria
	return encrypted.Value != nil
}

var validate = validator.New()
