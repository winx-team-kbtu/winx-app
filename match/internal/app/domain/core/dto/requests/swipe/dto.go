package swipe

// SwipeDTO — тело HTTP-запроса на свайп.
// Используется хендлером: internal/app/domain/handlers/swipe/handler.go
type SwipeDTO struct {
	SwipedID int64 `json:"swiped_id" validate:"required,min=1"`
}
