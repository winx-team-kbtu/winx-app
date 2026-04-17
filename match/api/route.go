package api

import (
	"winx-match/internal/app/core/http/middleware"
	"winx-match/internal/app/core/validation"
	matchHandler "winx-match/internal/app/domain/handlers/match"
	swipeHandler "winx-match/internal/app/domain/handlers/swipe"
	matchService "winx-match/internal/app/domain/services/match"

	"github.com/gin-gonic/gin"
)

func (s *Server) initRoutes() error {
	handler = router()

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	binder := validation.NewBinder(s.validator)

	// s.service is general.Service — used by the swipe handler so it can detect mutual
	// right-swipes and create matches in the same request.
	hSwipe := swipeHandler.NewHandler(s.service, binder)

	// Match CRUD operations use their own focused service.
	hMatch := matchHandler.NewHandler(matchService.NewService(s.db), binder)

	s.initSwipeRoutes(authUserRouter, hSwipe)
	s.initMatchRoutes(authUserRouter, hMatch)

	return nil
}

func (s *Server) initSwipeRoutes(r *gin.RouterGroup, h *swipeHandler.Handler) {
	swipes := r.Group("/swipes")
	swipes.POST("/left", h.SwipeLeft)
	swipes.POST("/right", h.SwipeRight)
}

func (s *Server) initMatchRoutes(r *gin.RouterGroup, h *matchHandler.Handler) {
	matches := r.Group("/matches")
	matches.GET("", h.List)
	matches.DELETE("/:id", h.Delete)
}
