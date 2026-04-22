package chat

import (
	"time"

	"winx-chat/internal/app/models/models"
)

type Resource struct {
	ID        int64     `json:"id"`
	MatchID   int64     `json:"match_id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResource(c models.Chat) *Resource {
	return &Resource{
		ID:        c.ID,
		MatchID:   c.MatchID,
		UserOneID: c.UserOneID,
		UserTwoID: c.UserTwoID,
		CreatedAt: c.CreatedAt,
	}
}

func NewCollection(chats []models.Chat) []*Resource {
	result := make([]*Resource, 0, len(chats))
	for _, c := range chats {
		result = append(result, NewResource(c))
	}
	return result
}
