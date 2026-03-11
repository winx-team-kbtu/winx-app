package models

import "time"

type PasswordReset struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index;not null"`
	PinCode   string    `gorm:"index;not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"index"`
}
