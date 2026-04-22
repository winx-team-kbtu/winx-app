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
	GetPhotos(ctx context.Context, headers map[string]string) (Response, error)
	StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ListInterests(ctx context.Context, headers map[string]string, query url.Values) (Response, error)
	LookupLocationByIP(ctx context.Context, headers map[string]string) (Response, error)
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

func (c *client) GetPhotos(ctx context.Context, headers map[string]string) (Response, error) {
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

func (c *client) ListInterests(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/interests",
		Headers: headers,
		Query:   cloneQuery(query),
	})
}

func (c *client) LookupLocationByIP(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/profile/location/ip",
		Headers: headers,
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
