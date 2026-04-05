package swipe

import (
	"time"

	"winx-match/internal/app/models/models"
)

// Resource — представление свайпа в HTTP-ответе.
// Добавляй/убирай поля по необходимости.
type Resource struct {
	ID        int64     `json:"id"`
	SwiperID  int64     `json:"swiper_id"`
	SwipedID  int64     `json:"swiped_id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResource(s models.Swipe) *Resource {
	return &Resource{
		ID:        s.ID,
		SwiperID:  s.SwiperID,
		SwipedID:  s.SwipedID,
		Direction: s.Direction,
		CreatedAt: s.CreatedAt,
	}
}
