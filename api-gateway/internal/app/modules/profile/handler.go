package profile

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

func (h *Handler) GetMe(ctx *gin.Context) {
	resp, err := h.service.GetMe(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) Store(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "invalid request body")
		return
	}

	resp, err := h.service.Store(
		ctx.Request.Context(),
		body,
		ctx.GetHeader("Content-Type"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) GetPhoto(ctx *gin.Context) {
	resp, err := h.service.GetPhoto(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) StorePhoto(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "invalid request body")
		return
	}

	resp, err := h.service.StorePhoto(
		ctx.Request.Context(),
		body,
		ctx.GetHeader("Content-Type"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) ListInterests(ctx *gin.Context) {
	resp, err := h.service.ListInterests(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) LookupLocation(ctx *gin.Context) {
	resp, err := h.service.LookupLocation(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
		ctx.Request.URL.Query(),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}