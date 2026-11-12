package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/app/core/validation"
	serviceDto "auth/internal/app/domain/core/dto/services"
	service "auth/internal/app/domain/services"
	tokenService "auth/internal/app/domain/services/token"
	uservalidationservice "auth/internal/app/domain/validation-services/user"
	"auth/internal/app/models/models"
	pkgvalidation "auth/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type consentService struct {
	registered []serviceDto.RegisterDTO
}

func (f *consentService) Login(context.Context, serviceDto.LoginDTO) (tokenService.Response, int, error) {
	return tokenService.Response{}, 0, nil
}

func (f *consentService) Register(_ context.Context, in serviceDto.RegisterDTO) (models.User, error) {
	f.registered = append(f.registered, in)

	version := service.ConsentNoticeVersion

	return models.User{Email: in.Email, ConsentNoticeVersion: &version}, nil
}

func (f *consentService) RefreshToken(context.Context, string) (tokenService.Response, error) {
	return tokenService.Response{}, nil
}

func (f *consentService) CheckToken(context.Context, string) (models.User, error) {
	return models.User{}, nil
}

func (f *consentService) Logout(context.Context, string, string) (bool, error) { return true, nil }

type alwaysUnique struct{}

func (alwaysUnique) Unique(context.Context, string, map[string]any, map[string]any) (bool, error) {
	return true, nil
}

func (alwaysUnique) Exists(context.Context, string, map[string]any, map[string]any) (bool, error) {
	return true, nil
}

func registerRouter(t *testing.T, svc service.Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	v, err := pkgvalidation.New()
	require.NoError(t, err)

	h := NewHandler(svc, validation.NewBinder(v), uservalidationservice.New(alwaysUnique{}))
	r := gin.New()
	r.POST("/register", h.Register)

	return r
}

func postRegister(r http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestRegister_BlockedWithoutConsent(t *testing.T) {
	bodies := map[string]string{
		"field omitted":  `{"email":"a@x.io","password":"pw"}`,
		"explicit false": `{"email":"a@x.io","password":"pw","consent_accepted":false}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			svc := &consentService{}

			rec := postRegister(registerRouter(t, svc), body)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, rec.Body.String(), "consent_accepted", "response must explain what is missing")
			assert.Empty(t, svc.registered, "registration must be blocked")
		})
	}
}

func TestRegister_WithConsentSucceedsAndReportsNoticeVersion(t *testing.T) {
	svc := &consentService{}

	rec := postRegister(registerRouter(t, svc), `{"email":"a@x.io","password":"pw","consent_accepted":true}`)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	require.Len(t, svc.registered, 1)
	assert.True(t, svc.registered[0].ConsentAccepted)
	assert.Contains(t, rec.Body.String(), service.ConsentNoticeVersion)
}

func TestRegister_ConsentMustBeABoolean(t *testing.T) {
	svc := &consentService{}

	rec := postRegister(registerRouter(t, svc), `{"email":"a@x.io","password":"pw","consent_accepted":"yes"}`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, svc.registered)
}
