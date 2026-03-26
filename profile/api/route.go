package api

import (
	"winx-profile/internal/app/core/http/middleware"
	"winx-profile/internal/app/core/validation"
	profileHandler "winx-profile/internal/app/domain/handlers/profile"

	"github.com/gin-gonic/gin"
)

var mainRouter *gin.RouterGroup
var authUserRouter *gin.RouterGroup

func (s *Server) initRoutes() error {
	handler = router()

	s.initDomainRoutes()

	return nil
}

func (s *Server) initDomainRoutes() {
	mainRouter = handler.Group("")
	mainRouter.Use(middleware.ApiKey())

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	authUserRouter = mainRouter.Group("")
	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	binder := validation.NewBinder(s.validator)

	handlerProfile := profileHandler.NewHandler(s.profileService, binder)

	s.initProfileRoutes(handlerProfile)
}

func (s *Server) initProfileRoutes(h *profileHandler.Handler) {
	profileRoutes := authUserRouter.Group("/profile")
	profileRoutes.GET("", h.Get)
	profileRoutes.PUT("", h.Update)

	adminRoutes := authUserRouter.Group("/admin")
	adminRoutes.Use(middleware.AdminOnly(s.db))
	adminRoutes.PUT("/profile/role", h.UpdateRole)
}
