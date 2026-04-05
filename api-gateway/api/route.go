package api

import (
	"winx-api-gateway/internal/app/modules/auth"
	"winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/modules/profile"

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
	mainRouter = handler.Group("")

	authHandler := auth.NewHandler(s.authService)
	notificationHandler := notification.NewHandler(s.notificationService)
	profileHandler := profile.NewHandler(s.profileService)

	s.initAuthRoutes(authHandler)
	s.initUserRoutes(authHandler)
	s.initPasswordRoutes(authHandler)
	s.initNotificationRoutes(notificationHandler)
	s.initProfileRoutes(profileHandler)
}

func (s *Server) initAuthRoutes(handler *auth.Handler) {
	authRoutes := mainRouter.Group("")
	authRoutes.POST("/login", handler.Login)
	authRoutes.POST("/register", handler.Register)
	authRoutes.POST("/refresh", handler.Refresh)
	authRoutes.POST("/check", handler.Check)
	authRoutes.POST("/logout", handler.Logout)
}

func (s *Server) initUserRoutes(handler *auth.Handler) {
	userRoutes := mainRouter.Group("/user")
	userRoutes.POST("/store", handler.CreateUser)
	userRoutes.DELETE("/delete", handler.DeleteUser)
	userRoutes.PUT("/update", handler.UpdateUser)
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

func (s *Server) initProfileRoutes(handler *profile.Handler) {
	profileRoutes := mainRouter.Group("/profile")
	profileRoutes.GET("/me", handler.GetMe)
	profileRoutes.POST("/store", handler.Store)
	profileRoutes.GET("/photo", handler.GetPhoto)
	profileRoutes.POST("/photo/store", handler.StorePhoto)
	profileRoutes.GET("/interests", handler.ListInterests)
	profileRoutes.GET("/location/ip", handler.LookupLocation)
}