package match

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

func (h *Handler) SwipeLeft(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "invalid request body")
		return
	}
	resp, err := h.service.SwipeLeft(
		ctx.Request.Context(),
		body,
		ctx.GetHeader("Content-Type"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach match service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) SwipeRight(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "invalid request body")
		return
	}
	resp, err := h.service.SwipeRight(
		ctx.Request.Context(),
		body,
		ctx.GetHeader("Content-Type"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach match service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) List(ctx *gin.Context) {
	resp, err := h.service.List(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach match service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) Delete(ctx *gin.Context) {
	resp, err := h.service.Delete(
		ctx.Request.Context(),
		ctx.Param("id"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach match service")
		return
	}
	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}
