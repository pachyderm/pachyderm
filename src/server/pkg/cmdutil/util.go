package cmdutil

import (
	"fmt"
	"io"
	"os"
)

func readLine(r io.Reader) (string, error) {
	// Read one byte at a time to avoid over-reading
	result := ""
	buf := make([]byte, 1)
	for string(buf) != "\n" {
		_, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		result += string(buf)
	}
	return result, nil
}

// InteractiveConfirm will ask the user to confirm an action on the command-line with a y/n response.
func InteractiveConfirm() (bool, error) {
	fmt.Fprintf(os.Stderr, "Are you sure you want to do this? (y/N): ")
	s, err := readLine(os.Stdin)
	if err != nil {
		return false, err
	}
	fmt.Printf("got confirm result: %s\n", s)
	if s[0] == 'y' || s[0] == 'Y' {
		return true, nil
	}
	return false, nil
}
