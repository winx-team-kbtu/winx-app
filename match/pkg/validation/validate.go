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
	initValidatorTagNames(validate)
	return &Validator{Validate: validate}, nil
}

func initValidatorTagNames(validate *validator.Validate) {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return ""
		}
		return name
	})
}
