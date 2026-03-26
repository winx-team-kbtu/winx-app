package profile

type UpdateDTO struct {
	FirstName *string `json:"first_name" validate:"omitempty,max=100"`
	LastName  *string `json:"last_name"  validate:"omitempty,max=100"`
	Bio       *string `json:"bio"        validate:"omitempty,max=1000"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url,max=500"`
}

type UpdateRoleDTO struct {
	UserID int64  `json:"user_id" validate:"required"`
	Role   string `json:"role"    validate:"required,oneof=user admin"`
}
