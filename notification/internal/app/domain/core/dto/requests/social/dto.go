package social

type RedirectDTO struct {
	Source string `json:"source" binding:"required,oneof=google yandex mailru"`
}
