package models

import "time"

// ProfileRecord mirrors the profiles table (Postgres).
type ProfileRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProfileRecord) TableName() string { return "profiles" }
