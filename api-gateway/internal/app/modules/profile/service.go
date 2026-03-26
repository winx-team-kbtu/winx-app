package profile

import (
	"context"
)

type Service interface {
	Get(ctx context.Context, headers map[string]string) (Response, error)
	Update(ctx context.Context, headers map[string]string, body []byte) (Response, error)
	UpdateRole(ctx context.Context, headers map[string]string, body []byte) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) Get(ctx context.Context, headers map[string]string) (Response, error) {
	return s.client.Get(ctx, headers)
}

func (s *service) Update(ctx context.Context, headers map[string]string, body []byte) (Response, error) {
	return s.client.Update(ctx, headers, body)
}

func (s *service) UpdateRole(ctx context.Context, headers map[string]string, body []byte) (Response, error) {
	return s.client.UpdateRole(ctx, headers, body)
}
