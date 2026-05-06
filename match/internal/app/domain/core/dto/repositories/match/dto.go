package match

import "time"

// CreateDTO — данные для репозитория при создании матча.
type CreateDTO struct {
	UserOneID int64
	UserTwoID int64
}

// ListDTO — критерии поиска матчей пользователя.
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}

// MatchRow is returned by ListWithProfiles — a match joined with the matched
// user's basic profile info.
type MatchRow struct {
	ID            int64
	UserOneID     int64
	UserTwoID     int64
	MatchedUserID int64
	Name          string
	City          string
	Country       string
	PhotoURL      string
	CreatedAt     time.Time
}
