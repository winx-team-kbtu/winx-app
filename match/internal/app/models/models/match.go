package models

import "time"

// Match создаётся когда два пользователя свайпнули right друг на друга.
type Match struct {
	ID        int64     `json:"id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Match) TableName() string { return "matches" }
