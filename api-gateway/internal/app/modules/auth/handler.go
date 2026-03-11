package auth

import (
	"context"
	"winx-api-gateway/internal/app/transport"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.Login)
}

func (h *Handler) Register(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.Register)
}

func (h *Handler) Refresh(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.Refresh)
}

func (h *Handler) Check(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.Check)
}

func (h *Handler) Logout(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.Logout)
}

func (h *Handler) ForgotPassword(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.ForgotPassword)
}

func (h *Handler) ResetPassword(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.ResetPassword)
}

func (h *Handler) ChangePassword(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.ChangePassword)
}

func (h *Handler) VerifyPin(ctx *gin.Context) {
	h.proxyPost(ctx, h.service.VerifyPin)
}

func (h *Handler) proxyPost(
	ctx *gin.Context,
	call func(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error),
) {
	body, err := transport.ReadBody(ctx)
	if err != nil {
		transport.WriteJSONError(ctx, 400, "invalid request body")
		return
	}

	resp, err := call(
		ctx.Request.Context(),
		body,
		ctx.GetHeader("Content-Type"),
		transport.ForwardHeaders(ctx, "Authorization"),
	)
	if err != nil {
		transport.WriteJSONError(ctx, 502, "failed to reach auth service")
		return
	}

	transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
}
