package api

import (
	"auth/configs"
	kafkacontract "auth/internal/app/core/contracts/microservices/kafka-contract"
	"auth/internal/app/core/helpers/errorhandler"
	oauthHelper "auth/internal/app/core/helpers/oauth"
	"auth/internal/app/core/helpers/token"
	"auth/internal/app/core/http"
	"auth/internal/app/core/http/middleware"
	dbvalidator "auth/internal/app/core/validation"
	gormvalidator "auth/internal/app/core/validation/gorm"
	oauthtoken "auth/internal/app/domain/services/token"
	"auth/pkg/cache"
	"auth/pkg/graylog/logger"
	kafka "auth/pkg/kafka"
	"auth/pkg/postgres"
	"auth/pkg/validation"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	gormTokenStore "auth/adapters/gorm-token-store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	oauthModels "github.com/go-oauth2/oauth2/v4/models"
	oauthServer "github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	pgdb        *gorm.DB
	rdb         *redis.Client
	cache       cache.Cache
	kafka       kafkacontract.Producer
	validator   *validation.Validator
	tokenStore  oauth2.TokenStore
	dbValidator dbvalidator.DBValidator
	oauthServer oauthtoken.OAuthServer
}

var handler *gin.Engine

func NewServer(ctx context.Context) error {
	s := &Server{}

	if err := s.initDeps(ctx); err != nil {
		errorhandler.FailOnError(err, "проблема с инициализацией зависимостей")

		return fmt.Errorf("server initDeps: %w", err)
	}

	return nil
}

func (s *Server) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		s.initConfig,
		s.setupLogger,
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

func (s *Server) setupLogger(_ context.Context) error {
	logger.SetupLogger()

	return nil
}

func (s *Server) initLayers(_ context.Context) error {
	s.pgdb = postgres.NewClient()

	validator, err := validation.New()
	if err != nil {
		return fmt.Errorf("failed to init validator: %w", err)
	}
	s.dbValidator = gormvalidator.New(s.pgdb)

	s.validator = validator

	s.rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", configs.Config.Redis.Host, configs.Config.Redis.Port),
	})
	s.cache = cache.NewRedisCache(s.rdb, "users")
	s.kafka, err = kafka.NewProducer(configs.Config.Kafka.Brokers)
	if err != nil {
		return fmt.Errorf("failed to init kafka producer: %w", err)
	}

	s.tokenStore = gormTokenStore.NewGormTokenStore(s.pgdb)

	manager := s.initManager()
	passwdHandler := oauthHelper.NewPasswordHandler(s.pgdb)
	s.oauthServer = s.initOAuthServer(manager, passwdHandler.Password)

	return s.initRoutes()
}

func (s *Server) initTestLayers(_ context.Context) error {
	s.pgdb = postgres.NewClient()

	validator, err := validation.New()
	if err != nil {
		return fmt.Errorf("failed to init validator: %w", err)
	}
	s.dbValidator = gormvalidator.New(s.pgdb)

	s.validator = validator

	s.rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", configs.Config.Redis.Host, configs.Config.Redis.Port),
	})
	s.cache = cache.NewRedisCache(s.rdb, "users")
	s.kafka, err = kafka.NewProducer(configs.Config.Kafka.Brokers)
	if err != nil {
		return fmt.Errorf("failed to init kafka producer: %w", err)
	}

	return s.initRoutes()
}

func NewTestEngine(ctx context.Context,
	oauth oauthtoken.OAuthServer,
	tokenStore oauth2.TokenStore,
) (*gin.Engine, *gorm.DB, error) {
	s := &Server{}
	s.oauthServer = oauth
	s.tokenStore = tokenStore

	if err := s.initConfig(ctx); err != nil {
		return nil, nil, fmt.Errorf("init config: %w", err)
	}

	if err := s.initTestLayers(ctx); err != nil {
		return nil, nil, fmt.Errorf("init layers: %w", err)
	}

	return handler, s.pgdb, nil
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

	r.Use(middleware.RecoveryWithLogger())

	return r
}

func (s *Server) initOAuthServer(manager oauth2.Manager, passwordHandler oauthServer.PasswordAuthorizationHandler) *oauthServer.Server {
	srv := oauthServer.NewDefaultServer(manager)
	srv.SetClientInfoHandler(oauthServer.ClientFormHandler)
	srv.SetPasswordAuthorizationHandler(passwordHandler)

	return srv
}

func (s *Server) initManager() *manage.Manager {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(s.tokenStore, nil)
	manager.SetPasswordTokenCfg(&manage.Config{
		AccessTokenExp:    time.Duration(configs.Config.Oauth.AccessTokenExp) * time.Minute,
		RefreshTokenExp:   time.Duration(configs.Config.Oauth.RefreshTokenExp) * time.Minute,
		IsGenerateRefresh: true,
	})
	manager.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:     time.Duration(configs.Config.Oauth.AccessTokenExp) * time.Minute,
		RefreshTokenExp:    time.Duration(configs.Config.Oauth.RefreshTokenExp) * time.Minute,
		IsRemoveAccess:     true,
		IsGenerateRefresh:  true,
		IsRemoveRefreshing: false,
	})

	clientStore := store.NewClientStore()
	if err := clientStore.Set(configs.Config.Oauth.ClientID, &oauthModels.Client{
		ID:     configs.Config.Oauth.ClientID,
		Secret: configs.Config.Oauth.ClientSecret,
		Domain: "http://localhost",
	}); err != nil {
		panic(fmt.Errorf("failed to set client store: %w", err))
	}

	manager.MapClientStorage(clientStore)
	manager.MapAccessGenerate(token.NewLongTokenGenerate(768, 768))

	return manager
}

func (s *Server) initServer(_ context.Context) error {
	httpCfg := configs.Config.App.Url
	httpServer := http.NewHttpServer(handler, http.Port(httpCfg))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case serv := <-interrupt:
		fmt.Println("winx-auth-api - запущен - signal: " + serv.String())
		logger.Log.Println("winx-auth-api - запущен - signal: " + serv.String())
	case err := <-httpServer.Notify():
		fmt.Println(fmt.Errorf("winx-auth-api - запущен - httpServer.Notify: %w", err))
		logger.Log.Println(fmt.Errorf("winx-auth-api - запущен - httpServer.Notify: %w", err))
	}

	if err := httpServer.Shutdown(); err != nil {
		fmt.Println(fmt.Errorf("winx-auth-api - запущен - httpServer.Shutdown: %w", err))
		logger.Log.Println(fmt.Errorf("winx-auth-api - запущен - httpServer.Shutdown: %w", err))
	}

	if s.kafka != nil {
		if err := s.kafka.Close(); err != nil {
			logger.Log.Println(fmt.Errorf("failed to close kafka producer: %w", err))
		}
	}

	return nil
}
