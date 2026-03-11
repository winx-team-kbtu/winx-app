package event

import "time"

type UserRegisteredDTO struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserPasswordDTO struct {
	Email     string    `json:"email"`
	PinCode   string    `json:"pin_code"`
	CreatedAt time.Time `json:"created_at"`
}
