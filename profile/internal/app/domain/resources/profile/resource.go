package profile

import (
	"winx-profile/internal/app/models/models"
	"time"
)

type Resource struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Bio       *string   `json:"bio"`
	AvatarURL *string   `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewResource(p models.Profile) *Resource {
	return &Resource{
		ID:        p.ID,
		UserID:    p.UserID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Bio:       p.Bio,
		AvatarURL: p.AvatarURL,
		Role:      p.Role,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
