package choice

type StoreDTO struct {
	CourseID    int64  `json:"course_id" validate:"required,gt=0"`
	CandidateID int64  `json:"candidate_id" validate:"required,gt=0"`
	Choice      string `json:"choice" validate:"required,oneof=like pass"`
}

type ShortlistDTO struct {
	CourseID int64 `form:"course_id" validate:"required,gt=0"`
	Limit    int64 `form:"limit" validate:"omitempty,min=1,max=50"`
	Offset   int64 `form:"offset" validate:"omitempty,min=0"`
}
