package role

type StoreDTO struct {
	Name string `json:"name" validate:"required,max=100"`
	Slug string `json:"slug" validate:"required,max=100"`
}

type UpdateDTO struct {
	Name string `json:"name" validate:"required,max=100"`
	Slug string `json:"slug" validate:"required,max=100"`
}
