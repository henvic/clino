package clino

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

// ExitError wraps the error, adding an exit code.
//
// You can use it to exit the process gracefully with a specific exit code when something goes wrong.
// You should only wrap error when err != nil. You don't need to wrap it if the exit code is *exec.ExitError.
type ExitError struct {
	Code int
	Err  error
}

// Error returns the original (wrapped) error message.
func (ee ExitError) Error() string {
	return fmt.Sprintf("%v", ee.Err)
}

// Unwrap error.
func (ee ExitError) Unwrap() error { return ee.Err }

// ExitCode from the command for the process to use when exiting.
// It returns 0 if the error is nil.
// If the error comes from *exec.Cmd Run, the same child process exit code
// is used. If the error is ExitError, it returns the Code field.
// Otherwise, return exit code 1.
// 	func main() {
//		p := clino.Program{
// 			Root: &RootCommand{},
// 		}
// 		if err := p.Run(context.Background(), os.Args[1:]...); err != nil {
// 			fmt.Fprintf(os.Stderr, "%+v\n", err)
// 			os.Exit(clino.ExitCode(err))
// 		}
// 	}
//
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var ee ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}

	var xe *exec.ExitError
	if errors.As(err, &xe) {
		if ws, ok := xe.Sys().(syscall.WaitStatus); ok && ws.Exited() {
			return ws.ExitStatus()
		}
	}

	return 1
}
