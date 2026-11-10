package user

import "time"

type CreateDTO struct {
	Email    string
	Password string

	// Consent is recorded only when the user accepted a notice at registration.
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
