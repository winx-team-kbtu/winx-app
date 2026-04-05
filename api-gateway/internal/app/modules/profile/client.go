package profile

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	GetMe(ctx context.Context, headers map[string]string) (Response, error)
	Store(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	GetPhoto(ctx context.Context, headers map[string]string) (Response, error)
	StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ListInterests(ctx context.Context, headers map[string]string) (Response, error)
	LookupLocation(ctx context.Context, headers map[string]string, query url.Values) (Response, error)
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

func (c *client) GetMe(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/me",
		Headers: headers,
	})
}

func (c *client) Store(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      http.MethodPost,
		Path:        "/profile/store",
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func (c *client) GetPhoto(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/photo",
		Headers: headers,
	})
}

func (c *client) StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      http.MethodPost,
		Path:        "/profile/photo/store",
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func (c *client) ListInterests(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/interests",
		Headers: headers,
	})
}

func (c *client) LookupLocation(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/location/ip",
		Headers: headers,
		Query:   query,
	})
}