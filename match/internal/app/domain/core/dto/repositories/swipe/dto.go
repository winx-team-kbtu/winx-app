package swipe

// CreateDTO — данные для репозитория при создании свайпа.
type CreateDTO struct {
	SwiperID  int64
	SwipedID  int64
	Direction string
}

// FindDTO — критерии поиска свайпа.
type FindDTO struct {
	SwiperID int64
	SwipedID int64
}

type DeleteDTO struct {
	SwiperID int64
	SwipedID int64
}
