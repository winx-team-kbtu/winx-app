package api

import (
	"winx-notification/internal/app/core/http/middleware"
	notificationHandler "winx-notification/internal/app/domain/handlers/notification"
	notificationService "winx-notification/internal/app/domain/services/notification"
)

func (s *Server) initRoutes() error {
	handler = router()

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	notifications := authUserRouter.Group("/notifications")
	service := notificationService.NewService(s.db)
	h := notificationHandler.NewHandler(service)

	notifications.GET("", h.List)
	notifications.DELETE("/:id", h.Delete)

	return nil
}
