package user

import (
	"auth/internal/app/core/validation"
	dto "auth/internal/app/domain/core/dto/services/user"
	validationservices "auth/internal/app/domain/validation-services"
	"auth/internal/app/models/models"
	"context"
	"fmt"
	"net/http"
)

type Validator struct {
	dbValidator validation.DBValidator
}

func New(dbValidator validation.DBValidator) *Validator {
	return &Validator{dbValidator: dbValidator}
}

func (v *Validator) CheckCreation(ctx context.Context, input dto.CreateDTO) *validationservices.Error {
	fErrs := validation.FieldErrors{}
	userModel := models.User{}

	unique, err := v.dbValidator.Unique(ctx, userModel.TableName(), map[string]any{
		userModel.EmailFieldName(): input.Email,
	}, nil)

	if err != nil {
		return &validationservices.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("failed checking email unique: %s", err),
		}
	}
	if !unique {
		fErrs[userModel.EmailFieldName()] = validationservices.UniqueMessage
	}

	if len(fErrs) > 0 {
		return &validationservices.Error{
			Code:    http.StatusUnprocessableEntity,
			Message: fErrs,
		}
	}

	return nil
}

func (v *Validator) CheckUpdate(ctx context.Context, id int64, input dto.UpdateDTO) *validationservices.Error {
	fErrs := validation.FieldErrors{}
	userModel := models.User{}

	unique, err := v.dbValidator.Unique(ctx, userModel.TableName(), map[string]any{
		userModel.EmailFieldName(): input.NewEmail,
	}, map[string]any{
		userModel.IDFieldName(): id,
	})

	if err != nil {
		return &validationservices.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("failed checking user email unique: %s", err),
		}
	}
	if !unique {
		fErrs[userModel.EmailFieldName()] = validationservices.UniqueMessage
	}

	if len(fErrs) > 0 {
		return &validationservices.Error{
			Code:    http.StatusUnprocessableEntity,
			Message: fErrs,
		}
	}

	return nil
}
