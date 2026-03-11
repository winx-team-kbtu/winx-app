package api

import (
	"winx-api-gateway/internal/app/modules/auth"
	notification "winx-api-gateway/internal/app/modules/notification"

	"github.com/gin-gonic/gin"
)

var mainRouter *gin.RouterGroup

func (s *Server) initRoutes() error {
	handler = router()

	if err := s.initHealthCheck(); err != nil {
		return err
	}

	s.initDomainRoutes()

	return nil
}

func (s *Server) initDomainRoutes() {
	mainRouter = handler.Group("/api/v1")

	authHandler := auth.NewHandler(s.authService)
	notificationHandler := notification.NewHandler(s.notificationService)

	s.initAuthRoutes(authHandler)
	s.initPasswordRoutes(authHandler)
	s.initNotificationRoutes(notificationHandler)
}

func (s *Server) initAuthRoutes(handler *auth.Handler) {
	authRoutes := mainRouter.Group("")
	authRoutes.POST("/login", handler.Login)
	authRoutes.POST("/register", handler.Register)
	authRoutes.POST("/refresh", handler.Refresh)
	authRoutes.POST("/check", handler.Check)
	authRoutes.POST("/logout", handler.Logout)
}

func (s *Server) initPasswordRoutes(handler *auth.Handler) {
	passwordRoutes := mainRouter.Group("/password")
	passwordRoutes.POST("/forgot", handler.ForgotPassword)
	passwordRoutes.POST("/reset", handler.ResetPassword)
	passwordRoutes.POST("/change", handler.ChangePassword)
	passwordRoutes.POST("/verify-pin", handler.VerifyPin)
}

func (s *Server) initNotificationRoutes(handler *notification.Handler) {
	notificationRoutes := mainRouter.Group("/notifications")
	notificationRoutes.GET("", handler.List)
	notificationRoutes.DELETE("/:id", handler.Delete)
}
