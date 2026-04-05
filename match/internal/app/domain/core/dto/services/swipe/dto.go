package swipe

// CreateDTO — входные данные для сервиса свайпа.
type CreateDTO struct {
	SwiperID  int64
	SwipedID  int64
	Direction string // "left" | "right"
}

type DeleteDTO struct {
	SwiperID int64
	SwipedID int64
}
