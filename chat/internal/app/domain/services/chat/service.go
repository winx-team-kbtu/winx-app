package chat

import (
	"context"
	"errors"
	"fmt"

	repodto "winx-chat/internal/app/domain/core/dto/repositories/chat"
	servicedto "winx-chat/internal/app/domain/core/dto/services/chat"
	"winx-chat/internal/app/domain/repositories/chat"
	"winx-chat/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("chat not found")

type Service interface {
	CreateFromMatch(ctx context.Context, matchID, userOneID, userTwoID int64) (models.Chat, error)
	List(ctx context.Context, input servicedto.ListDTO) ([]models.Chat, error)
	FindByID(ctx context.Context, id int64) (models.Chat, error)
}

type service struct {
	repo chat.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		repo: chat.NewRepository(db),
	}
}

func (s *service) CreateFromMatch(ctx context.Context, matchID, userOneID, userTwoID int64) (models.Chat, error) {
	// Idempotent: if the chat already exists for this match, return it.
	existing, err := s.repo.FindByMatchID(ctx, matchID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, chat.ErrNotFound) {
		return models.Chat{}, fmt.Errorf("check existing chat: %w", err)
	}

	c, err := s.repo.Create(ctx, repodto.CreateDTO{
		MatchID:   matchID,
		UserOneID: userOneID,
		UserTwoID: userTwoID,
	})
	if err != nil {
		return models.Chat{}, fmt.Errorf("create chat: %w", err)
	}

	return c, nil
}

func (s *service) List(ctx context.Context, input servicedto.ListDTO) ([]models.Chat, error) {
	return s.repo.ListByUserID(ctx, repodto.ListDTO{
		UserID: input.UserID,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
}

func (s *service) FindByID(ctx context.Context, id int64) (models.Chat, error) {
	c, err := s.repo.FindByID(ctx, id)
	if errors.Is(err, chat.ErrNotFound) {
		return models.Chat{}, ErrNotFound
	}
	return c, err
}
