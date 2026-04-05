package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"winx-match/configs"
	apphttp "winx-match/internal/app/core/http"
	"winx-match/internal/app/core/http/middleware"
	eventdto "winx-match/internal/app/domain/core/dto/services/event"
	service "winx-match/internal/app/domain/services/general"
	"winx-match/pkg/cache"
	"winx-match/pkg/graylog/logger"
	"winx-match/pkg/kafka"
	"winx-match/pkg/postgres"
	"winx-match/pkg/validation"

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
	service     service.Service
	readers     []*kafka.Consumer
	matchWriter *kafka.Producer
	httpServer  *apphttp.Server
	groupID     string
	brokers     []string
	topics      kafkaTopics
}

type kafkaTopics struct {
	swipeLeft    string
	swipeRight   string
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
		groupID:   configs.Config.Kafka.GroupID,
		brokers:   configs.Config.Kafka.Brokers,
		topics: kafkaTopics{
			swipeLeft:    configs.Config.Kafka.Topics.SwipeLeft,
			swipeRight:   configs.Config.Kafka.Topics.SwipeRight,
			matchCreated: configs.Config.Kafka.Topics.MatchCreated,
		},
	}

	s.service = service.NewService(db)

	if err := s.initRoutes(); err != nil {
		return nil, fmt.Errorf("init routes: %w", err)
	}

	s.httpServer = apphttp.NewHttpServer(handler, apphttp.Port(configs.Config.App.Url))

	return s, nil
}

func (s *Server) run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 3)

	if err := s.startConsumer(ctx, s.topics.swipeLeft, s.handleSwipeLeft, errCh); err != nil {
		return err
	}
	if err := s.startConsumer(ctx, s.topics.swipeRight, s.handleSwipeRight, errCh); err != nil {
		return err
	}

	producer, err := kafka.NewProducer(s.brokers)
	if err != nil {
		return fmt.Errorf("create kafka producer for %s: %w", s.topics.matchCreated, err)
	}
	s.matchWriter = producer

	logger.Log.Infof(
		"match listeners started for topics: %s, %s",
		s.topics.swipeLeft,
		s.topics.swipeRight,
	)

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
			"Authorization", "Content-Type", "X-Requested-With", "Accept",
			"Origin", "X-CSRF-Token", "Cache-Control", "Pragma",
			"X-Session-Id", "X-api-key", "X-User-Id", "X-User-Email",
		},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
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

func (s *Server) handleSwipeLeft(ctx context.Context, payload []byte) error {
	var event eventdto.SwipeEventDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode swipe.left event: %w", err)
	}

	if err := s.service.HandleSwipeLeft(ctx, event.SwiperID, event.SwipedID); err != nil {
		return fmt.Errorf("handle swipe left: %w", err)
	}

	logger.Log.Infof("swipe.left: user %d swiped left on %d", event.SwiperID, event.SwipedID)
	return nil
}

func (s *Server) handleSwipeRight(ctx context.Context, payload []byte) error {
	var event eventdto.SwipeEventDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode swipe.right event: %w", err)
	}

	logger.Log.Infof("swipe.right: user %d swiped right on %d", event.SwiperID, event.SwipedID)

	match, created, err := s.service.HandleSwipeRight(ctx, event.SwiperID, event.SwipedID)
	if err != nil {
		return fmt.Errorf("try create match: %w", err)
	}

	if created {
		if err := s.publishMatch(ctx, eventdto.MatchCreatedDTO{
			UserOneID: event.SwiperID,
			UserTwoID: event.SwipedID,
			CreatedAt: time.Now(),
			MatchID:   match.ID,
		}); err != nil {
			logger.Log.Errorf("publish match event: %v", err)
			return err
		}
	}

	return nil
}

func (s *Server) publishMatch(ctx context.Context, match eventdto.MatchCreatedDTO) error {
	payload, err := json.Marshal(match)
	if err != nil {
		return fmt.Errorf("marshal match event: %w", err)
	}
	if err := s.matchWriter.Publish(ctx, s.topics.matchCreated, "", payload); err != nil {
		return fmt.Errorf("publish match event: %w", err)
	}
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

	if s.matchWriter != nil {
		if err := s.matchWriter.Close(); err != nil {
			logger.Log.Errorf("close kafka match producer: %v", err)
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