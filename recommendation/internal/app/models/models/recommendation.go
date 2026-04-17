package models

import "time"

// Recommendation represents a suggested user profile for a given user.
type Recommendation struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	RecommendedID int64     `json:"recommended_id"`
	Score         float64   `json:"score"`
	CreatedAt     time.Time `json:"created_at"`
}

func (Recommendation) TableName() string { return "recommendations" }
