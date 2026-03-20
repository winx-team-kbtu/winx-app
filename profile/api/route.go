package api

import (
	"net/http"

	"winx-profile/internal/app/core/helpers/response"

	"github.com/gin-gonic/gin"
)

func (s *Server) initRoutes() error {
	handler = router()

	handler.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.OK))
	})

	return nil
}
