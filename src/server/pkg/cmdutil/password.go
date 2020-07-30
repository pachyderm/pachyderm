package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	if terminal.IsTerminal(syscall.Stdin) {
		pass, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", err
		}
		// print a newline, since `ReadPassword` gobbles the user-inputted one
		fmt.Fprintln(os.Stderr)
		return string(pass), nil
	}

	return bufio.NewReader(os.Stdin).ReadString('\n')
}
