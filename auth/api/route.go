package api

import (
	"time"

	profileclient "auth/adapters/profile-client"
	"auth/configs"
	"auth/internal/app/core/http/middleware"
	"auth/internal/app/core/validation"
	authHandler "auth/internal/app/domain/handlers"
	accountHandler "auth/internal/app/domain/handlers/account"
	passwordHandler "auth/internal/app/domain/handlers/password"
	roleHandler "auth/internal/app/domain/handlers/role"
	userHandler "auth/internal/app/domain/handlers/user"
	deletionRepository "auth/internal/app/domain/repositories/deletion"
	authService "auth/internal/app/domain/services"
	deletionService "auth/internal/app/domain/services/deletion"
	passwordService "auth/internal/app/domain/services/password"
	roleService "auth/internal/app/domain/services/role"
	tokenService "auth/internal/app/domain/services/token"
	userService "auth/internal/app/domain/services/user"
	uservalidationservice "auth/internal/app/domain/validation-services/user"

	"github.com/gin-gonic/gin"
)

var mainRouter *gin.RouterGroup
var authUserRouter *gin.RouterGroup
var adminRouter *gin.RouterGroup

func (s *Server) initRoutes() error {
	handler = router()

	if err := s.initHealthCheck(); err != nil {
		return err
	}

	s.initDomainRoutes()

	return nil
}

func (s *Server) initHealthCheck() error {
	return nil
}

func (s *Server) initDomainRoutes() {
	mainRouter = handler.Group("")
	mainRouter.Use(middleware.ApiKey())
	authUserRouter = mainRouter.Group("")

	authUserMiddleware := middleware.NewAuthUserMiddleware(s.cache)

	binder := validation.NewBinder(s.validator)

	authUserRouter.Use(authUserMiddleware.AuthUser(), authUserMiddleware.ContextWithAuthUser())

	adminRouter = authUserRouter.Group("")
	adminRouter.Use(middleware.AdminOnly(s.pgdb))

	serviceUser := userService.NewService(s.pgdb)
	userValidationService := uservalidationservice.New(s.dbValidator)
	handlerUser := userHandler.NewHandler(serviceUser, binder, userValidationService)
	serviceToken := tokenService.NewService(s.oauthServer)
	serviceAuth := authService.NewService(
		s.pgdb,
		s.cache,
		serviceToken,
		s.tokenStore,
		serviceUser,
		s.kafka,
	)
	handlerAuth := authHandler.NewHandler(serviceAuth, binder, userValidationService)
	servicePassword := passwordService.NewService(s.pgdb, serviceToken, s.kafka)
	handlerPassword := passwordHandler.NewHandler(servicePassword, binder)
	svcRole := roleService.NewService(s.pgdb)
	handlerRole := roleHandler.NewHandler(svcRole, binder)

	accountService := deletionService.NewService(
		deletionRepository.NewRepository(s.pgdb),
		deletionRepository.NewRepository(s.pgdb),
		s.cache,
		profileclient.New(configs.Config.Services.Profile.URL, configs.Config.Services.Profile.APIKey, 30*time.Second),
	)
	handlerAccount := accountHandler.NewHandler(accountService, binder)

	s.initAuthRoutes(handlerAuth, authUserMiddleware)
	s.initAccountRoutes(handlerAccount)
	s.initPasswordRoutes(handlerPassword)
	s.initUserRoutes(handlerUser)
	s.initRoleRoutes(handlerRole)
}

func (s *Server) initUserRoutes(handler *userHandler.Handler) {
	userRoutes := adminRouter.Group("/user")
	userRoutes.POST("/store", handler.Create)
	userRoutes.DELETE("/delete", handler.Delete)
	userRoutes.PUT("/update", handler.Update)
}

func (s *Server) initAuthRoutes(handler *authHandler.Handler, authUserMiddleware *middleware.AuthUserMiddleware) {
	authRoutes := mainRouter.Group("")
	authRoutes.POST("/register", handler.Register)
	authRoutes.POST("/login", handler.Login)
	authRoutes.POST("/refresh", handler.RefreshToken)
	authRoutes.POST("/check", authUserMiddleware.AuthUser(), handler.CheckToken)

	authUserRoutes := authUserRouter.Group("")
	authUserRoutes.POST("/logout", handler.Logout)
}

func (s *Server) initRoleRoutes(handler *roleHandler.Handler) {
	roleRoutes := adminRouter.Group("/roles")
	roleRoutes.POST("", handler.Store)
	roleRoutes.PUT("/:id", handler.Update)
	roleRoutes.DELETE("/:id", handler.Delete)
}

func (s *Server) initPasswordRoutes(handler *passwordHandler.Handler) {
	passwordRoutes := mainRouter.Group("/password")
	passwordRoutes.POST("/forgot", handler.ForgotPassword)
	passwordRoutes.POST("/reset", handler.ResetPassword)
	passwordRoutes.POST("/change", handler.ChangePassword)
	passwordRoutes.POST("/verify-pin", handler.VerifyPin)
}

func (s *Server) initAccountRoutes(handler *accountHandler.Handler) {
	authUserRouter.DELETE("/account", handler.Delete)
}
