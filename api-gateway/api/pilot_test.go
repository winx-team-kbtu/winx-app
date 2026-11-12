package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// socialRoutes are every legacy Social endpoint exposed by the gateway.
var socialRoutes = []struct{ method, path string }{
	{http.MethodGet, "/api/v1/profile/photo"},
	{http.MethodPost, "/api/v1/profile/photo/store"},
	{http.MethodPost, "/api/v1/swipes/left"},
	{http.MethodPost, "/api/v1/swipes/right"},
	{http.MethodGet, "/api/v1/matches"},
	{http.MethodDelete, "/api/v1/matches/1"},
	{http.MethodGet, "/api/v1/recommendations"},
	{http.MethodGet, "/api/v1/chats"},
	{http.MethodGet, "/api/v1/chats/1/messages"},
}

func newTestEngine(pilotMode bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	handler = gin.New()
	handler.Use(gin.Recovery())

	// Services are nil: in pilot mode no Social handler may ever be reached.
	s := &Server{pilotMode: pilotMode}
	s.initDomainRoutes()

	return handler
}

func TestPilotMode_SocialEndpointsAreDisabled(t *testing.T) {
	engine := newTestEngine(true)

	for _, r := range socialRoutes {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))

		assert.Equal(t, http.StatusNotFound, rec.Code, "%s %s", r.method, r.path)
		assert.Contains(t, rec.Body.String(), socialDisabledMessage, "%s %s", r.method, r.path)
	}
}

func TestPilotMode_AcademicAndOtherRoutesStayRegistered(t *testing.T) {
	engine := newTestEngine(true)

	registered := map[string]bool{}
	for _, r := range engine.Routes() {
		registered[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"POST /api/v1/register",
		"POST /api/v1/login",
		"GET /api/v1/profile/me",
		"POST /api/v1/profile/store",
		"GET /api/v1/profile/academic/courses",
		"PUT /api/v1/profile/academic/courses/me",
		"GET /api/v1/profile/academic/deck",
		"POST /api/v1/profile/academic/teams/join",
		"GET /api/v1/profile/academic/teams/roster",
	} {
		require.True(t, registered[want], "%s must stay available in pilot mode", want)
	}
}

func TestNonPilotMode_SocialHandlersAreReached(t *testing.T) {
	engine := newTestEngine(false)

	for _, r := range socialRoutes {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))

		assert.NotContains(t, rec.Body.String(), socialDisabledMessage, "%s %s must reach the real handler", r.method, r.path)
	}
}

func TestSocialHandler_OnlyWrapsInPilotMode(t *testing.T) {
	called := false
	inner := func(ctx *gin.Context) { called = true; ctx.Status(http.StatusOK) }

	run := func(h gin.HandlerFunc) int {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		h(ctx)

		return rec.Code
	}

	assert.Equal(t, http.StatusOK, run(socialHandler(false, inner)))
	assert.True(t, called)

	called = false

	assert.Equal(t, http.StatusNotFound, run(socialHandler(true, inner)))
	assert.False(t, called, "inner handler must not run in pilot mode")
}
