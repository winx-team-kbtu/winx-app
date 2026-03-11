package models

import (
	"time"

	"gorm.io/datatypes"
)

type Notification struct {
	ID                 int64          `gorm:"primaryKey;column:id"`
	NotificationTypeID int64          `gorm:"column:notification_type_id"`
	Recipient          string         `gorm:"column:recipient"`
	Subject            string         `gorm:"column:subject"`
	Body               string         `gorm:"column:body"`
	Payload            datatypes.JSON `gorm:"column:payload;type:jsonb;default:'{}'::jsonb"`
	Status             string         `gorm:"column:status"`
	Channel            string         `gorm:"column:channel"`
	ErrorMessage       *string        `gorm:"column:error_message"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	SentAt             *time.Time     `gorm:"column:sent_at"`

	Type NotificationType `gorm:"foreignKey:NotificationTypeID;references:ID"`
}

func (Notification) TableName() string {
	return "notifications"
}
