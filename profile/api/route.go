package api

import (
	"net/http"

	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/http/middleware"
	corevalidation "winx-profile/internal/app/core/validation"
	interestHandler "winx-profile/internal/app/domain/handlers/interest"
	locationHandler "winx-profile/internal/app/domain/handlers/location"
	photoHandler "winx-profile/internal/app/domain/handlers/photo"
	profileHandler "winx-profile/internal/app/domain/handlers/profile"
	interestService "winx-profile/internal/app/domain/services/interest"
	locationService "winx-profile/internal/app/domain/services/location"
	photoService "winx-profile/internal/app/domain/services/photo"
	profileService "winx-profile/internal/app/domain/services/profile"
	"winx-profile/pkg/cache"
	"winx-profile/pkg/geoip"

	"github.com/gin-gonic/gin"
)

var mainRouter *gin.RouterGroup
var authUserRouter *gin.RouterGroup

func (s *Server) initRoutes() error {
	handler = router()

	if err := s.initHealthCheck(); err != nil {
		return err
	}

	return s.initDomainRoutes()
}

func (s *Server) initHealthCheck() error {
	handler.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.OK))
	})

	return nil
}

func (s *Server) initDomainRoutes() error {
	mainRouter = handler.Group("")
	mainRouter.Use(middleware.ApiKey())
	authUserRouter = mainRouter.Group("")

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)
	binder := corevalidation.NewBinder(s.validator)

	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	profiles, err := profileService.NewService(s.mongoDB, s.pgdb)
	if err != nil {
		return err
	}

	photos := photoService.NewService(s.pgdb, s.s3Storage)
	interests := interestService.NewService(s.pgdb)
	geoipCache := cache.NewRedisCache(s.rdb, "geoip")
	locations := locationService.NewService(geoip.NewCachedClient(geoip.NewClient(), geoipCache))

	s.initProfileRoutes(profileHandler.NewHandler(profiles, binder))
	s.initPhotoRoutes(photoHandler.NewHandler(photos))
	s.initInterestRoutes(interestHandler.NewHandler(interests))
	s.initLocationRoutes(locationHandler.NewHandler(locations))

	return nil
}

func (s *Server) initProfileRoutes(handler *profileHandler.Handler) {
	profileRoutes := authUserRouter.Group("/profile")
	profileRoutes.GET("/me", handler.GetMe)
	profileRoutes.POST("/store", handler.Store)
}

func (s *Server) initPhotoRoutes(handler *photoHandler.Handler) {
	profileRoutes := authUserRouter.Group("/profile")
	profileRoutes.GET("/photo", handler.GetMe)
	profileRoutes.POST("/photo/store", handler.Store)
}

func (s *Server) initInterestRoutes(handler *interestHandler.Handler) {
	profileRoutes := authUserRouter.Group("/profile")
	profileRoutes.GET("/interests", handler.List)
}

func (s *Server) initLocationRoutes(handler *locationHandler.Handler) {
	profileRoutes := authUserRouter.Group("/profile")
	profileRoutes.GET("/location/ip", handler.LookupByIP)
}
