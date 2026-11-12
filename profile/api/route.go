package api

import (
	"net/http"

	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/http/middleware"
	corevalidation "winx-profile/internal/app/core/validation"
	academicHandler "winx-profile/internal/app/domain/handlers/academic"
	choiceHandler "winx-profile/internal/app/domain/handlers/choice"
	courseHandler "winx-profile/internal/app/domain/handlers/course"
	deckHandler "winx-profile/internal/app/domain/handlers/deck"
	interestHandler "winx-profile/internal/app/domain/handlers/interest"
	locationHandler "winx-profile/internal/app/domain/handlers/location"
	photoHandler "winx-profile/internal/app/domain/handlers/photo"
	profileHandler "winx-profile/internal/app/domain/handlers/profile"
	teamHandler "winx-profile/internal/app/domain/handlers/team"
	academicService "winx-profile/internal/app/domain/services/academic"
	choiceService "winx-profile/internal/app/domain/services/choice"
	courseService "winx-profile/internal/app/domain/services/course"
	deckService "winx-profile/internal/app/domain/services/deck"
	interestService "winx-profile/internal/app/domain/services/interest"
	locationService "winx-profile/internal/app/domain/services/location"
	photoService "winx-profile/internal/app/domain/services/photo"
	profileService "winx-profile/internal/app/domain/services/profile"
	teamService "winx-profile/internal/app/domain/services/team"
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
	s.initAcademicRoutes(academicHandler.NewHandler(academicService.NewService(s.pgdb), binder))
	s.initDeckRoutes(deckHandler.NewHandler(deckService.NewService(s.pgdb), binder))
	s.initChoiceRoutes(choiceHandler.NewHandler(choiceService.NewService(s.pgdb), binder))
	s.initTeamRoutes(teamHandler.NewHandler(teamService.NewService(s.pgdb), binder))
	s.initCourseRoutes(courseHandler.NewHandler(courseService.NewService(s.pgdb), binder))

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

func (s *Server) initCourseRoutes(handler *courseHandler.Handler) {
	courseRoutes := authUserRouter.Group("/profile/academic/courses")
	courseRoutes.GET("", handler.List)
	courseRoutes.GET("/me", handler.GetMine)
	courseRoutes.PUT("/me", handler.SaveMine)
}

func (s *Server) initAcademicRoutes(handler *academicHandler.Handler) {
	academicRoutes := authUserRouter.Group("/profile/academic")
	academicRoutes.GET("/me", handler.GetMe)
	academicRoutes.PUT("/me", handler.SaveMe)
}

func (s *Server) initDeckRoutes(handler *deckHandler.Handler) {
	deckRoutes := authUserRouter.Group("/profile/academic")
	deckRoutes.GET("/deck", handler.List)
}

func (s *Server) initChoiceRoutes(handler *choiceHandler.Handler) {
	choiceRoutes := authUserRouter.Group("/profile/academic")
	choiceRoutes.PUT("/choices", handler.Store)
	choiceRoutes.GET("/shortlist", handler.Shortlist)
}

func (s *Server) initTeamRoutes(handler *teamHandler.Handler) {
	teamRoutes := authUserRouter.Group("/profile/academic/teams")
	teamRoutes.POST("", handler.Store)
	teamRoutes.GET("/me", handler.GetMine)
}
