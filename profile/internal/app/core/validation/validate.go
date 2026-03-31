package validation

import (
	"errors"
	"net/http"
	"strings"

	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FieldErrors map[string]string

type Binder struct {
	v *pkgvalidation.Validator
}

func NewBinder(v *pkgvalidation.Validator) *Binder { return &Binder{v: v} }

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
	case "email":
		return "invalid email"
	case "url":
		return "invalid url"
	case "datetime":
		return "invalid datetime format"
	default:
		return fe.Error()
	}
}
