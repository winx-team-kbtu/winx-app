package token

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-oauth2/oauth2/v4"
)

type TokenInfoMock struct {
	AccessToken string
	UserID      string
}

func (t *TokenInfoMock) New() oauth2.TokenInfo                               { return nil }
func (t *TokenInfoMock) SetClientID(_ string)                                {}
func (t *TokenInfoMock) SetUserID(_ string)                                  {}
func (t *TokenInfoMock) GetRedirectURI() string                              { return "" }
func (t *TokenInfoMock) SetRedirectURI(_ string)                             {}
func (t *TokenInfoMock) SetScope(_ string)                                   {}
func (t *TokenInfoMock) GetCode() string                                     { return "" }
func (t *TokenInfoMock) SetCode(_ string)                                    {}
func (t *TokenInfoMock) GetCodeCreateAt() time.Time                          { return time.Now() }
func (t *TokenInfoMock) SetCodeCreateAt(_ time.Time)                         {}
func (t *TokenInfoMock) GetCodeExpiresIn() time.Duration                     { return time.Hour }
func (t *TokenInfoMock) SetCodeExpiresIn(_ time.Duration)                    {}
func (t *TokenInfoMock) GetCodeChallenge() string                            { return "" }
func (t *TokenInfoMock) SetCodeChallenge(_ string)                           {}
func (t *TokenInfoMock) GetCodeChallengeMethod() oauth2.CodeChallengeMethod  { return "" }
func (t *TokenInfoMock) SetCodeChallengeMethod(_ oauth2.CodeChallengeMethod) {}
func (t *TokenInfoMock) SetAccess(_ string)                                  {}
func (t *TokenInfoMock) SetAccessCreateAt(_ time.Time)                       {}
func (t *TokenInfoMock) GetAccessExpiresIn() time.Duration                   { return time.Hour }
func (t *TokenInfoMock) SetAccessExpiresIn(_ time.Duration)                  {}
func (t *TokenInfoMock) SetRefresh(_ string)                                 {}
func (t *TokenInfoMock) SetRefreshCreateAt(_ time.Time)                      {}
func (t *TokenInfoMock) GetRefreshExpiresIn() time.Duration                  { return time.Duration(0) }
func (t *TokenInfoMock) SetRefreshExpiresIn(_ time.Duration)                 {}
func (t *TokenInfoMock) GetAccess() string                                   { return t.AccessToken }
func (t *TokenInfoMock) GetRefresh() string                                  { return "" }
func (t *TokenInfoMock) GetClientID() string                                 { return "test-client" }
func (t *TokenInfoMock) GetUserID() string                                   { return t.UserID }
func (t *TokenInfoMock) GetScope() string                                    { return "" }
func (t *TokenInfoMock) GetIssuedAt() int64                                  { return time.Now().Unix() }
func (t *TokenInfoMock) GetExpiresIn() int64                                 { return 3600 }
func (t *TokenInfoMock) GetAccessCreateAt() time.Time                        { return time.Now() }
func (t *TokenInfoMock) GetRefreshCreateAt() time.Time {
	return time.Now()
}

type OAuthServerMock struct {
	HandleTokenRequestFn  func(w http.ResponseWriter, r *http.Request) error
	ValidateBearerTokenFn func(r *http.Request) (oauth2.TokenInfo, error)
}

func (m *OAuthServerMock) HandleTokenRequest(w http.ResponseWriter, r *http.Request) error {
	if m.HandleTokenRequestFn != nil {
		return m.HandleTokenRequestFn(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return json.NewEncoder(w).Encode(map[string]any{
		"access_token":  "test-access-token",
		"refresh_token": "test-refresh-token",
		"token_type":    "Bearer",
		"expires_in":    3600,
	})
}

func (m *OAuthServerMock) ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	if m.ValidateBearerTokenFn != nil {
		return m.ValidateBearerTokenFn(r)
	}

	return &TokenInfoMock{
		AccessToken: "test-access-token",
		UserID:      "1",
	}, nil
}
