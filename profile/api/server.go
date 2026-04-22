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
	"winx-profile/pkg/mongodb"
	"winx-profile/pkg/postgres"
	"winx-profile/pkg/s3storage"
	"winx-profile/pkg/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

type Server struct {
	pgdb        *gorm.DB
	mongoClient *mongodriver.Client
	mongoDB     *mongodriver.Database
	s3Storage   s3storage.Storage
	rdb         *redis.Client
	cache       cache.Cache
	validator   *validation.Validator
	httpServer  *http.Server
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	configs.InitConfig()
	logger.SetupLogger()

	server, err := newServer(ctx)
	if err != nil {
		return err
	}
	defer server.close()

	return server.run(ctx)
}

func newServer(ctx context.Context) (*Server, error) {
	validator, err := validation.New()
	if err != nil {
		return nil, fmt.Errorf("init validator: %w", err)
	}

	mongoClient, mongoDB, err := mongodb.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("init mongodb: %w", err)
	}

	s3Client, err := s3storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("init s3 storage: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", configs.Config.Redis.Host, configs.Config.Redis.Port),
	})

	s := &Server{
		pgdb:        postgres.NewClient(),
		mongoClient: mongoClient,
		mongoDB:     mongoDB,
		s3Storage:   s3Client,
		rdb:         rdb,
		cache:       cache.NewRedisCache(rdb, "users"),
		validator:   validator,
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

	if s.mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.mongoClient.Disconnect(ctx); err != nil {
			logger.Log.Errorf("close mongodb client: %v", err)
		}
	}

	if s.pgdb != nil {
		sqlDB, err := s.pgdb.DB()
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

	r.Use(middleware.RequestLogger())
	r.Use(middleware.RecoveryWithLogger())

	return r
}
