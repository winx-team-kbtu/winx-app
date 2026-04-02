package role

type StoreDTO struct {
	Name string
	Slug string
}

type UpdateDTO struct {
	ID   int64
	Name string
	Slug string
}
