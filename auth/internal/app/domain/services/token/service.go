package token

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
)

type Response struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

type Service struct {
	OauthServer OAuthServer
}

type OAuthServer interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
}

func NewService(oauthServer OAuthServer) Service {
	return Service{
		OauthServer: oauthServer,
	}
}

func (ts *Service) IssueToken(_ context.Context, params map[string]string) (Response, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Response{}, fmt.Errorf("failed IssueToken when create new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("client-id", "client-secret")

	rec := httptest.NewRecorder()
	if err = ts.OauthServer.HandleTokenRequest(rec, req); err != nil {
		return Response{}, fmt.Errorf("failed IssueToken when HandleTokenRequest: %w", err)
	}

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed IssueToken when ReadAll body: %v\n", err)
	}

	var tr Response
	if err = json.Unmarshal(body, &tr); err != nil {
		return Response{}, fmt.Errorf("invalid token response: %v body=%s", err, string(body))
	}
	if tr.Error != "" {
		return tr, fmt.Errorf("oauth error: %s %s", tr.Error, tr.ErrorDesc)
	}

	return tr, nil
}

func (ts *Service) ValidateToken(ctx context.Context, token string) (oauth2.TokenInfo, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("empty token")
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	ti, err := ts.OauthServer.ValidationBearerToken(req)
	if err != nil {
		return nil, fmt.Errorf("validate bearer token: %w", err)
	}

	return ti, nil
}
