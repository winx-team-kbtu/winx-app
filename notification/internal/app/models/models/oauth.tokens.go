package models

import (
	"time"

	"gorm.io/datatypes"
)

func (o *OAuthToken) TableName() string {
	return "oauth_tokens"
}

type OAuthToken struct {
	ID               int64          `gorm:"primaryKey;column:id"`
	ClientID         string         `gorm:"column:client_id"`
	UserID           *string        `gorm:"column:user_id"`
	RedirectURI      *string        `gorm:"column:redirect_uri"`
	Scope            *string        `gorm:"column:scope"`
	Code             *string        `gorm:"uniqueIndex;column:code"`
	CodeExpiresAt    *time.Time     `gorm:"column:code_expires_at"`
	CodeExpiresIn    time.Duration  `gorm:"column:code_expires_in"`
	Access           *string        `gorm:"uniqueIndex;column:access"`
	AccessExpiresAt  *time.Time     `gorm:"column:access_expires_at"`
	Refresh          *string        `gorm:"uniqueIndex;column:refresh"`
	RefreshExpiresAt *time.Time     `gorm:"column:refresh_expires_at"`
	Payload          datatypes.JSON `gorm:"column:payload;type:jsonb;default:'{}'::jsonb"`
	CreatedAt        time.Time      `gorm:"column:created_at"`

	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}
