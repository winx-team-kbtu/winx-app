package models

import "time"

type ProfilePhoto struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"object_key"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	Width     *int      `json:"width"`
	Height    *int      `json:"height"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProfilePhoto) TableName() string {
	return "profile_photos"
}
