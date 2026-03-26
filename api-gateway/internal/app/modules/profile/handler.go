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

func (h *Handler) Get(ctx *gin.Context) {
	resp, err := h.service.Get(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) Update(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "failed to read request body")
		return
	}

	resp, err := h.service.Update(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
		body,
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}

func (h *Handler) UpdateRole(ctx *gin.Context) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "failed to read request body")
		return
	}

	resp, err := h.service.UpdateRole(
		ctx.Request.Context(),
		transport.ForwardHeaders(ctx, "Authorization"),
		body,
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach profile service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}
