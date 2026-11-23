package api

import (
	"net/http"
	"winx-api-gateway/internal/app/modules/auth"
	"winx-api-gateway/internal/app/modules/chat"
	"winx-api-gateway/internal/app/modules/match"
	notification "winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/modules/profile"
	"winx-api-gateway/internal/app/modules/recommendation"
	"winx-api-gateway/internal/app/proxy"
	"winx-api-gateway/internal/app/transport"

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
	s.initAdminRoutes()
}

func (s *Server) initAdminRoutes() {
	r := mainRouter.Group("/admin")
	forward := func(path string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Header("Cache-Control", "no-store")
			destination := s.adminAuth
			if path == "/admin/records" && c.Query("entity") == "profile_documents" {
				destination = s.adminProfile
			}
			if destination == nil {
				transport.WriteJSONError(c, 503, "Сервис администрирования недоступен.")
				return
			}
			var body []byte
			if c.Request.Method != http.MethodGet {
				c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128<<10)
				var err error
				body, err = transport.ReadBody(c)
				if err != nil {
					transport.WriteJSONError(c, 400, "Неверный формат записи.")
					return
				}
			}
			result, err := destination.Do(c.Request.Context(), proxy.Request{
				Method: c.Request.Method, Path: path, Query: c.Request.URL.Query(), Body: body,
				ContentType: c.GetHeader("Content-Type"), Headers: transport.ForwardHeaders(c, "Authorization"),
			})
			if err != nil {
				transport.WriteJSONError(c, 502, "Сервис администрирования недоступен.")
				return
			}
			transport.WriteProxyResponse(c, result.StatusCode, result.ContentType, result.Body)
		}
	}
	r.GET("/catalog", forward("/admin/catalog"))
	r.GET("/analytics", forward("/admin/analytics"))
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete} {
		r.Handle(method, "/records", forward("/admin/records"))
	}
}

func (s *Server) initAuthRoutes(handler *auth.Handler) {
	authRoutes := mainRouter.Group("")
	authRoutes.POST("/login", handler.Login)
	authRoutes.POST("/register", handler.Register)
	authRoutes.POST("/refresh", handler.Refresh)
	authRoutes.POST("/check", handler.Check)
	authRoutes.POST("/logout", handler.Logout)
	authRoutes.DELETE("/account", handler.DeleteAccount)
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
	profileRoutes.GET("/interests", handler.ListInterests)
	profileRoutes.GET("/location/ip", handler.LookupLocationByIP)

	for _, route := range profile.AcademicRoutes {
		profileRoutes.Handle(route.Method, route.Path, handler.Academic(route))
	}
}

func (s *Server) initMatchRoutes(handler *match.Handler) {
	swipeRoutes := mainRouter.Group("/swipes")
	swipeRoutes.POST("/left", s.social(handler.SwipeLeft))
	swipeRoutes.POST("/right", s.social(handler.SwipeRight))

	matchRoutes := mainRouter.Group("/matches")
	matchRoutes.GET("", s.social(handler.List))
	matchRoutes.DELETE("/:id", s.social(handler.Delete))
}

func (s *Server) initRecommendationRoutes(handler *recommendation.Handler) {
	recRoutes := mainRouter.Group("/recommendations")
	recRoutes.GET("", s.social(handler.List))
}

func (s *Server) initChatRoutes(handler *chat.Handler) {
	chatRoutes := mainRouter.Group("/chats")
	chatRoutes.GET("", s.social(handler.List))
	chatRoutes.GET("/:id/messages", s.social(handler.Messages))
}
