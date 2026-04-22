package message

type CreateDTO struct {
	ChatID   int64
	SenderID int64
	Content  string
}

type ListDTO struct {
	ChatID int64
	Limit  int
	Offset int
}
