package api

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"winx-profile/configs"
	"winx-profile/internal/app/core/http"
	"winx-profile/internal/app/core/http/middleware"
	"winx-profile/pkg/cache"
	"winx-profile/pkg/graylog/logger"
	"winx-profile/pkg/postgres"
	"winx-profile/pkg/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	db         *gorm.DB
	rdb        *redis.Client
	cache      cache.Cache
	validator  *validation.Validator
	httpServer *http.Server
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	configs.InitConfig()
	logger.SetupLogger()

	server, err := newServer()
	if err != nil {
		return err
	}
	defer server.close()

	return server.run(ctx)
}

func newServer() (*Server, error) {
	validator, err := validation.New()
	if err != nil {
		return nil, fmt.Errorf("init validator: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", configs.Config.Redis.Host, configs.Config.Redis.Port),
	})

	s := &Server{
		db:        postgres.NewClient(),
		rdb:       rdb,
		cache:     cache.NewRedisCache(rdb, "users"),
		validator: validator,
	}

	if err := s.initRoutes(); err != nil {
		return nil, fmt.Errorf("init routes: %w", err)
	}

	s.httpServer = http.NewHttpServer(handler, http.Port(configs.Config.App.Url))

	return s, nil
}

func (s *Server) run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		return nil
	case err := <-s.httpServer.Notify():
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}

		return fmt.Errorf("http server: %w", err)
	}
}

func (s *Server) close() {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(); err != nil && !errors.Is(err, context.Canceled) {
			logger.Log.Errorf("shutdown http server: %v", err)
		}
	}

	if s.rdb != nil {
		if err := s.rdb.Close(); err != nil {
			logger.Log.Errorf("close redis client: %v", err)
		}
	}

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			if cerr := sqlDB.Close(); cerr != nil {
				logger.Log.Errorf("close postgres client: %v", cerr)
			}
		}
	}
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
			"X-User-Id",
			"X-User-Email",
		},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(middleware.RecoveryWithLogger())

	return r
}
