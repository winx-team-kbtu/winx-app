package models

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type Profile struct {
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

func (Profile) TableName() string { return "profiles" }
