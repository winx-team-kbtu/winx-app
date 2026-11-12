package deck

type ListDTO struct {
	CourseID int64 `form:"course_id" validate:"required,gt=0"`
	Limit    int64 `form:"limit" validate:"omitempty,min=1,max=50"`
	Offset   int64 `form:"offset" validate:"omitempty,min=0"`
}
