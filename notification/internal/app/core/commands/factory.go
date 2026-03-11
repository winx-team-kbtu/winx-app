package commands

import (
	tokenCommands "winx-notification/internal/app/domain/core/commands"
	repository "winx-notification/internal/app/domain/repositories"
	"winx-notification/pkg/postgres"
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
