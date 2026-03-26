package profile

import (
	"context"
	"net/http"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	Get(ctx context.Context, headers map[string]string) (Response, error)
	Update(ctx context.Context, headers map[string]string, body []byte) (Response, error)
	UpdateRole(ctx context.Context, headers map[string]string, body []byte) (Response, error)
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

func (c *client) Get(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile",
		Headers: headers,
	})
}

func (c *client) Update(ctx context.Context, headers map[string]string, body []byte) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodPut,
		Path:    "/profile",
		Headers: headers,
		Body:    body,
	})
}

func (c *client) UpdateRole(ctx context.Context, headers map[string]string, body []byte) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodPut,
		Path:    "/admin/profile/role",
		Headers: headers,
		Body:    body,
	})
}
