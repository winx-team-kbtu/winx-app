package erasure

import (
	"net/http"
	"strconv"

	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	service "winx-profile/internal/app/domain/services/erasure"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// DeleteUser is an internal, service-to-service endpoint protected by the API key.
// It is not exposed through the gateway.
func (h *Handler) DeleteUser(ctx *gin.Context) {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid user id"))

		return
	}

	if err := h.service.DeleteUserData(ctx.Request.Context(), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "DeleteUserData service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.Deleted))
}
