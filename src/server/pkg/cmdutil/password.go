package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// ReadPassword reads a password from stdin. If stdin is a TTY, a password
// prompt will be displayed and the input will be obscured.
func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	// If stdin is a attached to the TTY (rather than being piped in), use a
	// terminal password prompt, which will hide the input
	if terminal.IsTerminal(syscall.Stdin) {
		pass, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", err
		}
		// print a newline, since `ReadPassword` gobbles the user-inputted one
		fmt.Fprintln(os.Stderr)
		return string(pass), nil
	}

	// If stdin is being piped in, just read it until newline
	return bufio.NewReader(os.Stdin).ReadString('\n')
}
