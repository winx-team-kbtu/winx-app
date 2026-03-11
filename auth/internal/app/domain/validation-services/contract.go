package validation_services

import "fmt"

type Error struct {
	Code    int
	Message any
}

func (e Error) Error() string {
	return fmt.Sprint(e.Message)
}

var (
	UniqueMessage  = "unique"
	InvalidMessage = "invalid"
)
