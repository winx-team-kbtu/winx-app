package models

import "time"

const (
	RoleSlugUser  = "user"
	RoleSlugAdmin = "admin"
)

type Role struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Role) TableName() string { return "roles" }
