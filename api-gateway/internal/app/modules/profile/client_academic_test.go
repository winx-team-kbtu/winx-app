package profile

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_AcademicTargetsProfileServiceWithKeyAndAuth(t *testing.T) {
	var (
		gotMethod, gotPath, gotQuery, gotKey, gotAuth, gotBody, gotType string
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		gotKey, gotAuth, gotBody, gotType = r.Header.Get("x-api-key"), r.Header.Get("Authorization"), string(body), r.Header.Get("Content-Type")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	c := NewClient(upstream.URL, "internal-key", 2*time.Second)

	resp, err := c.Academic(context.Background(), http.MethodPost, "/academic/teams",
		[]byte(`{"course_id":1,"name":"Alpha"}`), "application/json",
		map[string]string{"Authorization": "Bearer abc"}, url.Values{"x": {"1"}})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, string(resp.Body))
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/profile/academic/teams", gotPath, "gateway path maps 1:1 under /profile")
	assert.Equal(t, "x=1", gotQuery)
	assert.Equal(t, "internal-key", gotKey)
	assert.Equal(t, "Bearer abc", gotAuth)
	assert.JSONEq(t, `{"course_id":1,"name":"Alpha"}`, gotBody)
	assert.Equal(t, "application/json", gotType)
}
