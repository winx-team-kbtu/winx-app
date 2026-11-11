package api

import (
	"winx-api-gateway/internal/app/transport"

	"github.com/gin-gonic/gin"
)

const socialDisabledMessage = "this feature is not available in the pilot"

// social wraps a handler of a legacy Social feature (photos, swipes, matches,
// recommendations, chat). In pilot mode the real handler is never reached and the
// client gets a 404, so no Social data can be created or read through the gateway.
func (s *Server) social(h gin.HandlerFunc) gin.HandlerFunc {
	return socialHandler(s.pilotMode, h)
}

func socialHandler(pilotMode bool, h gin.HandlerFunc) gin.HandlerFunc {
	if !pilotMode {
		return h
	}

	return func(ctx *gin.Context) {
		transport.WriteJSONError(ctx, 404, socialDisabledMessage)
		ctx.Abort()
	}
}
