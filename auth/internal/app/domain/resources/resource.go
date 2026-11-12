package resources

import (
	"time"

	"auth/internal/app/models/models"
)

type Resource struct {
	Email                string     `json:"email"`
	ConsentNoticeVersion *string    `json:"consent_notice_version,omitempty"`
	ConsentAcceptedAt    *time.Time `json:"consent_accepted_at,omitempty"`
}

func NewResource(user models.User) *Resource {
	return &Resource{
		Email:                user.Email,
		ConsentNoticeVersion: user.ConsentNoticeVersion,
		ConsentAcceptedAt:    user.ConsentAcceptedAt,
	}
}
