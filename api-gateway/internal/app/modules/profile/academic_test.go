package profile

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockService only implements Academic; the embedded interface is nil, so any
// other method call would panic and fail the test loudly.
type MockService struct {
	mock.Mock
	Service
}

func (m *MockService) Academic(ctx context.Context, method, path string, body []byte, contentType string, headers map[string]string, query url.Values) (Response, error) {
	args := m.Called(ctx, method, path, body, contentType, headers, query)

	return args.Get(0).(Response), args.Error(1)
}

func academicRouter(t *testing.T, svc Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	h := NewHandler(svc)
	r := gin.New()
	group := r.Group("/profile")

	for _, route := range AcademicRoutes {
		group.Handle(route.Method, route.Path, h.Academic(route))
	}

	return r
}

func TestAcademicRoutes_AreTheExpectedAllowList(t *testing.T) {
	got := make([]string, 0, len(AcademicRoutes))
	for _, r := range AcademicRoutes {
		got = append(got, r.Method+" "+r.Path)
	}

	assert.ElementsMatch(t, []string{
		"GET /academic/courses",
		"GET /academic/courses/me",
		"PUT /academic/courses/me",
		"GET /academic/me",
		"PUT /academic/me",
		"GET /academic/deck",
		"PUT /academic/choices",
		"GET /academic/shortlist",
		"POST /academic/teams",
		"GET /academic/teams/me",
		"POST /academic/teams/join",
		"GET /academic/teams/roster",
	}, got)

	for _, r := range AcademicRoutes {
		assert.NotContains(t, r.Path, "..", "no traversal")
		assert.Contains(t, r.Path, "/academic/", "only academic paths may be proxied")
	}
}

func TestAcademic_ForwardsMethodPathBodyQueryAndAuthorization(t *testing.T) {
	svc := new(MockService)
	svc.On("Academic", mock.Anything, http.MethodPut, "/academic/courses/me",
		[]byte(`{"course_ids":[1]}`), "application/json",
		map[string]string{"Authorization": "Bearer t"}, url.Values{}).
		Return(Response{StatusCode: http.StatusOK, ContentType: "application/json", Body: []byte(`{"success":true}`)}, nil)

	req := httptest.NewRequest(http.MethodPut, "/profile/academic/courses/me", bytes.NewBufferString(`{"course_ids":[1]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()

	academicRouter(t, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"success":true}`, rec.Body.String())
	svc.AssertExpectations(t)
}

func TestAcademic_GetForwardsQueryWithoutBody(t *testing.T) {
	svc := new(MockService)
	svc.On("Academic", mock.Anything, http.MethodGet, "/academic/deck",
		[]byte(nil), mock.Anything, mock.Anything, url.Values{"course_id": {"10"}, "limit": {"5"}}).
		Return(Response{StatusCode: http.StatusOK, ContentType: "application/json", Body: []byte(`{}`)}, nil)

	req := httptest.NewRequest(http.MethodGet, "/profile/academic/deck?course_id=10&limit=5", nil)
	rec := httptest.NewRecorder()

	academicRouter(t, svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAcademic_PassesUpstreamStatusThrough(t *testing.T) {
	svc := new(MockService)
	svc.On("Academic", mock.Anything, http.MethodPost, "/academic/teams/join", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(Response{StatusCode: http.StatusConflict, ContentType: "application/json", Body: []byte(`{"message":"team is full"}`)}, nil)

	req := httptest.NewRequest(http.MethodPost, "/profile/academic/teams/join", bytes.NewBufferString(`{"join_code":"ABC234"}`))
	rec := httptest.NewRecorder()

	academicRouter(t, svc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code, "business errors must reach the client unchanged")
	assert.Contains(t, rec.Body.String(), "team is full")
}

func TestAcademic_UpstreamFailureIs502(t *testing.T) {
	svc := new(MockService)
	svc.On("Academic", mock.Anything, http.MethodGet, "/academic/me", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(Response{}, errors.New("connection refused"))

	rec := httptest.NewRecorder()
	academicRouter(t, svc).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/profile/academic/me", nil))

	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestAcademic_UnlistedPathAndMethodAreNotProxied(t *testing.T) {
	svc := new(MockService)
	r := academicRouter(t, svc)

	for _, c := range []struct{ method, path string }{
		{http.MethodDelete, "/profile/academic/me"},
		{http.MethodGet, "/profile/academic/teams/join"},
		{http.MethodGet, "/profile/academic/../me"},
		{http.MethodGet, "/profile/photo"},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))

		assert.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusMovedPermanently}, rec.Code, "%s %s", c.method, c.path)
	}

	svc.AssertNotCalled(t, "Academic", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
