package user

import "time"

type CreateDTO struct {
	Email    string
	Password string

	ConsentNoticeVersion string
	ConsentAcceptedAt    time.Time
}

type DeleteDTO struct {
	Email string
}

type UpdateDTO struct {
	Email    string
	NewEmail string
	Password *string
}
