package profile

import (
	"context"
	"net/url"
)

type Service interface {
	GetMe(ctx context.Context, headers map[string]string) (Response, error)
	Store(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	GetPhoto(ctx context.Context, headers map[string]string) (Response, error)
	StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ListInterests(ctx context.Context, headers map[string]string) (Response, error)
	LookupLocation(ctx context.Context, headers map[string]string, query url.Values) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) GetMe(ctx context.Context, headers map[string]string) (Response, error) {
	return s.client.GetMe(ctx, headers)
}

func (s *service) Store(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Store(ctx, body, contentType, headers)
}

func (s *service) GetPhoto(ctx context.Context, headers map[string]string) (Response, error) {
	return s.client.GetPhoto(ctx, headers)
}

func (s *service) StorePhoto(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.StorePhoto(ctx, body, contentType, headers)
}

func (s *service) ListInterests(ctx context.Context, headers map[string]string) (Response, error) {
	return s.client.ListInterests(ctx, headers)
}

func (s *service) LookupLocation(ctx context.Context, headers map[string]string, query url.Values) (Response, error) {
	return s.client.LookupLocation(ctx, headers, query)
}