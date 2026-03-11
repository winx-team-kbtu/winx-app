package commands

type Command interface {
	Handle() error
}
