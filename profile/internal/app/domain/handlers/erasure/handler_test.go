package erasure

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeService struct {
	gotID int64
	err   error
	calls int
}

func (f *fakeService) DeleteUserData(_ context.Context, userID int64) error {
	f.calls++
	f.gotID = userID

	return f.err
}

func do(svc *fakeService, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/internal/users/:id", NewHandler(svc).DeleteUser)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, target, nil))

	return rec
}

func TestDeleteUser_Success(t *testing.T) {
	svc := &fakeService{}

	rec := do(svc, "/internal/users/42")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(42), svc.gotID)
}

func TestDeleteUser_InvalidIDIs400WithoutServiceCall(t *testing.T) {
	for _, id := range []string{"abc", "0", "-5", "1.5"} {
		svc := &fakeService{}

		rec := do(svc, "/internal/users/"+id)

		assert.Equal(t, http.StatusBadRequest, rec.Code, id)
		assert.Zero(t, svc.calls, id)
	}
}

func TestDeleteUser_FailureIsNotReportedAsSuccess(t *testing.T) {
	rec := do(&fakeService{err: errors.New("mongo down")}, "/internal/users/42")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "mongo down", "internal errors must not leak")
}
