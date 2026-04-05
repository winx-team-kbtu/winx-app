package match

// CreateDTO — данные для репозитория при создании матча.
type CreateDTO struct {
	UserOneID int64
	UserTwoID int64
}

// ListDTO — критерии поиска матчей пользователя.
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}
