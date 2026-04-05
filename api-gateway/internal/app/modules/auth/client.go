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
	CreateUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	UpdateUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	DeleteUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
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
	return &client{proxy: proxy.NewClient(baseURL, internalAPIKey, timeout)}
}

func (c *client) Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/login", body, contentType, headers)
}

func (c *client) Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/register", body, contentType, headers)
}

func (c *client) Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/refresh", body, contentType, headers)
}

func (c *client) Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/check", body, contentType, headers)
}

func (c *client) Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/logout", body, contentType, headers)
}

func (c *client) CreateUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.do(ctx, http.MethodPost, "/user/store", body, contentType, headers)
}

func (c *client) UpdateUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.do(ctx, http.MethodPut, "/user/update", body, contentType, headers)
}

func (c *client) DeleteUser(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.do(ctx, http.MethodDelete, "/user/delete", body, contentType, headers)
}

func (c *client) ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/password/forgot", body, contentType, headers)
}

func (c *client) ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/password/reset", body, contentType, headers)
}

func (c *client) ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/password/change", body, contentType, headers)
}

func (c *client) VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.post(ctx, "/password/verify-pin", body, contentType, headers)
}

func (c *client) post(ctx context.Context, path string, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.do(ctx, http.MethodPost, path, body, contentType, headers)
}

func (c *client) do(ctx context.Context, method, path string, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      method,
		Path:        path,
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func cloneQuery(values url.Values) url.Values { return proxy.CloneQuery(values) }