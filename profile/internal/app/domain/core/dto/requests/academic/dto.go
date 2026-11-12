package academic

type SaveDTO struct {
	Skills        []string `json:"skills" validate:"omitempty,max=40"`
	PreferredRole string   `json:"preferred_role" validate:"omitempty,max=200"`
}
