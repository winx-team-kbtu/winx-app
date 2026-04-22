package match

import "context"

type Service interface {
	SwipeLeft(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	SwipeRight(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	List(ctx context.Context, headers map[string]string) (Response, error)
	Delete(ctx context.Context, id string, headers map[string]string) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) SwipeLeft(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.SwipeLeft(ctx, body, contentType, headers)
}

func (s *service) SwipeRight(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.SwipeRight(ctx, body, contentType, headers)
}

func (s *service) List(ctx context.Context, headers map[string]string) (Response, error) {
	return s.client.List(ctx, headers)
}

func (s *service) Delete(ctx context.Context, id string, headers map[string]string) (Response, error) {
	return s.client.Delete(ctx, id, headers)
}
