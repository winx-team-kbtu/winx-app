package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"winx-chat/configs"
	apphttp "winx-chat/internal/app/core/http"
	"winx-chat/internal/app/core/http/middleware"
	eventdto "winx-chat/internal/app/domain/core/dto/services/event"
	chatService "winx-chat/internal/app/domain/services/chat"
	msgService "winx-chat/internal/app/domain/services/message"
	"winx-chat/pkg/cache"
	"winx-chat/pkg/graylog/logger"
	"winx-chat/pkg/kafka"
	"winx-chat/pkg/postgres"
	"winx-chat/pkg/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	db          *gorm.DB
	rdb         *redis.Client
	cache       cache.Cache
	validator   *validation.Validator
	chatService chatService.Service
	msgService  msgService.Service
	hub         *Hub
	readers     []*kafka.Consumer
	httpServer  *apphttp.Server
	groupID     string
	brokers     []string
	topics      kafkaTopics
}

type kafkaTopics struct {
	matchCreated string
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
	db := postgres.NewClient()

	validator, err := validation.New()
	if err != nil {
		return nil, fmt.Errorf("init validator: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", configs.Config.Redis.Host, configs.Config.Redis.Port),
	})

	s := &Server{
		db:        db,
		rdb:       rdb,
		cache:     cache.NewRedisCache(rdb, "users"),
		validator: validator,
		hub:       NewHub(),
		groupID:   configs.Config.Kafka.GroupID,
		brokers:   configs.Config.Kafka.Brokers,
		topics: kafkaTopics{
			matchCreated: configs.Config.Kafka.Topics.MatchCreated,
		},
	}

	s.chatService = chatService.NewService(db)
	s.msgService = msgService.NewService(db)

	if err := s.initRoutes(); err != nil {
		return nil, fmt.Errorf("init routes: %w", err)
	}

	s.httpServer = apphttp.NewHttpServer(handler, apphttp.Port(configs.Config.App.Url))

	return s, nil
}

func (s *Server) run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)

	// -----------------------------------------------------------------------
	// Kafka consumers
	// -----------------------------------------------------------------------
	if err := s.startConsumer(ctx, s.topics.matchCreated, s.handleMatchCreated, errCh); err != nil {
		return err
	}

	logger.Log.Infof("chat listener started for topic: %s", s.topics.matchCreated)

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

func (s *Server) startConsumer(
	ctx context.Context,
	topic string,
	h func(context.Context, []byte) error,
	errCh chan<- error,
) error {
	consumer, err := kafka.NewConsumer(s.brokers, topic, s.groupID)
	if err != nil {
		return fmt.Errorf("create kafka consumer for %s: %w", topic, err)
	}

	s.readers = append(s.readers, consumer)

	go func() {
		errCh <- consumer.Consume(ctx, h)
	}()

	return nil
}

// -----------------------------------------------------------------------
// Kafka event handlers
// -----------------------------------------------------------------------

func (s *Server) handleMatchCreated(ctx context.Context, payload []byte) error {
	var event eventdto.MatchCreatedDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode match.created event: %w", err)
	}

	chat, err := s.chatService.CreateFromMatch(ctx, event.MatchID, event.UserOneID, event.UserTwoID)
	if err != nil {
		return fmt.Errorf("create chat from match: %w", err)
	}

	logger.Log.Infof(
		"chat created: id=%d match=%d users=(%d, %d)",
		chat.ID, event.MatchID, event.UserOneID, event.UserTwoID,
	)

	return nil
}

func (s *Server) close() {
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown()
	}

	for _, reader := range s.readers {
		if err := reader.Close(); err != nil {
			logger.Log.Errorf("close kafka consumer: %v", err)
		}
	}

	if s.rdb != nil {
		_ = s.rdb.Close()
	}

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}
