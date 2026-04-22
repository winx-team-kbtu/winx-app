package message

import (
	"context"
	"errors"
	"fmt"

	repodto "winx-chat/internal/app/domain/core/dto/repositories/message"
	servicedto "winx-chat/internal/app/domain/core/dto/services/message"
	chatRepo "winx-chat/internal/app/domain/repositories/chat"
	messageRepo "winx-chat/internal/app/domain/repositories/message"
	"winx-chat/internal/app/models/models"

	"gorm.io/gorm"
)

var (
	ErrChatNotFound = errors.New("chat not found")
	ErrNotMember    = errors.New("user is not a member of this chat")
)

type Service interface {
	Create(ctx context.Context, input servicedto.CreateDTO) (models.Message, error)
	List(ctx context.Context, input servicedto.ListDTO) ([]models.Message, error)
}

type service struct {
	msgRepo  messageRepo.Repository
	chatRepo chatRepo.Repository
}

func NewService(db *gorm.DB) Service {
	return &service{
		msgRepo:  messageRepo.NewRepository(db),
		chatRepo: chatRepo.NewRepository(db),
	}
}

func (s *service) Create(ctx context.Context, input servicedto.CreateDTO) (models.Message, error) {
	chat, err := s.chatRepo.FindByID(ctx, input.ChatID)
	if errors.Is(err, chatRepo.ErrNotFound) {
		return models.Message{}, ErrChatNotFound
	}
	if err != nil {
		return models.Message{}, fmt.Errorf("find chat: %w", err)
	}

	if chat.UserOneID != input.SenderID && chat.UserTwoID != input.SenderID {
		return models.Message{}, ErrNotMember
	}

	msg, err := s.msgRepo.Create(ctx, repodto.CreateDTO{
		ChatID:   input.ChatID,
		SenderID: input.SenderID,
		Content:  input.Content,
	})
	if err != nil {
		return models.Message{}, fmt.Errorf("create message: %w", err)
	}

	return msg, nil
}

func (s *service) List(ctx context.Context, input servicedto.ListDTO) ([]models.Message, error) {
	chat, err := s.chatRepo.FindByID(ctx, input.ChatID)
	if errors.Is(err, chatRepo.ErrNotFound) {
		return nil, ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find chat: %w", err)
	}

	if chat.UserOneID != input.UserID && chat.UserTwoID != input.UserID {
		return nil, ErrNotMember
	}

	return s.msgRepo.ListByChatID(ctx, repodto.ListDTO{
		ChatID: input.ChatID,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
}
