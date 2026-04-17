package models

import "time"

// UserInterest mirrors the user_interests junction table (Postgres).
type UserInterest struct {
	UserID     int64     `json:"user_id"`
	InterestID int64     `json:"interest_id"`
	AddedAt    time.Time `json:"added_at"`
}

func (UserInterest) TableName() string { return "user_interests" }
