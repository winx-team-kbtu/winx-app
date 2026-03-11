package models

import (
	"time"

	"github.com/lib/pq"
)

func (o *OAuthClient) TableName() string {
	return "oauth_clients"
}

type OAuthClient struct {
	ID           string         `gorm:"primaryKey;column:id"`
	Secret       string         `gorm:"column:secret"`
	RedirectURIs pq.StringArray `gorm:"type:text[];column:redirect_uris"`
	Scopes       pq.StringArray `gorm:"type:text[];column:scopes"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
}
