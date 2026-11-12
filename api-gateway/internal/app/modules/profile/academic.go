package profile

import (
	"net/http"

	"winx-api-gateway/internal/app/transport"

	"github.com/gin-gonic/gin"
)

// AcademicRoute is one endpoint of the academic pilot that the gateway exposes.
// Only exact method+path pairs from AcademicRoutes are proxied, never a wildcard,
// so a client cannot reach other profile service paths through this handler.
type AcademicRoute struct {
	Method string
	// Path is relative to /profile on both the gateway and the profile service.
	Path string
}

var AcademicRoutes = []AcademicRoute{
	{http.MethodGet, "/academic/courses"},
	{http.MethodGet, "/academic/courses/me"},
	{http.MethodPut, "/academic/courses/me"},
	{http.MethodGet, "/academic/me"},
	{http.MethodPut, "/academic/me"},
	{http.MethodGet, "/academic/deck"},
	{http.MethodPut, "/academic/choices"},
	{http.MethodGet, "/academic/shortlist"},
	{http.MethodPost, "/academic/teams"},
	{http.MethodGet, "/academic/teams/me"},
	{http.MethodPost, "/academic/teams/join"},
	{http.MethodGet, "/academic/teams/roster"},
}

// Academic returns a handler that proxies one fixed academic endpoint.
func (h *Handler) Academic(route AcademicRoute) gin.HandlerFunc {
	hasBody := route.Method == http.MethodPost || route.Method == http.MethodPut

	return func(ctx *gin.Context) {
		var body []byte

		if hasBody {
			var err error

			body, err = transport.ReadBody(ctx)
			if err != nil {
				transport.WriteJSONError(ctx, 400, "invalid request body")
				return
			}
		}

		resp, err := h.service.Academic(
			ctx.Request.Context(),
			route.Method,
			route.Path,
			body,
			ctx.GetHeader("Content-Type"),
			transport.ForwardHeaders(ctx, "Authorization"),
			ctx.Request.URL.Query(),
		)
		if err != nil {
			transport.WriteJSONError(ctx, 502, "failed to reach profile service")
			return
		}

		transport.WriteProxyResponse(ctx, resp.StatusCode, resp.ContentType, resp.Body)
	}
}
