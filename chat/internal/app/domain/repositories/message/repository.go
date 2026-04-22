package message

import (
	"context"
	"fmt"

	repodto "winx-chat/internal/app/domain/core/dto/repositories/message"
	"winx-chat/internal/app/models/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, input repodto.CreateDTO) (models.Message, error)
	ListByChatID(ctx context.Context, input repodto.ListDTO) ([]models.Message, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input repodto.CreateDTO) (models.Message, error) {
	msg := models.Message{
		ChatID:   input.ChatID,
		SenderID: input.SenderID,
		Content:  input.Content,
	}

	if err := r.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return models.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

func (r *repository) ListByChatID(ctx context.Context, input repodto.ListDTO) ([]models.Message, error) {
	var messages []models.Message

	q := r.db.WithContext(ctx).
		Where("chat_id = ?", input.ChatID).
		Order("created_at ASC")

	if input.Limit > 0 {
		q = q.Limit(input.Limit)
	}
	if input.Offset > 0 {
		q = q.Offset(input.Offset)
	}

	if err := q.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return messages, nil
}
