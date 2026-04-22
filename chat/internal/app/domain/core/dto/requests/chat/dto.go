package chat

type ListDTO struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}
