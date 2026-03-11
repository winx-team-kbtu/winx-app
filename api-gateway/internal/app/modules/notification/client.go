package notification

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	List(ctx context.Context, headers map[string]string, query url.Values) (Response, error)
	Delete(ctx context.Context, id string, headers map[string]string) (Response, error)
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

func (c *client) List(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/notifications",
		Headers: headers,
		Query:   cloneQuery(query),
	})
}

func (c *client) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodDelete,
		Path:    "/notifications/" + id,
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
