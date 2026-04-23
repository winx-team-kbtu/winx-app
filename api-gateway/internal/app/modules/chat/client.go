package chat

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"winx-api-gateway/internal/app/proxy"
)

type Client interface {
	List(ctx context.Context, query url.Values, headers map[string]string) (Response, error)
	Messages(ctx context.Context, chatID string, query url.Values, headers map[string]string) (Response, error)
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

func (c *client) List(ctx context.Context, query url.Values, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/chats",
		Query:   query,
		Headers: headers,
	})
}

func (c *client) Messages(ctx context.Context, chatID string, query url.Values, headers map[string]string) (Response, error) {
	return c.proxy.Do(ctx, proxy.Request{
		Method:  http.MethodGet,
		Path:    "/chats/" + chatID + "/messages",
		Query:   query,
		Headers: headers,
	})
}
