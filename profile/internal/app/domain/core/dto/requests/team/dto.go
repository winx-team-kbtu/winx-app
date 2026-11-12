package team

type StoreDTO struct {
	CourseID int64  `json:"course_id" validate:"required,gt=0"`
	Name     string `json:"name" validate:"required,max=200"`
}

type MineDTO struct {
	CourseID int64 `form:"course_id" validate:"required,gt=0"`
}
