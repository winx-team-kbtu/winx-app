package choice

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/validation"
	service "winx-profile/internal/app/domain/services/choice"
	"winx-profile/internal/app/models/models"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	err   error
	calls int
	got   struct {
		user, candidate, course int64
		choice                  string
	}
}

func (f *fakeService) Choose(_ context.Context, userID, candidateID, courseID int64, choice string) error {
	f.calls++
	f.got.user, f.got.candidate, f.got.course, f.got.choice = userID, candidateID, courseID, choice

	return f.err
}

func (f *fakeService) Shortlist(context.Context, int64, int64, int64, int64) ([]models.DeckCandidate, error) {
	return []models.DeckCandidate{{UserID: 2, Name: "Ann"}}, f.err
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
	r.PUT("/choices", h.Store)
	r.GET("/shortlist", h.Shortlist)

	return r
}

func do(r http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestStore_Success(t *testing.T) {
	svc := &fakeService{}

	rec := do(newRouter(t, svc, true), http.MethodPut, "/choices", `{"course_id":10,"candidate_id":2,"choice":"like"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}

	if svc.got.user != 7 || svc.got.candidate != 2 || svc.got.course != 10 || svc.got.choice != "like" {
		t.Fatalf("service got %+v", svc.got)
	}
}

func TestStore_ErrorMapping(t *testing.T) {
	cases := map[string]struct {
		err  error
		want int
	}{
		"course not selected": {service.ErrCourseNotSelected, http.StatusForbidden},
		"invalid candidate":   {service.ErrInvalidCandidate, http.StatusUnprocessableEntity},
		"invalid choice":      {service.ErrInvalidChoice, http.StatusUnprocessableEntity},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := do(newRouter(t, &fakeService{err: tc.err}, true), http.MethodPut, "/choices",
				`{"course_id":10,"candidate_id":2,"choice":"pass"}`)
			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestStore_InvalidPayloadIs422WithoutServiceCall(t *testing.T) {
	bodies := []string{
		`{}`,
		`{"course_id":10,"candidate_id":2,"choice":"superlike"}`,
		`{"course_id":0,"candidate_id":2,"choice":"like"}`,
		`{"course_id":10,"candidate_id":-1,"choice":"like"}`,
	}

	for _, body := range bodies {
		svc := &fakeService{}

		if rec := do(newRouter(t, svc, true), http.MethodPut, "/choices", body); rec.Code != http.StatusUnprocessableEntity || svc.calls != 0 {
			t.Fatalf("%s: expected 422 and no call, got %d calls=%d", body, rec.Code, svc.calls)
		}
	}
}

func TestShortlist_ReturnsEntriesAndValidatesQuery(t *testing.T) {
	r := newRouter(t, &fakeService{}, true)

	if rec := do(r, http.MethodGet, "/shortlist?course_id=10", ""); rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	if rec := do(r, http.MethodGet, "/shortlist", ""); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing course_id expected 422, got %d", rec.Code)
	}
}

func TestShortlist_UnselectedCourseIs403(t *testing.T) {
	rec := do(newRouter(t, &fakeService{err: service.ErrCourseNotSelected}, true), http.MethodGet, "/shortlist?course_id=10", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequiresAuthUser(t *testing.T) {
	r := newRouter(t, &fakeService{}, false)

	if rec := do(r, http.MethodPut, "/choices", `{"course_id":10,"candidate_id":2,"choice":"like"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT expected 401, got %d", rec.Code)
	}

	if rec := do(r, http.MethodGet, "/shortlist?course_id=10", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET expected 401, got %d", rec.Code)
	}
}
