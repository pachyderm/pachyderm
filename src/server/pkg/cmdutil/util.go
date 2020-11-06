package cmdutil

import (
	"bufio"
	"fmt"
	"os"
)

// InteractiveConfirm will ask the user to confirm an action on the command-line with a y/n response.
func InteractiveConfirm() (bool, error) {
	fmt.Fprintf(os.Stderr, "Are you sure you want to do this? (y/N): ")
	r := bufio.NewReader(os.Stdin)
	bytes, err := r.ReadBytes('\n')
	if err != nil {
		return false, err
	}
	if bytes[0] == 'y' || bytes[0] == 'Y' {
		return true, nil
	}
	return false, nil
}
