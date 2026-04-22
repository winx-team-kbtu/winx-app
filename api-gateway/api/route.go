package api

import (
	"winx-api-gateway/internal/app/modules/auth"
	"winx-api-gateway/internal/app/modules/chat"
	"winx-api-gateway/internal/app/modules/match"
	notification "winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/modules/profile"
	"winx-api-gateway/internal/app/modules/recommendation"

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
	profileHandler := profile.NewHandler(s.profileService)
	matchHandler := match.NewHandler(s.matchService)
	recommendationHandler := recommendation.NewHandler(s.recommendationService)
	chatHandler := chat.NewHandler(s.chatService)

	s.initAuthRoutes(authHandler)
	s.initPasswordRoutes(authHandler)
	s.initNotificationRoutes(notificationHandler)
	s.initProfileRoutes(profileHandler)
	s.initMatchRoutes(matchHandler)
	s.initRecommendationRoutes(recommendationHandler)
	s.initChatRoutes(chatHandler)
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

func (s *Server) initProfileRoutes(handler *profile.Handler) {
	profileRoutes := mainRouter.Group("/profile")
	profileRoutes.GET("/me", handler.GetMe)
	profileRoutes.POST("/store", handler.Store)
	profileRoutes.GET("/photo", handler.GetPhotos)
	profileRoutes.POST("/photo/store", handler.StorePhoto)
	profileRoutes.GET("/interests", handler.ListInterests)
	profileRoutes.GET("/location/ip", handler.LookupLocationByIP)
}

func (s *Server) initMatchRoutes(handler *match.Handler) {
	swipeRoutes := mainRouter.Group("/swipes")
	swipeRoutes.POST("/left", handler.SwipeLeft)
	swipeRoutes.POST("/right", handler.SwipeRight)

	matchRoutes := mainRouter.Group("/matches")
	matchRoutes.GET("", handler.List)
	matchRoutes.DELETE("/:id", handler.Delete)
}

func (s *Server) initRecommendationRoutes(handler *recommendation.Handler) {
	recRoutes := mainRouter.Group("/recommendations")
	recRoutes.GET("", handler.List)
}

func (s *Server) initChatRoutes(handler *chat.Handler) {
	chatRoutes := mainRouter.Group("/chats")
	chatRoutes.GET("", handler.List)
	chatRoutes.GET("/:id/messages", handler.Messages)
}
