package match

// ListDTO — входные данные для сервиса листинга матчей.
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}

// DeleteDTO — входные данные для удаления матча.
type DeleteDTO struct {
	ID     int64
	UserID int64
}
