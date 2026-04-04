package event

import "time"

// SwipeEventDTO — Kafka-событие о свайпе.
// Публикуется сервисом, который обрабатывает действия пользователя (например, profile или отдельный swipe API).
type SwipeEventDTO struct {
	SwiperID  int64     `json:"swiper_id"`
	SwipedID  int64     `json:"swiped_id"`
	Direction string    `json:"direction"` // "left" | "right"
	CreatedAt time.Time `json:"created_at"`
}

// MatchCreatedDTO — Kafka-событие о создании матча.
// Публикуется match-сервисом при взаимном right-свайпе.
type MatchCreatedDTO struct {
	MatchID   int64     `json:"match_id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}
