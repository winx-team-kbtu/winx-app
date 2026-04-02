package match

// ListDTO — параметры для листинга матчей (опциональные query-параметры).
// Используется хендлером: internal/app/domain/handlers/match/handler.go
type ListDTO struct {
	Limit  int `form:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int `form:"offset" validate:"omitempty,min=0"`
}

// DeleteDTO — параметр пути для удаления матча.
type DeleteDTO struct {
	ID int64 `uri:"id" validate:"required,min=1"`
}
