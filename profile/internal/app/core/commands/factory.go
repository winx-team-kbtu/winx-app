package commands

import "fmt"

type noopCommand struct{}

func MakeCommand(signature string) (Command, error) {
	switch signature {
	case "noop":
		return noopCommand{}, nil
	default:
		return nil, fmt.Errorf("unknown signature: %s", signature)
	}
}

func (noopCommand) Handle() error {
	return nil
}
