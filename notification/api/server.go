package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"winx-notification/configs"
	"winx-notification/internal/app/core/http"
	"winx-notification/internal/app/core/http/middleware"
	eventdto "winx-notification/internal/app/domain/core/dto/services/event"
	"winx-notification/internal/app/notifications"
	"winx-notification/pkg/cache"
	"winx-notification/pkg/email"
	"winx-notification/pkg/graylog/logger"
	"winx-notification/pkg/kafka"
	"winx-notification/pkg/postgres"
	"winx-notification/pkg/validation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Server struct {
	db         *gorm.DB
	rdb        *redis.Client
	cache      cache.Cache
	validator  *validation.Validator
	mailer     *email.SMTPMailer
	store      *notifications.Store
	readers    []*kafka.Consumer
	httpServer *http.Server
	groupID    string
	brokers    []string
	topics     kafkaTopics
}

type kafkaTopics struct {
	userRegistered string
	userPassword   string
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
		mailer: email.NewSMTPMailer(
			configs.Config.SMTP.Host,
			configs.Config.SMTP.Port,
			configs.Config.SMTP.Username,
			configs.Config.SMTP.Password,
			configs.Config.SMTP.FromEmail,
			configs.Config.SMTP.FromName,
		),
		groupID: configs.Config.Kafka.GroupID,
		brokers: configs.Config.Kafka.Brokers,
		topics: kafkaTopics{
			userRegistered: configs.Config.Kafka.Topics.UserRegistered,
			userPassword:   configs.Config.Kafka.Topics.UserPassword,
		},
	}
	s.store = notifications.NewStore(db)

	if err := s.initRoutes(); err != nil {
		return nil, fmt.Errorf("init routes: %w", err)
	}

	s.httpServer = http.NewHttpServer(handler, http.Port(configs.Config.App.Url))

	return s, nil
}

func (s *Server) run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 3)

	if err := s.startConsumer(ctx, s.topics.userRegistered, s.handleUserRegistered, errCh); err != nil {
		return err
	}
	if err := s.startConsumer(ctx, s.topics.userPassword, s.handleUserPassword, errCh); err != nil {
		return err
	}

	go s.runDeliveryLoop(ctx)

	logger.Log.Infof(
		"notification listeners started for topics: %s, %s",
		s.topics.userRegistered,
		s.topics.userPassword,
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

	subject := "Welcome to Winx"
	body := fmt.Sprintf(
		"Hi,\n\nYour account has been created successfully on %s.\n\nThanks,\n%s",
		event.CreatedAt.Format("2006-01-02 15:04:05"),
		s.mailer.FromName(),
	)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal user registered payload: %w", err)
	}

	if _, err := s.store.Create(ctx, notifications.CreateInput{
		TypeCode:  notifications.TypeWelcome,
		Recipient: event.Email,
		Subject:   subject,
		Body:      body,
		Payload:   datatypes.JSON(data),
	}); err != nil {
		return fmt.Errorf("store welcome notification: %w", err)
	}

	logger.Log.Infof("welcome notification queued for %s", event.Email)
	return nil
}

func (s *Server) handleUserPassword(ctx context.Context, payload []byte) error {
	var event eventdto.UserPasswordDTO
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode user password event: %w", err)
	}

	subject := "Your Winx password reset PIN"
	body := fmt.Sprintf(
		"Hi,\n\nYour password reset PIN is: %s\n\nThis code was requested at %s.\n\nIf you did not request it, ignore this email.\n\nThanks,\n%s",
		event.PinCode,
		event.CreatedAt.Format("2006-01-02 15:04:05"),
		s.mailer.FromName(),
	)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal user password payload: %w", err)
	}

	if _, err := s.store.Create(ctx, notifications.CreateInput{
		TypeCode:  notifications.TypePasswordReset,
		Recipient: event.Email,
		Subject:   subject,
		Body:      body,
		Payload:   datatypes.JSON(data),
	}); err != nil {
		return fmt.Errorf("store password reset notification: %w", err)
	}

	logger.Log.Infof("password reset notification queued for %s", event.Email)
	return nil
}

func (s *Server) runDeliveryLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.deliverPending(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Log.Errorf("deliver pending notifications: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) deliverPending(ctx context.Context) error {
	items, err := s.store.ClaimPending(ctx, 25)
	if err != nil {
		return err
	}

	for _, item := range items {
		mock := !s.mailer.IsConfigured()
		if !mock {
			if err := s.mailer.Send(ctx, item.Recipient, item.Subject, item.Body); err != nil {
				if markErr := s.store.MarkFailed(ctx, item.ID, err.Error()); markErr != nil {
					logger.Log.Errorf("mark notification failed: %v", markErr)
				}
				continue
			}
		}

		if err := s.store.MarkSent(ctx, item.ID, mock); err != nil {
			logger.Log.Errorf("mark notification sent: %v", err)
			continue
		}

		logger.Log.Infof("notification %d delivered to %s with status %s", item.ID, item.Recipient, sentStatus(mock))
	}

	return nil
}

func sentStatus(mock bool) string {
	if mock {
		return notifications.StatusSentMock
	}

	return notifications.StatusSent
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
