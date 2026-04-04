package models

import "time"

// Swipe хранит запись о свайпе одного пользователя на другого.
// Direction: "left" | "right"
type Swipe struct {
	ID        int64     `json:"id"`
	SwiperID  int64     `json:"swiper_id"`
	SwipedID  int64     `json:"swiped_id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

func (Swipe) TableName() string { return "swipes" }
