package handlers

import (
	headercontract "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/internal/app/core/validation"
	dto "auth/internal/app/domain/core/dto/requests"
	serviceDto "auth/internal/app/domain/core/dto/services"
	userServiceDto "auth/internal/app/domain/core/dto/services/user"
	"auth/internal/app/domain/resources"
	service "auth/internal/app/domain/services"
	uservalidationservice "auth/internal/app/domain/validation-services/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound:        http.StatusNotFound,
	service.ErrFailedLogin:     http.StatusUnauthorized,
	service.ErrUnauthenticated: http.StatusUnauthorized,
	service.ErrFailedCache:     http.StatusInternalServerError,
	service.ErrFailedPublish:   http.StatusInternalServerError,
}

type Handler struct {
	service           service.Service
	binder            *validation.Binder
	validationService *uservalidationservice.Validator
}

func NewHandler(service service.Service, binder *validation.Binder, vs *uservalidationservice.Validator) *Handler {
	return &Handler{service: service, binder: binder, validationService: vs}
}

func (h *Handler) Register(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.RegisterDTO](h.binder, ctx)
	if !ok {
		return
	}

	inputDto := userServiceDto.CreateDTO{
		Email:    payload.Email,
		Password: payload.Password,
	}

	vErr := h.validationService.CheckCreation(ctx, inputDto)
	if vErr != nil {
		ctx.JSON(vErr.Code, response.ValidationErrorResponse(vErr.Message))
		errorhandler.FailOnError(vErr, "validation error")

		return
	}

	user, err := h.service.Register(ctx, serviceDto.RegisterDTO{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "Register service error")

		return
	}

	userResource := resources.NewResource(user)
	ctx.JSON(http.StatusCreated, response.SuccessResponse(userResource, response.Created))
}

func (h *Handler) Login(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.LoginDTO](h.binder, ctx)
	if !ok {
		return
	}

	token, status, err := h.service.Login(ctx, serviceDto.LoginDTO{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "Login service error")

		return
	}

	ctx.JSON(status, response.SuccessResponse(token, response.OK))
}

func (h *Handler) RefreshToken(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.RefreshTokenDTO](h.binder, ctx)
	if !ok {
		return
	}

	token, err := h.service.RefreshToken(ctx, payload.RefreshToken)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "UserRefreshToken service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(token, response.OK))
}

func (h *Handler) CheckToken(ctx *gin.Context) {
	user, err := h.service.CheckToken(ctx, ctx.GetHeader("Authorization"))
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "CheckToken service error")

		return
	}

	userResource := resources.NewResource(user)
	ctx.JSON(http.StatusOK, response.SuccessResponse(userResource, response.OK))
}

func (h *Handler) Logout(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "failed Logout Auth handler")

		return
	}

	logout, err := h.service.Logout(ctx.Request.Context(), authUser.Email, ctx.GetHeader("Authorization"))
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "Logout service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(logout, response.OK))
}
