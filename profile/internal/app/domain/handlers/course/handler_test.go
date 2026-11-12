package course

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/validation"
	service "winx-profile/internal/app/domain/services/course"
	"winx-profile/internal/app/models/models"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	userCourses []models.Course
	saveErr     error
	savedIDs    []int64
}

func (f *fakeService) List(context.Context) ([]models.Course, error) {
	return []models.Course{{ID: 1, Code: "ALG201", Title: "Algorithms"}}, nil
}

func (f *fakeService) GetUserCourses(context.Context, int64) ([]models.Course, error) {
	return f.userCourses, nil
}

func (f *fakeService) SaveUserCourses(_ context.Context, _ int64, ids []int64) ([]models.Course, error) {
	f.savedIDs = ids
	if f.saveErr != nil {
		return nil, f.saveErr
	}

	return []models.Course{{ID: 1, Code: "ALG201", Title: "Algorithms"}}, nil
}

func newRouter(t *testing.T, svc service.Service, authenticated bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	v, err := pkgvalidation.New()
	if err != nil {
		t.Fatalf("validator: %v", err)
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
	r.GET("/courses", h.List)
	r.GET("/courses/me", h.GetMine)
	r.PUT("/courses/me", h.SaveMine)

	return r
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestList_ReturnsCatalogue(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, true), http.MethodGet, "/courses", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if len(body.Data) != 1 || body.Data[0]["code"] != "ALG201" {
		t.Fatalf("unexpected body: %s", rec.Body)
	}
}

func TestSaveMine_Success(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(t, svc, true), http.MethodPut, "/courses/me", `{"course_ids":[1]}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}

	if len(svc.savedIDs) != 1 || svc.savedIDs[0] != 1 {
		t.Fatalf("service got %v", svc.savedIDs)
	}
}

func TestSaveMine_UnknownCourseIs422(t *testing.T) {
	svc := &fakeService{saveErr: service.ErrInvalidCourseIDs}
	rec := do(newRouter(t, svc, true), http.MethodPut, "/courses/me", `{"course_ids":[999]}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body)
	}
}

func TestSaveMine_RejectsNonPositiveIDs(t *testing.T) {
	svc := &fakeService{}
	rec := do(newRouter(t, svc, true), http.MethodPut, "/courses/me", `{"course_ids":[0,-3]}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body)
	}

	if svc.savedIDs != nil {
		t.Fatalf("service must not be called, got %v", svc.savedIDs)
	}
}

func TestSaveMine_MalformedJSONIs400(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, true), http.MethodPut, "/courses/me", `{"course_ids":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMine_RequiresAuthUser(t *testing.T) {
	r := newRouter(t, &fakeService{}, false)

	if rec := do(r, http.MethodGet, "/courses/me", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET expected 401, got %d", rec.Code)
	}

	if rec := do(r, http.MethodPut, "/courses/me", `{"course_ids":[1]}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT expected 401, got %d", rec.Code)
	}
}
