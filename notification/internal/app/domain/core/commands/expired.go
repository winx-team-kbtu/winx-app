package commands

import (
	repository "winx-notification/internal/app/domain/repositories"
	"context"
)

type DeleteExpiredTokenCommand struct {
	repository repository.Repository
}

func NewDeleteExpiredTokenCommand(repository repository.Repository) *DeleteExpiredTokenCommand {
	return &DeleteExpiredTokenCommand{
		repository: repository,
	}
}

func (c *DeleteExpiredTokenCommand) Handle() error {
	return c.repository.DeleteExpiredToken(context.Background())
}
