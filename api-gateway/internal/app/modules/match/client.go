package match

import (
	"context"
	"net/http"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	SwipeLeft(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	SwipeRight(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	List(ctx context.Context, headers map[string]string) (Response, error)
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

func (c *client) SwipeLeft(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      http.MethodPost,
		Path:        "/swipes/left",
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func (c *client) SwipeRight(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:      http.MethodPost,
		Path:        "/swipes/right",
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
}

func (c *client) List(ctx context.Context, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/matches",
		Headers: headers,
	})
}

func (c *client) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodDelete,
		Path:    "/matches/" + id,
		Headers: headers,
	})
}
