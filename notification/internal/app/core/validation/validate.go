package validation

import (
	"winx-notification/internal/app/core/helpers/errorhandler"
	"winx-notification/internal/app/core/helpers/response"
	"winx-notification/pkg/validation"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FieldErrors map[string]string

type Binder struct {
	v *validation.Validator
}

func NewBinder(v *validation.Validator) *Binder { return &Binder{v: v} }

func BindAndValidate[T any](b *Binder, c *gin.Context) (T, bool) {
	var req T
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(err.Error()))
		errorhandler.FailOnError(err, "validation error")

		return req, false
	}

	if err := b.v.Validate.Struct(req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			out := FieldErrors{}
			for _, fe := range verrs {
				path := fe.Namespace()
				if i := strings.IndexByte(path, '.'); i >= 0 {
					path = path[i+1:]
				}

				out[path] = human(fe)
			}

			c.JSON(http.StatusUnprocessableEntity, response.ValidationErrorResponse(out))
			errorhandler.FailOnError(err, "validation error")

			return req, false
		}

		c.JSON(http.StatusUnprocessableEntity, response.ValidationErrorResponse(err.Error()))
		errorhandler.FailOnError(err, "validation error")

		return req, false
	}

	return req, true
}

func human(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "len":
		return "invalid length: " + fe.Param()
	case "email":
		return "invalid email"
	case "numeric":
		return "must be numeric"
	case "password":
		return "Password must contain at least one uppercase letter, one lowercase letter, one digit, one special character, and be at least 8 characters long"
	default:
		return fe.Error()
	}
}
