package user

import (
	"winx-notification/internal/app/core/helpers/errorhandler"
	"winx-notification/internal/app/core/helpers/response"
	"winx-notification/internal/app/core/validation"
	dto "winx-notification/internal/app/domain/core/dto/requests/user"
	serviceDto "winx-notification/internal/app/domain/core/dto/services/user"
	"winx-notification/internal/app/domain/resources"
	service "winx-notification/internal/app/domain/services/user"
	validationservice "winx-notification/internal/app/domain/validation-services/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound: http.StatusNotFound,
}

type Handler struct {
	service           service.Service
	binder            *validation.Binder
	validationService *validationservice.Validator
}

func NewHandler(
	service service.Service,
	binder *validation.Binder,
	vs *validationservice.Validator,
) *Handler {
	return &Handler{service: service, binder: binder, validationService: vs}
}

func (h *Handler) Create(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.CreateDTO](h.binder, ctx)
	if !ok {
		return
	}

	inputDto := serviceDto.CreateDTO{
		Email:    payload.Email,
		Password: payload.Password,
	}

	vErr := h.validationService.CheckCreation(ctx, inputDto)
	if vErr != nil {
		ctx.JSON(vErr.Code, response.ValidationErrorResponse(vErr.Message))
		errorhandler.FailOnError(vErr, "validation error")

		return
	}

	user, err := h.service.Create(ctx, inputDto)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "CreateUser service error")

		return
	}

	userResource := resources.NewResource(user)
	ctx.JSON(http.StatusOK, response.SuccessResponse(userResource, response.OK))
}

func (h *Handler) Delete(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.DeleteDTO](h.binder, ctx)
	if !ok {
		return
	}

	ok, err := h.service.Delete(ctx, serviceDto.DeleteDTO{
		Email: payload.Email,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "DeleteUser service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(ok, response.Deleted))
}

func (h *Handler) Update(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[dto.UpdateDTO](h.binder, ctx)
	if !ok {
		return
	}

	inputDto := serviceDto.UpdateDTO{
		Email:    payload.Email,
		NewEmail: payload.NewEmail,
		Password: payload.Password,
	}

	user, err := h.service.GetByEmail(ctx, inputDto.Email)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "GetByEmail service error")

		return
	}

	vErr := h.validationService.CheckUpdate(ctx, user.ID, inputDto)
	if vErr != nil {
		ctx.JSON(vErr.Code, response.ValidationErrorResponse(vErr.Message))
		errorhandler.FailOnError(vErr, "validation error")

		return
	}

	user, err = h.service.Update(ctx, inputDto)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "UpdateUser service error")

		return
	}

	userResource := resources.NewResource(user)
	ctx.JSON(http.StatusOK, response.SuccessResponse(userResource, response.Updated))
}
