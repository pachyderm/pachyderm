package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ReadPassword reads a password from stdin. If stdin is a TTY, a password
// prompt will be displayed and the input will be obscured.
func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	// If stdin is a attached to the TTY (rather than being piped in), use a
	// terminal password prompt, which will hide the input
	if term.IsTerminal(int(syscall.Stdin)) {
		pass, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", errors.EnsureStack(err)
		}
		// print a newline, since `ReadPassword` gobbles the user-inputted one
		fmt.Fprintln(os.Stderr)
		return string(pass), nil
	}

	// If stdin is being piped in, just read it until newline
	result, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return result, errors.EnsureStack(err)
}
