package api

import (
	"net/http"
	"winx-notification/internal/app/core/http/middleware"
	notificationHandler "winx-notification/internal/app/domain/handlers/notification"
	notificationService "winx-notification/internal/app/domain/services/notification"

	"github.com/gin-gonic/gin"
)

func (s *Server) initRoutes() error {
	handler = router()

	handler.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "notification service is healthy"})
	})

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	svc := notificationService.NewService(s.db)
	h := notificationHandler.NewHandler(svc)

	notifications := authUserRouter.Group("/notifications")
	notifications.GET("", h.List)
	notifications.DELETE("/:id", h.Delete)

	return nil
}