package models

import "time"

type Course struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func (Course) TableName() string {
	return "courses"
}
