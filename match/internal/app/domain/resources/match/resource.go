package match

import (
	"time"

	"winx-match/internal/app/models/models"
)

// Resource — представление матча в HTTP-ответе.
// Добавляй/убирай поля по необходимости.
type Resource struct {
	ID        int64     `json:"id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResource(m models.Match) *Resource {
	return &Resource{
		ID:        m.ID,
		UserOneID: m.UserOneID,
		UserTwoID: m.UserTwoID,
		CreatedAt: m.CreatedAt,
	}
}

func NewCollection(matches []models.Match) []*Resource {
	result := make([]*Resource, 0, len(matches))
	for _, m := range matches {
		result = append(result, NewResource(m))
	}
	return result
}
