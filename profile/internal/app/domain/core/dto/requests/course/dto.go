package course

type SaveUserCoursesDTO struct {
	CourseIDs []int64 `json:"course_ids" validate:"omitempty,max=10,dive,gt=0"`
}
