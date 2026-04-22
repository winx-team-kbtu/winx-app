package chat

import (
	"context"
	"errors"
	"fmt"

	repodto "winx-chat/internal/app/domain/core/dto/repositories/chat"
	"winx-chat/internal/app/models/models"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("chat not found")

type Repository interface {
	Create(ctx context.Context, input repodto.CreateDTO) (models.Chat, error)
	FindByMatchID(ctx context.Context, matchID int64) (models.Chat, error)
	ListByUserID(ctx context.Context, input repodto.ListDTO) ([]models.Chat, error)
	FindByID(ctx context.Context, id int64) (models.Chat, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input repodto.CreateDTO) (models.Chat, error) {
	chat := models.Chat{
		MatchID:   input.MatchID,
		UserOneID: input.UserOneID,
		UserTwoID: input.UserTwoID,
	}

	if err := r.db.WithContext(ctx).Create(&chat).Error; err != nil {
		return models.Chat{}, fmt.Errorf("failed to create chat: %w", err)
	}

	return chat, nil
}

func (r *repository) FindByMatchID(ctx context.Context, matchID int64) (models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).Where("match_id = ?", matchID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Chat{}, ErrNotFound
	}
	if err != nil {
		return models.Chat{}, fmt.Errorf("failed to find chat by match id: %w", err)
	}
	return chat, nil
}

func (r *repository) ListByUserID(ctx context.Context, input repodto.ListDTO) ([]models.Chat, error) {
	var chats []models.Chat

	q := r.db.WithContext(ctx).
		Where("user_one_id = ? OR user_two_id = ?", input.UserID, input.UserID).
		Order("created_at DESC")

	if input.Limit > 0 {
		q = q.Limit(input.Limit)
	}
	if input.Offset > 0 {
		q = q.Offset(input.Offset)
	}

	if err := q.Find(&chats).Error; err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	return chats, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Chat{}, ErrNotFound
	}
	if err != nil {
		return models.Chat{}, fmt.Errorf("failed to find chat: %w", err)
	}
	return chat, nil
}
