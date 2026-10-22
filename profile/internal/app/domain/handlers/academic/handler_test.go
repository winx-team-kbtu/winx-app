package academic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/validation"
	service "winx-profile/internal/app/domain/services/academic"
	"winx-profile/internal/app/models/models"
	pkgvalidation "winx-profile/pkg/validation"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	saved *service.SaveInput
	err   error
}

func (f *fakeService) Get(context.Context, int64) (models.AcademicProfile, error) {
	return models.AcademicProfile{Skills: []string{"Go"}, PreferredRole: "Backend"}, nil
}

func (f *fakeService) Save(_ context.Context, _ int64, in service.SaveInput) (models.AcademicProfile, error) {
	f.saved = &in
	if f.err != nil {
		return models.AcademicProfile{}, f.err
	}

	return models.AcademicProfile{Skills: in.Skills, PreferredRole: in.PreferredRole}, nil
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
	r.GET("/me", h.GetMe)
	r.PUT("/me", h.SaveMe)

	return r
}

func do(r http.Handler, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestGetMe(t *testing.T) {
	rec := do(newRouter(t, &fakeService{}, true), http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var body struct {
		Data struct {
			Skills        []string `json:"skills"`
			PreferredRole string   `json:"preferred_role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if len(body.Data.Skills) != 1 || body.Data.PreferredRole != "Backend" {
		t.Fatalf("unexpected: %s", rec.Body)
	}
}

func TestSaveMe_Success(t *testing.T) {
	svc := &fakeService{}

	rec := do(newRouter(t, svc, true), http.MethodPut, `{"skills":["Go"],"preferred_role":"Backend"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}

	if svc.saved == nil || svc.saved.PreferredRole != "Backend" {
		t.Fatalf("service got %+v", svc.saved)
	}
}

func TestSaveMe_InvalidInputIs422(t *testing.T) {
	svc := &fakeService{err: service.ErrInvalidInput}

	rec := do(newRouter(t, svc, true), http.MethodPut, `{"skills":["x"]}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestSaveMe_OversizedPayloadRejectedBeforeService(t *testing.T) {
	svc := &fakeService{}
	big := make([]byte, 0)
	big = append(big, `{"preferred_role":"`...)
	for i := 0; i < 300; i++ {
		big = append(big, 'r')
	}
	big = append(big, `"}`...)

	rec := do(newRouter(t, svc, true), http.MethodPut, string(big))
	if rec.Code != http.StatusUnprocessableEntity || svc.saved != nil {
		t.Fatalf("expected 422 and no service call, got %d saved=%v", rec.Code, svc.saved)
	}
}

func TestSaveMe_MalformedJSONIs400(t *testing.T) {
	if rec := do(newRouter(t, &fakeService{}, true), http.MethodPut, `{`); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequiresAuthUser(t *testing.T) {
	r := newRouter(t, &fakeService{}, false)

	if rec := do(r, http.MethodGet, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET expected 401, got %d", rec.Code)
	}

	if rec := do(r, http.MethodPut, `{}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT expected 401, got %d", rec.Code)
	}
}
