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
	notification "winx-api-gateway/internal/app/modules/notification"
	"winx-api-gateway/internal/app/swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	authService         auth.Service
	notificationService notification.Service
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	s := &Server{}

	if err := s.initDeps(ctx); err != nil {
		return fmt.Errorf("server initDeps: %w", err)
	}

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
			return err
		}
	}

	return nil
}

func (s *Server) initConfig(_ context.Context) error {
	configs.InitConfig()
	return nil
}

func (s *Server) initLayers(_ context.Context) error {
	authClient := auth.NewClient(
		configs.Config.Services.Auth.URL,
		configs.Config.Services.Auth.APIKey,
		15*time.Second,
	)
	s.authService = auth.NewService(authClient)

	notificationClient := notification.NewClient(
		configs.Config.Services.Notification.URL,
		configs.Config.Services.Notification.APIKey,
		15*time.Second,
	)
	s.notificationService = notification.NewService(notificationClient)

	return s.initRoutes()
}

func router() *gin.Engine {
	if configs.Config.App.Environment != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

	return r
}

func (s *Server) initServer(_ context.Context) error {
	httpServer := http.NewHttpServer(handler, http.Port(configs.Config.App.Url))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case serv := <-interrupt:
		fmt.Println("winx-api-gateway - started - signal: " + serv.String())
	case err := <-httpServer.Notify():
		fmt.Println(fmt.Errorf("winx-api-gateway - httpServer.Notify: %w", err))
	}

	if err := httpServer.Shutdown(); err != nil {
		fmt.Println(fmt.Errorf("winx-api-gateway - httpServer.Shutdown: %w", err))
	}

	return nil
}

func (s *Server) initHealthCheck() error {
	handler.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": true, "message": "ok"})
	})
	handler.GET("/swagger", swagger.UI)
	handler.GET("/swagger/openapi.yaml", swagger.Spec)

	return nil
}
