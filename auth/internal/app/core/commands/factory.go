package commands

import (
	tokenCommands "auth/internal/app/domain/core/commands"
	repository "auth/internal/app/domain/repositories"
	"auth/pkg/postgres"
	"fmt"
)

func MakeCommand(signature string) (Command, error) {
	switch signature {
	case "delete-expired-tokens":
		return deleteExpiredTokenCommand(), nil
	default:
		return nil, fmt.Errorf("unknown signature: %s", signature)
	}
}

func deleteExpiredTokenCommand() Command {
	db := postgres.NewClient()
	repo := repository.NewRepository(db)

	return tokenCommands.NewDeleteExpiredTokenCommand(repo)
}
