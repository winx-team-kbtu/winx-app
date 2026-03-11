package auth

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
}

type Response = proxy.Response

type client struct {
	proxy *proxy.Client
}

func NewClient(baseURL, internalAPIKey string, timeout time.Duration) Client {
	return &client{
		proxy: proxy.NewClient(baseURL, internalAPIKey, timeout),
	}
}

func (c *client) Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/login", body, contentType, headers)
}

func (c *client) Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/register", body, contentType, headers)
}

func (c *client) Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/refresh", body, contentType, headers)
}

func (c *client) Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/check", body, contentType, headers)
}

func (c *client) Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/logout", body, contentType, headers)
}

func (c *client) ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/password/forgot", body, contentType, headers)
}

func (c *client) ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/password/reset", body, contentType, headers)
}

func (c *client) ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/password/change", body, contentType, headers)
}

func (c *client) VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.doPost(ctx, "/password/verify-pin", body, contentType, headers)
}

func (c *client) doPost(ctx context.Context, path string, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      http.MethodPost,
		Path:        path,
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func cloneQuery(values url.Values) url.Values {
	out := make(url.Values, len(values))
	for key, vals := range values {
		copied := make([]string, len(vals))
		copy(copied, vals)
		out[key] = copied
	}
	return out
}
