//go:build darwin

package cmds

import "fmt"

func PrintWarning() {
	fmt.Println(`WARNING: Download the latest macOS and macFUSE versions or use mount on Linux for the best experience.`)
}
