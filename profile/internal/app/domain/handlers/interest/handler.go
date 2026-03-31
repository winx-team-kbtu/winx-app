package interest

import (
	"net/http"

	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	resources "winx-profile/internal/app/domain/resources/interest"
	service "winx-profile/internal/app/domain/services/interest"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(ctx *gin.Context) {
	items, err := h.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "ListInterests service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.OK))
}
