package models

import "time"

type NotificationType struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code"`
	Name      string    `gorm:"column:name"`
	Channel   string    `gorm:"column:channel"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (NotificationType) TableName() string {
	return "notification_types"
}
