package deck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/validation"
	service "winx-profile/internal/app/domain/services/deck"
	"winx-profile/internal/app/models/models"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	items     []models.DeckCandidate
	err       error
	calls     int
	gotCourse int64
	gotLimit  int64
	gotOffset int64
}

func (f *fakeService) Candidates(_ context.Context, _, courseID, limit, offset int64) ([]models.DeckCandidate, error) {
	f.calls++
	f.gotCourse, f.gotLimit, f.gotOffset = courseID, limit, offset

	return f.items, f.err
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
	r.GET("/deck", h.List)

	return r
}

func get(r http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return rec
}

func TestList_ReturnsOnlyAcademicFields(t *testing.T) {
	svc := &fakeService{items: []models.DeckCandidate{{UserID: 2, Name: "Ann", Skills: []string{"Go"}, PreferredRole: "Backend"}}}

	rec := get(newRouter(t, svc, true), "/deck?course_id=10&limit=5&offset=2")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if len(body.Data) != 1 {
		t.Fatalf("unexpected body %s", rec.Body)
	}

	allowed := map[string]bool{"user_id": true, "name": true, "skills": true, "preferred_role": true}
	for key := range body.Data[0] {
		if !allowed[key] {
			t.Fatalf("deck leaks non-academic field %q", key)
		}
	}

	if svc.gotCourse != 10 || svc.gotLimit != 5 || svc.gotOffset != 2 {
		t.Fatalf("query not forwarded: %+v", svc)
	}
}

func TestList_EmptyDeckIsEmptyArray(t *testing.T) {
	rec := get(newRouter(t, &fakeService{}, true), "/deck?course_id=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var body struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("expected [] got %s", rec.Body)
	}
}

func TestList_UnselectedCourseIs403(t *testing.T) {
	rec := get(newRouter(t, &fakeService{err: service.ErrCourseNotSelected}, true), "/deck?course_id=10")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestList_InvalidQueryIs422WithoutServiceCall(t *testing.T) {
	for _, target := range []string{"/deck", "/deck?course_id=0", "/deck?course_id=1&limit=500", "/deck?course_id=1&offset=-1"} {
		svc := &fakeService{}

		if rec := get(newRouter(t, svc, true), target); rec.Code != http.StatusUnprocessableEntity || svc.calls != 0 {
			t.Fatalf("%s: expected 422 and no call, got %d calls=%d", target, rec.Code, svc.calls)
		}
	}
}

func TestList_RequiresAuthUser(t *testing.T) {
	if rec := get(newRouter(t, &fakeService{}, false), "/deck?course_id=10"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
