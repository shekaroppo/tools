package todo

import "fmt"

type CommandArgumentError int
type InvalidArgument struct {
	value string
	field string
}
type TaskNotFound int

func (e CommandArgumentError) Error() string {
	return fmt.Sprintf("This command exactly requires '%d' arguments", int(e))
}

func (e InvalidArgument) Error() string {
	return fmt.Sprintf("Invalid argument '%s' for %s", e.value, e.field)
}

func (e TaskNotFound) Error() string {
	return fmt.Sprintf("Task with id=%d not found", int(e))
}
