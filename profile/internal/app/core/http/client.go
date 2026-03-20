package http

import (
	"context"
	"net"
	"net/http"
	"time"
)

type ClientBase struct {
	Client *http.Client
}

type ClientConfig struct {
	Timeout time.Duration
}

func New(cfg ClientConfig) *ClientBase {
	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	return &ClientBase{
		Client: &http.Client{
			Transport: tr,
			Timeout:   cfg.Timeout,
		},
	}
}

func (c *ClientBase) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	return c.Client.Do(req)
}
