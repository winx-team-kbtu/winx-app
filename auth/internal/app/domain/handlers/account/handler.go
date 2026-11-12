package account

import (
	"errors"
	"net/http"

	headercontract "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/internal/app/core/validation"
	service "auth/internal/app/domain/services/deletion"

	"github.com/gin-gonic/gin"
)

type DeleteDTO struct {
	Password string `json:"password" validate:"required,max=255"`
}

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

// Delete removes the authenticated user's account after password confirmation.
func (h *Handler) Delete(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	payload, ok := validation.BindAndValidate[DeleteDTO](h.binder, ctx)
	if !ok {
		return
	}

	if err := h.service.Delete(ctx.Request.Context(), authUser.ID, payload.Password); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			ctx.JSON(http.StatusForbidden, response.ErrorResponse(service.ErrInvalidPassword.Error()))
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(service.ErrNotFound.Error()))
		case errors.Is(err, service.ErrPending):
			// The account is already blocked, but cleanup did not finish. Not a success.
			ctx.JSON(http.StatusServiceUnavailable, response.ErrorResponse(service.ErrPending.Error()))
		default:
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "DeleteAccount service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.Deleted))
}
