package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"winx-profile/configs"
	"winx-profile/internal/app/core/http"
	"winx-profile/internal/app/core/http/middleware"
	eventdto "winx-profile/internal/app/domain/core/dto/services/event"
	serviceDto "winx-profile/internal/app/domain/core/dto/services/profile"
	profileService "winx-profile/internal/app/domain/services/profile"
	"winx-profile/pkg/cache"
	"winx-profile/pkg/graylog/logger"
	"winx-profile/pkg/kafka"
	"winx-profile/pkg/postgres"
	"winx-profile/pkg/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	db             *gorm.DB
	rdb            *redis.Client
	cache          cache.Cache
	validator      *validation.Validator
	profileService profileService.Service
	readers        []*kafka.Consumer
	httpServer     *http.Server
	groupID        string
	brokers        []string
	topicUserReg   string
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

	db := postgres.NewClient()

	s := &Server{
		db:             db,
		rdb:            rdb,
		cache:          cache.NewRedisCache(rdb, "users"),
		validator:      validator,
		profileService: profileService.NewService(db),
		groupID:        configs.Config.Kafka.GroupID,
		brokers:        configs.Config.Kafka.Brokers,
		topicUserReg:   configs.Config.Kafka.Topics.UserRegistered,
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

	errCh := make(chan error, 2)

	if err := s.startConsumer(ctx, s.topicUserReg, s.handleUserRegistered, errCh); err != nil {
		return err
	}

	logger.Log.Infof("profile listener started for topic: %s", s.topicUserReg)

	select {
	case <-ctx.Done():
		return nil
	case err := <-s.httpServer.Notify():
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

func (s *Server) startConsumer(
	ctx context.Context,
	topic string,
	handler func(context.Context, []byte) error,
	errCh chan<- error,
) error {
	consumer, err := kafka.NewConsumer(s.brokers, topic, s.groupID)
	if err != nil {
		return fmt.Errorf("create kafka consumer for %s: %w", topic, err)
	}

	s.readers = append(s.readers, consumer)

	go func() {
		errCh <- consumer.Consume(ctx, handler)
	}()

	return nil
}

func (s *Server) handleUserRegistered(ctx context.Context, payload []byte) error {
	var event eventdto.UserRegisteredDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode user registered event: %w", err)
	}

	_, err := s.profileService.Create(ctx, serviceDto.CreateDTO{
		UserID: event.UserID,
	})
	if err != nil {
		return fmt.Errorf("create profile for user %d: %w", event.UserID, err)
	}

	logger.Log.Infof("profile auto-created for user_id=%d", event.UserID)
	return nil
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

func (s *Server) close() {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(); err != nil && !errors.Is(err, context.Canceled) {
			logger.Log.Errorf("shutdown http server: %v", err)
		}
	}

	for _, reader := range s.readers {
		if err := reader.Close(); err != nil {
			logger.Log.Errorf("close kafka consumer: %v", err)
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
