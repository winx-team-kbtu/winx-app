package notification

import (
	"time"

	"winx-notification/internal/app/models/models"
)

type Resource struct {
	ID         int64      `json:"id"`
	Type       string     `json:"type"`
	Channel    string     `json:"channel"`
	Recipient  string     `json:"recipient"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	SentAt     *time.Time `json:"sent_at"`
	Error      *string    `json:"error_message"`
}

func NewResource(item models.Notification) *Resource {
	return &Resource{
		ID:        item.ID,
		Type:      item.Type.Code,
		Channel:   item.Channel,
		Recipient: item.Recipient,
		Subject:   item.Subject,
		Body:      item.Body,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		SentAt:    item.SentAt,
		Error:     item.ErrorMessage,
	}
}

func NewCollection(items []models.Notification) []*Resource {
	out := make([]*Resource, 0, len(items))
	for _, item := range items {
		out = append(out, NewResource(item))
	}

	return out
}
