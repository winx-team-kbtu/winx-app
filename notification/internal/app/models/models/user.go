package models

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string          { return "users" }
func (User) IDFieldName() string        { return "id" }
func (User) EmailFieldName() string     { return "email" }
func (User) PasswordFieldName() string  { return "password" }
func (User) CreatedAtFieldName() string { return "created_at" }
func (User) UpdatedAtFieldName() string { return "updated_at" }
