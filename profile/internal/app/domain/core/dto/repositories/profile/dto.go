package profile

type UpdateDTO struct {
	UserID    int64
	FirstName *string
	LastName  *string
	Bio       *string
	AvatarURL *string
}

type UpdateRoleDTO struct {
	UserID int64
	Role   string
}

type CreateDTO struct {
	UserID int64
}
