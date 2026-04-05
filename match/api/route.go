package api

import (
	"net/http"
	"winx-match/internal/app/core/http/middleware"
	"winx-match/internal/app/core/validation"
	matchHandler "winx-match/internal/app/domain/handlers/match"
	swipeHandler "winx-match/internal/app/domain/handlers/swipe"
	swipeService "winx-match/internal/app/domain/services/swipe"

	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------
// initRoutes — точка входа для всех маршрутов сервиса.
//
// Структура роутов:
//
//	mainRouter       — защищён middleware ApiKey (X-api-key header)
//	authUserRouter   — дополнительно требует валидного Bearer-токена
//
// Чтобы добавить новую группу эндпоинтов:
//  1. Создай handler    в internal/app/domain/handlers/<name>/handler.go
//  2. Создай service    в internal/app/domain/services/<name>/service.go
//  3. Создай repository в internal/app/domain/repositories/<name>/repository.go
//  4. Создай DTO        в internal/app/domain/core/dto/{requests,services,repositories}/<name>/dto.go
//  5. Зарегистрируй маршруты здесь по аналогии с initSwipeRoutes / initMatchRoutes
//
// -----------------------------------------------------------------------
func (s *Server) initRoutes() error {
	handler = router()

	handler.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "match service is healthy"})
	})

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	binder := validation.NewBinder(s.validator)

	svcSwipe := swipeService.NewService(s.db)
	hSwipe := swipeHandler.NewHandler(svcSwipe, binder)

	//hMatch := matchHandler.NewHandler(s.service.MatchService, binder)

	s.initSwipeRoutes(authUserRouter, hSwipe)
	//s.initMatchRoutes(authUserRouter, hMatch)

	return nil
}

// initSwipeRoutes регистрирует эндпоинты для свайпов.
// Файл для расширения: internal/app/domain/handlers/swipe/handler.go
func (s *Server) initSwipeRoutes(r *gin.RouterGroup, h *swipeHandler.Handler) {
	swipes := r.Group("/swipes")
	swipes.POST("/left", h.SwipeLeft)
	swipes.POST("/right", h.SwipeRight)
}

// initMatchRoutes регистрирует эндпоинты для матчей.
// Файл для расширения: internal/app/domain/handlers/match/handler.go
func (s *Server) initMatchRoutes(r *gin.RouterGroup, h *matchHandler.Handler) {
	matches := r.Group("/matches")
	matches.GET("", h.List)
	matches.DELETE("/:id", h.Delete)
}
