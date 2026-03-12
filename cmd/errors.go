package cmd

import "fmt"

// ExitError wraps an error with a process exit code.
type ExitError struct {
	Code  int
	Err   error
	Quiet bool // when true, Execute() exits without printing the error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

func exitValidation(err error) *ExitError  { return &ExitError{Code: 1, Err: err} }
func exitRuntime(err error) *ExitError     { return &ExitError{Code: 2, Err: err} }
func exitConfig(err error) *ExitError      { return &ExitError{Code: 3, Err: err} }
func exitStaleness(err error) *ExitError   { return &ExitError{Code: 4, Err: err} }

func exitQuiet(code int) *ExitError {
	return &ExitError{Code: code, Quiet: true}
}
