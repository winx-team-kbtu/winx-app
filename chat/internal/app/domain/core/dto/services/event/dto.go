package event

import "time"

// MatchCreatedDTO — Kafka-событие о создании матча.
// Публикуется match-сервисом при взаимном right-свайпе.
type MatchCreatedDTO struct {
	MatchID   int64     `json:"match_id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}
