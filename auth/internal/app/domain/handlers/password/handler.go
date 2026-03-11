package password

import (
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/internal/app/core/validation"
	dto "auth/internal/app/domain/core/dto/requests/password"
	serviceDto "auth/internal/app/domain/core/dto/services/password"
	resources "auth/internal/app/domain/resources/password"
	service "auth/internal/app/domain/services/password"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrUnauthenticated: http.StatusUnauthorized,
	service.ErrNotFound:        http.StatusNotFound,
	service.ErrInvalidToken:    http.StatusUnauthorized,
	service.ErrUserNotFound:    http.StatusNotFound,
	service.ErrTokenExpired:    http.StatusUnauthorized,
	service.ErrInvalidUser:     http.StatusUnauthorized,
	service.ErrInvalidPassword: http.StatusUnprocessableEntity,
	service.ErrInvalidPinCode:  http.StatusUnprocessableEntity,
	service.ErrFailedPublish:   http.StatusInternalServerError,
}

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

func (h *Handler) ForgotPassword(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.ForgotPasswordDTO](h.binder, ctx)
	if !ok {
		return
	}

	err := h.service.ForgotPassword(ctx, serviceDto.ForgotPasswordDTO{
		Email: payload.Email,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "ForgotPassword service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.SentResetToken))
}

func (h *Handler) ResetPassword(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.ResetPasswordDTO](h.binder, ctx)
	if !ok {
		return
	}

	err := h.service.ResetPassword(ctx, serviceDto.ResetPasswordDTO{
		Email:                   payload.Email,
		Token:                   payload.Token,
		NewPassword:             payload.NewPassword,
		NewPasswordConfirmation: payload.NewPasswordConfirmation,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "ResetPassword service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.Updated))
}

func (h *Handler) ChangePassword(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.ChangePasswordDTO](h.binder, ctx)
	if !ok {
		return
	}
	token := ctx.GetHeader("Authorization")

	err := h.service.ChangePassword(ctx, serviceDto.ChangePasswordDTO{
		Password:        payload.Password,
		NewPassword:     payload.NewPassword,
		ConfirmPassword: payload.ConfirmPassword,
		Token:           token,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "ChangePassword service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.OK))
}

func (h *Handler) VerifyPin(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.VerifyPinDTO](h.binder, ctx)
	if !ok {
		return
	}

	user, err := h.service.VerifyPin(ctx, serviceDto.VerifyPinDTO{
		Email:   payload.Email,
		PinCode: payload.PinCode,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "VerifyPin service error")

		return
	}

	tokenResource := resources.NewResource(user)
	ctx.JSON(http.StatusOK, response.SuccessResponse(tokenResource, response.OK))
}
