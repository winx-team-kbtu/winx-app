package role

import (
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/internal/app/core/validation"
	reqDto "auth/internal/app/domain/core/dto/requests/role"
	serviceDto "auth/internal/app/domain/core/dto/services/role"
	roleResource "auth/internal/app/domain/resources/role"
	service "auth/internal/app/domain/services/role"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound:  http.StatusNotFound,
	service.ErrDuplicate: http.StatusConflict,
}

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

func (h *Handler) Store(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[reqDto.StoreDTO](h.binder, ctx)
	if !ok {
		return
	}

	role, err := h.service.Store(ctx.Request.Context(), serviceDto.StoreDTO{
		Name: payload.Name,
		Slug: payload.Slug,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Store role service error")
		return
	}

	ctx.JSON(http.StatusCreated, response.SuccessResponse(roleResource.NewResource(role), response.Created))
}

func (h *Handler) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid role id"))
		return
	}

	payload, ok := validation.BindAndValidate[reqDto.UpdateDTO](h.binder, ctx)
	if !ok {
		return
	}

	role, err := h.service.Update(ctx.Request.Context(), serviceDto.UpdateDTO{
		ID:   id,
		Name: payload.Name,
		Slug: payload.Slug,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Update role service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(roleResource.NewResource(role), response.Updated))
}

func (h *Handler) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid role id"))
		return
	}

	ok, err := h.service.Delete(ctx.Request.Context(), id)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Delete role service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(ok, response.Deleted))
}
