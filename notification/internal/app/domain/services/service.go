package services

import (
	"winx-notification/configs"
	kafkacontract "winx-notification/internal/app/core/contracts/microservices/kafka-contract"
	dto "winx-notification/internal/app/domain/core/dto/services"
	eventdto "winx-notification/internal/app/domain/core/dto/services/event"
	userDto "winx-notification/internal/app/domain/core/dto/services/user"
	repository "winx-notification/internal/app/domain/repositories"
	tokenService "winx-notification/internal/app/domain/services/token"
	userService "winx-notification/internal/app/domain/services/user"
	"winx-notification/internal/app/models/models"
	"winx-notification/pkg/cache"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	oauthErrs "github.com/go-oauth2/oauth2/v4/errors"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("user not found")
	ErrFailedLogin     = errors.New("login failed")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrFailedCache     = errors.New("cache failed")
	ErrFailedPublish   = errors.New("failed to publish user registered event")
)

type Service interface {
	Login(ctx context.Context, dto dto.LoginDTO) (tokenService.Response, int, error)
	Register(ctx context.Context, dto dto.RegisterDTO) (models.User, error)
	RefreshToken(ctx context.Context, refresh string) (tokenService.Response, error)
	CheckToken(ctx context.Context, access string) (models.User, error)
	Logout(ctx context.Context, email string, access string) (bool, error)
}

type service struct {
	db           *gorm.DB
	cache        cache.Cache
	repository   repository.Repository
	tokenService tokenService.Service
	tokenStore   oauth2.TokenStore
	userService  userService.Service
	kafka        kafkacontract.Producer
}

func NewService(
	db *gorm.DB,
	cache cache.Cache,
	tokenService tokenService.Service,
	tokenStore oauth2.TokenStore,
	userService userService.Service,
	kafka kafkacontract.Producer,
) Service {
	return &service{
		db:           db,
		cache:        cache,
		repository:   repository.NewRepository(db),
		tokenService: tokenService,
		tokenStore:   tokenStore,
		userService:  userService,
		kafka:        kafka,
	}
}

func (s *service) Login(ctx context.Context, input dto.LoginDTO) (tokenService.Response, int, error) {
	token, err := s.tokenService.IssueToken(ctx, map[string]string{
		"grant_type":    "password",
		"client_id":     configs.Config.Oauth.ClientID,
		"client_secret": configs.Config.Oauth.ClientSecret,
		"username":      input.Email,
		"password":      input.Password,
	})
	if err != nil {
		return tokenService.Response{}, http.StatusNotFound, ErrFailedLogin
	}

	user, err := s.userService.GetByEmail(ctx, input.Email)
	if err != nil {
		return tokenService.Response{}, http.StatusNotFound, ErrNotFound
	}

	err = s.saveTokenToCache(ctx, user.ID, user.Email, token.AccessToken, token.ExpiresIn)
	if err != nil {
		return tokenService.Response{}, http.StatusInternalServerError, ErrFailedCache
	}

	return token, http.StatusOK, err
}

func (s *service) Register(ctx context.Context, input dto.RegisterDTO) (models.User, error) {
	user, err := s.userService.Create(ctx, userDto.CreateDTO{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return models.User{}, err
	}

	payload, err := json.Marshal(eventdto.UserRegisteredDTO{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		return models.User{}, ErrFailedPublish
	}

	if err = s.kafka.Publish(ctx, configs.Config.Kafka.Topics.UserRegistered, fmt.Sprintf("%d", user.ID), payload); err != nil {
		return models.User{}, ErrFailedPublish
	}

	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, refresh string) (tokenService.Response, error) {
	token, err := s.tokenService.IssueToken(ctx, map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     configs.Config.Oauth.ClientID,
		"client_secret": configs.Config.Oauth.ClientSecret,
		"refresh_token": refresh,
	})
	if err != nil {
		if strings.Contains(err.Error(), oauthErrs.ErrInvalidGrant.Error()) {
			return tokenService.Response{}, ErrUnauthenticated
		}
	}

	user, err := s.repository.GetByAccess(ctx, token.AccessToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return tokenService.Response{}, ErrNotFound
		}

		return tokenService.Response{}, err
	}

	err = s.saveTokenToCache(ctx, user.ID, user.Email, token.AccessToken, token.ExpiresIn)
	if err != nil {
		return tokenService.Response{}, ErrFailedCache
	}

	return token, err
}

func (s *service) CheckToken(ctx context.Context, access string) (models.User, error) {
	tokenInfo, err := s.tokenService.ValidateToken(ctx, access)
	if err != nil {
		return models.User{}, ErrUnauthenticated
	}

	intUserID, err := strconv.ParseInt(tokenInfo.GetUserID(), 10, 64)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to convert user id to int: %w", err)
	}

	return s.repository.GetById(ctx, intUserID)
}

func (s *service) Logout(ctx context.Context, email string, access string) (bool, error) {
	tokenInfo, err := s.tokenService.ValidateToken(ctx, access)
	if err != nil {
		return false, ErrUnauthenticated
	}

	if err = s.tokenStore.RemoveByAccess(ctx, tokenInfo.GetAccess()); err != nil {
		return false, fmt.Errorf("failed to remove access: %w", err)
	}

	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return false, ErrNotFound
	}

	if err = s.deleteTokenFromCache(ctx, user.ID, access); err != nil {
		return false, ErrFailedCache
	}

	return true, nil
}

func (s *service) saveTokenToCache(ctx context.Context, userID int64, email string, accessToken string, ttl int64) error {
	userIDPrompt := fmt.Sprintf("user_id:%d", userID)
	userEmailPrompt := fmt.Sprintf("user_email:%d", userID)
	accessTokenPrompt := fmt.Sprintf("access_token:%s", accessToken)

	err := s.cache.Set(ctx, userIDPrompt, []byte(accessToken), time.Duration(ttl)*time.Second)
	if err != nil {
		return fmt.Errorf("save token to cache: %w, key: %s", err, userIDPrompt)
	}

	err = s.cache.Set(ctx, accessTokenPrompt, []byte(strconv.FormatInt(userID, 10)), time.Duration(ttl)*time.Second)
	if err != nil {
		return fmt.Errorf("save token to cache: %w, key: %s", err, accessTokenPrompt)
	}

	err = s.cache.Set(ctx, userEmailPrompt, []byte(email), time.Duration(ttl)*time.Second)
	if err != nil {
		return fmt.Errorf("save token to cache: %w, key: %s", err, userEmailPrompt)
	}

	return nil
}

func (s *service) deleteTokenFromCache(ctx context.Context, userID int64, accessToken string) error {
	userIDPrompt := fmt.Sprintf("user_id:%d", userID)
	userEmailPrompt := fmt.Sprintf("user_email:%d", userID)
	accessTokenPrompt := fmt.Sprintf("access_token:%s", accessToken)

	err := s.cache.Delete(ctx, userIDPrompt, userEmailPrompt)
	if err != nil {
		return fmt.Errorf("delete token from cache: %w, key: %s", err, userIDPrompt)
	}

	err = s.cache.Delete(ctx, accessTokenPrompt, strconv.FormatInt(userID, 10))
	if err != nil {
		return fmt.Errorf("save token to cache: %w, key: %s", err, accessTokenPrompt)
	}

	return nil
}
