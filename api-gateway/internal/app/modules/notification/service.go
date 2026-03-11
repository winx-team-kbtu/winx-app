package notification

import (
	"context"
	"net/url"
)

type Service interface {
	List(ctx context.Context, headers map[string]string, query url.Values) (Response, error)
	Delete(ctx context.Context, id string, headers map[string]string) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) List(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	return s.client.List(ctx, headers, query)
}

func (s *service) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	return s.client.Delete(ctx, id, headers)
}
