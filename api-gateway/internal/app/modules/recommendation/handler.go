package recommendation

import (
	"winx-api-gateway/internal/app/transport"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(ctx *gin.Context) {
	resp, err := h.service.List(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
		ctx.Request.URL.Query(),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach recommendation service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}
