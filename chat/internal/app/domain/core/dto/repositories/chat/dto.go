package chat

type CreateDTO struct {
	MatchID   int64
	UserOneID int64
	UserTwoID int64
}

type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}
