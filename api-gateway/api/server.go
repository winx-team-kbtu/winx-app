package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"winx-api-gateway/configs"
	"winx-api-gateway/internal/app/core/http"
	"winx-api-gateway/internal/app/modules/auth"
	"winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/modules/profile"
	"winx-api-gateway/internal/app/swagger"
	"winx-api-gateway/pkg/graylog/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	authService         auth.Service
	notificationService notification.Service
	profileService      profile.Service
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	logger.LogInfo("Starting API Gateway server initialization")

	s := &Server{}

	if err := s.initDeps(ctx); err != nil {
		logger.LogError("Server initialization failed", err)
		return fmt.Errorf("server initDeps: %w", err)
	}

	logger.LogInfo("API Gateway server started successfully")
	return nil
}

func (s *Server) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		s.initConfig,
		s.initLayers,
		s.initServer,
	}

	for _, f := range inits {
		if err := f(ctx); err != nil {
			logger.LogError("Dependency initialization failed", err)
			return err
		}
	}

	return nil
}

func (s *Server) initConfig(_ context.Context) error {
	configs.InitConfig()
	logger.LogInfo("Configuration loaded successfully")
	return nil
}

func (s *Server) initLayers(_ context.Context) error {
	logger.LogInfo("Initializing service clients")

	authClient := auth.NewClient(
		configs.Config.Services.Auth.URL,
		configs.Config.Services.Auth.APIKey,
		15*time.Second,
	)
	s.authService = auth.NewService(authClient)
	logger.LogInfo("Auth service client initialized")

	notificationClient := notification.NewClient(
		configs.Config.Services.Notification.URL,
		configs.Config.Services.Notification.APIKey,
		15*time.Second,
	)
	s.notificationService = notification.NewService(notificationClient)
	logger.LogInfo("Notification service client initialized")

	profileClient := profile.NewClient(
		configs.Config.Services.Profile.URL,
		configs.Config.Services.Profile.APIKey,
		15*time.Second,
	)
	s.profileService = profile.NewService(profileClient)
	logger.LogInfo("Profile service client initialized")

	return s.initRoutes()
}

func router() *gin.Engine {
	if configs.Config.App.Environment != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// No API key middleware - removed for testing

	return r
}

func (s *Server) initServer(_ context.Context) error {
	httpServer := http.NewHttpServer(handler, http.Port(configs.Config.App.Url))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case serv := <-interrupt:
		logger.LogInfo(fmt.Sprintf("Received signal: %s", serv.String()))
	case err := <-httpServer.Notify():
		logger.LogError("HTTP server error", err)
	}

	if err := httpServer.Shutdown(); err != nil {
		logger.LogError("HTTP server shutdown error", err)
	}

	logger.LogInfo("API Gateway stopped")
	return nil
}

func (s *Server) initHealthCheck() error {
	handler.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": true, "message": "ok"})
	})
	handler.GET("/swagger", swagger.UI)
	handler.GET("/swagger/openapi.yaml", swagger.Spec)
	handler.GET("/swagger/doc.json", swagger.JSONSpec)

	return nil
}