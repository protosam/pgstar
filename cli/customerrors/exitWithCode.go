package customerrors

import "fmt"

var ErrUnspecified = fmt.Errorf("unspecified error")

type ExitWithCode struct {
	Code    int
	Message string
}

func (e *ExitWithCode) Error() string {
	if e.Message == "" {
		return ErrUnspecified.Error()
	}
	return e.Message
}
