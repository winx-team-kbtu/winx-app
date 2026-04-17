package recommendation

// ListDTO — query parameters for GET /recommendations.
type ListDTO struct {
	Limit  int `form:"limit"  validate:"omitempty,min=1,max=100"`
	Offset int `form:"offset" validate:"omitempty,min=0"`
}
