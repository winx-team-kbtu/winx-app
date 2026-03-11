package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type Request struct {
	Method      string
	Path        string
	ContentType string
	Body        []byte
	Headers     map[string]string
	Query       url.Values
}

type Client struct {
	baseURL        string
	internalAPIKey string
	httpClient     *http.Client
}

func NewClient(baseURL, internalAPIKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		internalAPIKey: internalAPIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Do(ctx context.Context, req Request) (Response, error) {
	method := req.Method
	if strings.TrimSpace(method) == "" {
		method = http.MethodGet
	}

	requestURL := c.baseURL + req.Path
	if len(req.Query) > 0 {
		requestURL += "?" + req.Query.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(req.Body))
	if err != nil {
		return Response{}, fmt.Errorf("create proxy request: %w", err)
	}

	if strings.TrimSpace(req.ContentType) != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}
	if strings.TrimSpace(c.internalAPIKey) != "" {
		httpReq.Header.Set("x-api-key", c.internalAPIKey)
	}
	for key, value := range req.Headers {
		if strings.TrimSpace(value) != "" {
			httpReq.Header.Set(key, value)
		}
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("perform proxy request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read proxy response: %w", err)
	}

	return Response{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        responseBody,
	}, nil
}
