package api

import (
	"winx-chat/internal/app/core/http/middleware"
	"winx-chat/internal/app/core/validation"
	chatHandler "winx-chat/internal/app/domain/handlers/chat"

	"github.com/gin-gonic/gin"
)

func (s *Server) initRoutes() error {
	handler = router()

	// WebSocket endpoint — auth is done via query param, no middleware needed.
	handler.GET("/ws", s.handleWS)

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	binder := validation.NewBinder(s.validator)

	hChat := chatHandler.NewHandler(s.chatService, s.msgService, binder)

	s.initChatRoutes(authUserRouter, hChat)

	return nil
}

func (s *Server) initChatRoutes(r *gin.RouterGroup, h *chatHandler.Handler) {
	chats := r.Group("/chats")
	chats.GET("", h.List)
	chats.GET("/:id/messages", h.Messages)
}
