package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth/configs"
	kafkacontract "auth/internal/app/core/contracts/microservices/kafka-contract"
	dto "auth/internal/app/domain/core/dto/services"
	eventdto "auth/internal/app/domain/core/dto/services/event"
	userDto "auth/internal/app/domain/core/dto/services/user"
	repository "auth/internal/app/domain/repositories"
	tokenService "auth/internal/app/domain/services/token"
	userService "auth/internal/app/domain/services/user"
	"auth/internal/app/models/models"
	"auth/pkg/cache"
	"auth/pkg/graylog/logger"

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
	logger.LogInfo("Auth service initialized")
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
	logger.LogInfo(fmt.Sprintf("Login attempt for email: %s", input.Email))

	token, err := s.tokenService.IssueToken(ctx, map[string]string{
		"grant_type":    "password",
		"client_id":     configs.Config.Oauth.ClientID,
		"client_secret": configs.Config.Oauth.ClientSecret,
		"username":      input.Email,
		"password":      input.Password,
	})
	if err != nil {
		logger.LogWarning(fmt.Sprintf("Login failed for %s: %v", input.Email, err))
		return tokenService.Response{}, http.StatusNotFound, ErrFailedLogin
	}

	user, err := s.userService.GetByEmail(ctx, input.Email)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get user by email %s", input.Email), err)
		return tokenService.Response{}, http.StatusNotFound, ErrNotFound
	}

	err = s.saveTokenToCache(ctx, user.ID, user.Email, token.AccessToken, token.ExpiresIn)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to save token to cache for user %d", user.ID), err)
		return tokenService.Response{}, http.StatusInternalServerError, ErrFailedCache
	}

	logger.LogInfo(fmt.Sprintf("User %s logged in successfully", input.Email))
	return token, http.StatusOK, err
}

func (s *service) Register(ctx context.Context, input dto.RegisterDTO) (models.User, error) {
	logger.LogInfo(fmt.Sprintf("Registration attempt for email: %s", input.Email))

	user, err := s.userService.Create(ctx, userDto.CreateDTO{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create user %s", input.Email), err)
		return models.User{}, err
	}

	payload, err := json.Marshal(eventdto.UserRegisteredDTO{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		logger.LogError("Failed to marshal user registered event", err)
		return models.User{}, ErrFailedPublish
	}

	if err = s.kafka.Publish(ctx, configs.Config.Kafka.Topics.UserRegistered, fmt.Sprintf("%d", user.ID), payload); err != nil {
		logger.LogError("Failed to publish user registered event to Kafka", err)
		return models.User{}, ErrFailedPublish
	}

	logger.LogInfo(fmt.Sprintf("User %s registered successfully with ID %d", input.Email, user.ID))
	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, refresh string) (tokenService.Response, error) {
	logger.LogDebug("Refreshing token")

	token, err := s.tokenService.IssueToken(ctx, map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     configs.Config.Oauth.ClientID,
		"client_secret": configs.Config.Oauth.ClientSecret,
		"refresh_token": refresh,
	})
	if err != nil {
		if strings.Contains(err.Error(), oauthErrs.ErrInvalidGrant.Error()) {
			logger.LogWarning("Invalid refresh token")
			return tokenService.Response{}, ErrUnauthenticated
		}
		logger.LogError("Failed to issue refresh token", err)
	}

	user, err := s.repository.GetByAccess(ctx, token.AccessToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			logger.LogWarning("User not found for access token")
			return tokenService.Response{}, ErrNotFound
		}
		logger.LogError("Failed to get user by access token", err)
		return tokenService.Response{}, err
	}

	err = s.saveTokenToCache(ctx, user.ID, user.Email, token.AccessToken, token.ExpiresIn)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to save refreshed token to cache for user %d", user.ID), err)
		return tokenService.Response{}, ErrFailedCache
	}

	logger.LogInfo(fmt.Sprintf("Token refreshed for user %d", user.ID))
	return token, err
}

func (s *service) CheckToken(ctx context.Context, access string) (models.User, error) {
	logger.LogDebug("Checking token validity")

	tokenInfo, err := s.tokenService.ValidateToken(ctx, access)
	if err != nil {
		logger.LogWarning("Token validation failed")
		return models.User{}, ErrUnauthenticated
	}

	intUserID, err := strconv.ParseInt(tokenInfo.GetUserID(), 10, 64)
	if err != nil {
		logger.LogError("Failed to convert user ID to int", err)
		return models.User{}, fmt.Errorf("failed to convert user id to int: %w", err)
	}

	user, err := s.repository.GetById(ctx, intUserID)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get user by ID %d", intUserID), err)
		return models.User{}, err
	}

	logger.LogDebug(fmt.Sprintf("Token valid for user %d", intUserID))
	return user, nil
}

func (s *service) Logout(ctx context.Context, email string, access string) (bool, error) {
	logger.LogInfo(fmt.Sprintf("Logout attempt for email: %s", email))

	tokenInfo, err := s.tokenService.ValidateToken(ctx, access)
	if err != nil {
		logger.LogWarning("Invalid token during logout")
		return false, ErrUnauthenticated
	}

	if err = s.tokenStore.RemoveByAccess(ctx, tokenInfo.GetAccess()); err != nil {
		logger.LogError("Failed to remove access token from store", err)
		return false, fmt.Errorf("failed to remove access: %w", err)
	}

	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to get user by email %s", email), err)
		return false, ErrNotFound
	}

	if err = s.deleteTokenFromCache(ctx, user.ID, access); err != nil {
		logger.LogError(fmt.Sprintf("Failed to delete token from cache for user %d", user.ID), err)
		return false, ErrFailedCache
	}

	logger.LogInfo(fmt.Sprintf("User %s logged out successfully", email))
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