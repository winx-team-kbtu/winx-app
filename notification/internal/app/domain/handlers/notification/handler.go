package notification

import (
	"net/http"
	"strconv"

	headercontract "winx-notification/internal/app/core/contracts/microservices/header-contract"
	"winx-notification/internal/app/core/helpers/errorhandler"
	"winx-notification/internal/app/core/helpers/response"
	resources "winx-notification/internal/app/domain/resources/notification"
	service "winx-notification/internal/app/domain/services/notification"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound: http.StatusNotFound,
}

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	items, err := h.service.ListByRecipient(ctx.Request.Context(), authUser.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "ListNotifications service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.OK))
}

func (h *Handler) Delete(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid notification id"))
		errorhandler.FailOnError(err, "invalid notification id")
		return
	}

	ok, err := h.service.DeleteByIDAndRecipient(ctx.Request.Context(), id, authUser.Email)
	if err != nil {
		if code, found := errStatusMap[err]; found {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}

		errorhandler.FailOnError(err, "DeleteNotification service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(ok, response.Deleted))
}
