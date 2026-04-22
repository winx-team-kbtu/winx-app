package models

import "time"

type Chat struct {
	ID        int64     `json:"id"`
	MatchID   int64     `json:"match_id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Chat) TableName() string { return "chats" }
