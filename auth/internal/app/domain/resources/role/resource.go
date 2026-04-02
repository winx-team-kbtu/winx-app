package role

import (
	"auth/internal/app/models/models"
	"time"
)

type Resource struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewResource(role models.Role) *Resource {
	return &Resource{
		ID:        role.ID,
		Name:      role.Name,
		Slug:      role.Slug,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}
