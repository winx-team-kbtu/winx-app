package chat

import (
	"context"
	"net/url"
)

type Service interface {
	List(ctx context.Context, query url.Values, headers map[string]string) (Response, error)
	Messages(ctx context.Context, chatID string, query url.Values, headers map[string]string) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) List(ctx context.Context, query url.Values, headers map[string]string) (Response, error) {
	return s.client.List(ctx, query, headers)
}

func (s *service) Messages(ctx context.Context, chatID string, query url.Values, headers map[string]string) (Response, error) {
	return s.client.Messages(ctx, chatID, query, headers)
}
