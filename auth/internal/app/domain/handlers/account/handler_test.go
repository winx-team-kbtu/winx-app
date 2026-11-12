package account

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/validation"
	service "auth/internal/app/domain/services/deletion"
	pkgvalidation "auth/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	err      error
	calls    int
	gotID    int64
	gotPassw string
}

func (f *fakeService) Delete(_ context.Context, userID int64, password string) error {
	f.calls++
	f.gotID, f.gotPassw = userID, password

	return f.err
}

func do(t *testing.T, svc *fakeService, authenticated bool, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	v, err := pkgvalidation.New()
	require.NoError(t, err)

	h := NewHandler(svc, validation.NewBinder(v))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if authenticated {
			ctx := context.WithValue(c.Request.Context(), headercontract.AuthUserKey{}, headercontract.AuthUser{ID: 42, Email: "u@test.io"})
			c.Request = c.Request.WithContext(ctx)
		}
	})
	r.DELETE("/account", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/account", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestDelete_Success(t *testing.T) {
	svc := &fakeService{}

	rec := do(t, svc, true, `{"password":"secret"}`)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(42), svc.gotID, "the account to delete comes from the session, never from the body")
	assert.Equal(t, "secret", svc.gotPassw)
}

func TestDelete_ErrorMapping(t *testing.T) {
	cases := map[string]struct {
		err  error
		want int
	}{
		"wrong password": {service.ErrInvalidPassword, http.StatusForbidden},
		"not found":      {service.ErrNotFound, http.StatusNotFound},
		"pending":        {errors.Join(service.ErrPending, errors.New("profile down")), http.StatusServiceUnavailable},
		"unexpected":     {errors.New("kaboom"), http.StatusInternalServerError},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := do(t, &fakeService{err: tc.err}, true, `{"password":"secret"}`)

			assert.Equal(t, tc.want, rec.Code)
			assert.NotContains(t, rec.Body.String(), "kaboom", "internal errors must not leak")
			assert.NotContains(t, rec.Body.String(), "profile down", "internal causes must not leak")
		})
	}
}

func TestDelete_PendingIsNeverASuccessResponse(t *testing.T) {
	rec := do(t, &fakeService{err: service.ErrPending}, true, `{"password":"secret"}`)

	assert.NotEqual(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"success":false`)
}

func TestDelete_RequiresPasswordConfirmation(t *testing.T) {
	for _, body := range []string{`{}`, `{"password":""}`} {
		svc := &fakeService{}

		rec := do(t, svc, true, body)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, body)
		assert.Zero(t, svc.calls, "no deletion without confirmation: %s", body)
	}
}

func TestDelete_RequiresAuthenticatedUser(t *testing.T) {
	svc := &fakeService{}

	rec := do(t, svc, false, `{"password":"secret"}`)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Zero(t, svc.calls)
}
