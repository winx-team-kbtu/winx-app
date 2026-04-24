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
	"winx-api-gateway/internal/app/core/logger"
	"winx-api-gateway/internal/app/modules/auth"
	"winx-api-gateway/internal/app/modules/chat"
	"winx-api-gateway/internal/app/modules/match"
	notification "winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/modules/profile"
	"winx-api-gateway/internal/app/modules/recommendation"
	"winx-api-gateway/internal/app/swagger"
	"winx-api-gateway/pkg/metrics"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	authService           auth.Service
	notificationService   notification.Service
	profileService        profile.Service
	matchService          match.Service
	recommendationService recommendation.Service
	chatService           chat.Service
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	logger.InitLogger(logger.LevelDebug, nil)
	logger.Info("Starting API Gateway initialization...")

	s := &Server{}

	if err := s.initDeps(ctx); err != nil {
		logger.Error("Server initialization failed: %v", err)
		return fmt.Errorf("server initDeps: %w", err)
	}

	logger.Info("API Gateway started successfully")
	return nil
}

func (s *Server) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		s.initConfig,
		s.initLayers,
		s.initServer,
	}

	for i, f := range inits {
		logger.Debug("Running initialization step %d", i)
		if err := f(ctx); err != nil {
			logger.Error("Initialization step %d failed: %v", i, err)
			return err
		}
	}

	return nil
}

func (s *Server) initConfig(_ context.Context) error {
	logger.Info("Loading configuration...")
	configs.InitConfig()
	logger.Debug("Configuration loaded: environment=%s, url=%s",
		configs.Config.App.Environment, configs.Config.App.Url)
	return nil
}

func (s *Server) initLayers(_ context.Context) error {
	logger.Info("Initializing service clients...")

	authClient := auth.NewClient(
		configs.Config.Services.Auth.URL,
		configs.Config.Services.Auth.APIKey,
		15*time.Second,
	)
	s.authService = auth.NewService(authClient)
	logger.Info("Auth service client initialized (URL: %s)", configs.Config.Services.Auth.URL)

	notificationClient := notification.NewClient(
		configs.Config.Services.Notification.URL,
		configs.Config.Services.Notification.APIKey,
		15*time.Second,
	)
	s.notificationService = notification.NewService(notificationClient)
	logger.Info("Notification service client initialized (URL: %s)", configs.Config.Services.Notification.URL)

	profileClient := profile.NewClient(
		configs.Config.Services.Profile.URL,
		configs.Config.Services.Profile.APIKey,
		15*time.Second,
	)
	s.profileService = profile.NewService(profileClient)
	logger.Info("Profile service client initialized (URL: %s)", configs.Config.Services.Profile.URL)

	matchClient := match.NewClient(
		configs.Config.Services.Match.URL,
		configs.Config.Services.Match.APIKey,
		15*time.Second,
	)
	s.matchService = match.NewService(matchClient)
	logger.Info("Match service client initialized (URL: %s)", configs.Config.Services.Match.URL)

	recommendationClient := recommendation.NewClient(
		configs.Config.Services.Recommendation.URL,
		configs.Config.Services.Recommendation.APIKey,
		15*time.Second,
	)
	s.recommendationService = recommendation.NewService(recommendationClient)
	logger.Info("Recommendation service client initialized (URL: %s)", configs.Config.Services.Recommendation.URL)

	chatClient := chat.NewClient(
		configs.Config.Services.Chat.URL,
		configs.Config.Services.Chat.APIKey,
		15*time.Second,
	)
	s.chatService = chat.NewService(chatClient)
	logger.Info("Chat service client initialized (URL: %s)", configs.Config.Services.Chat.URL)

	logger.Debug("Initializing routes...")
	return s.initRoutes()
}

func router() *gin.Engine {
	if configs.Config.App.Environment != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(logger.GinLogger())

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			logger.Debug("CORS request from origin: %s", origin)
			return true
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Requested-With",
			"Accept",
			"Origin",
			"X-CSRF-Token",
			"Cache-Control",
			"Pragma",
			"X-Session-Id",
			"X-api-key",
		},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, err interface{}) {
		logger.Error("Panic recovered: %v, path=%s", err, c.Request.URL.Path)
	}))

	r.Use(metrics.Middleware("api-gateway"))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

func (s *Server) initServer(_ context.Context) error {
	logger.Info("Starting HTTP server on %s", configs.Config.App.Url)
	httpServer := http.NewHttpServer(handler, http.Port(configs.Config.App.Url))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case serv := <-interrupt:
		logger.Info("Received shutdown signal: %s", serv.String())
	case err := <-httpServer.Notify():
		logger.Error("HTTP server error: %v", err)
	}

	logger.Info("Shutting down HTTP server...")
	if err := httpServer.Shutdown(); err != nil {
		logger.Error("HTTP server shutdown error: %v", err)
		return err
	}

	logger.Info("Server shutdown completed")
	return nil
}

func (s *Server) initHealthCheck() error {
	logger.Debug("Initializing health check endpoints")

	handler.GET("/healthz", func(ctx *gin.Context) {
		logger.Debug("Health check requested from %s", ctx.ClientIP())
		ctx.JSON(200, gin.H{"success": true, "message": "ok"})
	})
	handler.GET("/swagger", swagger.UI)
	handler.GET("/swagger/openapi.yaml", swagger.Spec)

	return nil
}
