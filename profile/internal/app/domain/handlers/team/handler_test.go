package team

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/validation"
	service "winx-profile/internal/app/domain/services/team"
	"winx-profile/internal/app/models/models"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	err   error
	calls int
}

func (f *fakeService) Create(_ context.Context, _, courseID int64, name string) (models.Team, error) {
	f.calls++
	if f.err != nil {
		return models.Team{}, f.err
	}

	return models.Team{ID: 1, CourseID: courseID, Name: name, JoinCode: "ABC234", MemberCount: 1}, nil
}

func (f *fakeService) GetMine(_ context.Context, _, courseID int64) (models.Team, error) {
	f.calls++
	if f.err != nil {
		return models.Team{}, f.err
	}

	return models.Team{ID: 1, CourseID: courseID, Name: "T", JoinCode: "ABC234", MemberCount: 1}, nil
}

func newRouter(t *testing.T, svc service.Service, authenticated bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	v, err := pkgvalidation.New()
	if err != nil {
		t.Fatal(err)
	}

	h := NewHandler(svc, validation.NewBinder(v))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if authenticated {
			ctx := context.WithValue(c.Request.Context(), headercontract.AuthUserKey{},
				headercontract.AuthUser{ID: 7, Email: "u@example.com"})
			c.Request = c.Request.WithContext(ctx)
		}
	})
	r.POST("/teams", h.Store)
	r.GET("/teams/me", h.GetMine)

	return r
}

func do(r http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestStore_Created(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, true), http.MethodPost, "/teams", `{"course_id":10,"name":"Dream Team"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"join_code":"ABC234"`)) {
		t.Fatalf("join code missing: %s", rec.Body)
	}
}

func TestStore_ErrorMapping(t *testing.T) {
	cases := map[string]struct {
		err  error
		want int
	}{
		"already in team":     {service.ErrAlreadyInTeam, http.StatusConflict},
		"course not selected": {service.ErrCourseNotSelected, http.StatusForbidden},
		"invalid name":        {service.ErrInvalidName, http.StatusUnprocessableEntity},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := do(newRouter(t, &fakeService{err: tc.err}, true), http.MethodPost, "/teams", `{"course_id":10,"name":"Team"}`)
			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestStore_InvalidPayloadIs422WithoutServiceCall(t *testing.T) {
	for _, body := range []string{`{}`, `{"course_id":10}`, `{"name":"Team"}`, `{"course_id":-1,"name":"Team"}`} {
		svc := &fakeService{}

		if rec := do(newRouter(t, svc, true), http.MethodPost, "/teams", body); rec.Code != http.StatusUnprocessableEntity || svc.calls != 0 {
			t.Fatalf("%s: expected 422 and no call, got %d calls=%d", body, rec.Code, svc.calls)
		}
	}
}

func TestGetMine_NotFoundIs404(t *testing.T) {
	rec := do(newRouter(t, &fakeService{err: service.ErrNotFound}, true), http.MethodGet, "/teams/me?course_id=10", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetMine_RequiresCourseID(t *testing.T) {
	if rec := do(newRouter(t, &fakeService{}, true), http.MethodGet, "/teams/me", ""); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestRequiresAuthUser(t *testing.T) {
	r := newRouter(t, &fakeService{}, false)

	if rec := do(r, http.MethodPost, "/teams", `{"course_id":10,"name":"Team"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST expected 401, got %d", rec.Code)
	}

	if rec := do(r, http.MethodGet, "/teams/me?course_id=10", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET expected 401, got %d", rec.Code)
	}
}
