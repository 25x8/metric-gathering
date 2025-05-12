// Package pattern demonstrates returning from main instead of calling os.Exit
package pattern

import (
	"errors"
	"fmt"
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func Run() error {
	shouldFail := true

	if shouldFail {
		return &ExitError{
			Code: 1,
			Err:  errors.New("operation failed"),
		}
	}

	fmt.Println("Operation succeeded")
	return nil
}
