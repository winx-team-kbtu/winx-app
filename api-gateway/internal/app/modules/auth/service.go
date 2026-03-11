package auth

import "context"

type Service interface {
	Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
	VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error)
}

type service struct {
	client Client
}

func NewService(client Client) Service {
	return &service{client: client}
}

func (s *service) Login(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Login(ctx, body, contentType, headers)
}

func (s *service) Register(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Register(ctx, body, contentType, headers)
}

func (s *service) Refresh(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Refresh(ctx, body, contentType, headers)
}

func (s *service) Check(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Check(ctx, body, contentType, headers)
}

func (s *service) Logout(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.Logout(ctx, body, contentType, headers)
}

func (s *service) ForgotPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.ForgotPassword(ctx, body, contentType, headers)
}

func (s *service) ResetPassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.ResetPassword(ctx, body, contentType, headers)
}

func (s *service) ChangePassword(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.ChangePassword(ctx, body, contentType, headers)
}

func (s *service) VerifyPin(ctx context.Context, body []byte, contentType string, headers map[string]string) (Response, error) {
	return s.client.VerifyPin(ctx, body, contentType, headers)
}
