package api

import (
	"winx-recommendation/internal/app/core/http/middleware"
	"winx-recommendation/internal/app/core/validation"
	recommendationHandler "winx-recommendation/internal/app/domain/handlers/recommendation"
	recommendationService "winx-recommendation/internal/app/domain/services/recommendation"

	"github.com/gin-gonic/gin"
)

// initRoutes wires all HTTP routes for the recommendation service.
//
// Route groups:
//
//	mainRouter     — requires valid X-api-key header
//	authUserRouter — additionally requires a valid Bearer token (Redis-cached)
//
// To add new endpoints:
//  1. Create handler in internal/app/domain/handlers/<name>/handler.go
//  2. Create service in internal/app/domain/services/<name>/service.go
//  3. Create repository in internal/app/domain/repositories/<name>/repository.go
//  4. Register routes here following the initRecommendationRoutes pattern
func (s *Server) initRoutes() error {
	handler = router()

	mainRouter := handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter := mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	binder := validation.NewBinder(s.validator)

	svcRecommendation := recommendationService.NewService(s.db, s.mongoDB, s.cache)
	hRecommendation := recommendationHandler.NewHandler(svcRecommendation, binder)

	s.initRecommendationRoutes(authUserRouter, hRecommendation)

	return nil
}

// initRecommendationRoutes registers recommendation endpoints.
func (s *Server) initRecommendationRoutes(r *gin.RouterGroup, h *recommendationHandler.Handler) {
	recs := r.Group("/recommendations")
	recs.GET("", h.List)
}
