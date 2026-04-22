package chat

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
		ctx.Request.URL.Query(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach chat service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) Messages(ctx *gin.Context) {
	resp, err := h.service.Messages(
		ctx.Request.Context(),
		ctx.Param("id"),
		ctx.Request.URL.Query(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach chat service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}
