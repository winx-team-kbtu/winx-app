package models

import "time"

type Interest struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	IconURL   string    `json:"icon_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (Interest) TableName() string {
	return "interests"
}
