package photo

import (
	"time"

	"winx-profile/internal/app/models/models"
)

type Resource struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"object_key"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type,omitempty"`
	SizeBytes int64     `json:"size_bytes"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewResource(photo models.ProfilePhoto) *Resource {
	return &Resource{
		ID:        photo.ID,
		UserID:    photo.UserID,
		Email:     photo.Email,
		Provider:  photo.Provider,
		Bucket:    photo.Bucket,
		ObjectKey: photo.ObjectKey,
		URL:       photo.URL,
		MimeType:  photo.MimeType,
		SizeBytes: photo.SizeBytes,
		Width:     photo.Width,
		Height:    photo.Height,
		CreatedAt: photo.CreatedAt,
		UpdatedAt: photo.UpdatedAt,
	}
}
