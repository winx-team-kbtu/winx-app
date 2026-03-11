package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validate *validator.Validate
}

func New() (*Validator, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := registerValidations(validate); err != nil {
		return nil, err
	}

	initValidatorTagNames(validate)

	return &Validator{Validate: validate}, nil
}

func registerValidations(validator *validator.Validate) error {
	err := validator.RegisterValidation("password", password)
	if err != nil {
		return err
	}

	return nil
}

func initValidatorTagNames(validator *validator.Validate) {
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return ""
		}
		return name
	})
}
