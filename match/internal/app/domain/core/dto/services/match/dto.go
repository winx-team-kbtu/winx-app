package match

import "time"

// ListDTO represents pagination input for listing matches.
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}

// DeleteDTO — входные данные для удаления матча.
type DeleteDTO struct {
	ID     int64
	UserID int64
}

// MatchWithProfile is a match enriched with the matched user's profile info.
type MatchWithProfile struct {
	ID            int64
	MatchedUserID int64
	Name          string
	City          string
	Country       string
	PhotoURL      string
	CreatedAt     time.Time
}
