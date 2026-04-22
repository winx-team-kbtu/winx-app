package message

import (
	"time"

	"winx-chat/internal/app/models/models"
)

type Resource struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResource(m models.Message) *Resource {
	return &Resource{
		ID:        m.ID,
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func NewCollection(messages []models.Message) []*Resource {
	result := make([]*Resource, 0, len(messages))
	for _, m := range messages {
		result = append(result, NewResource(m))
	}
	return result
}
