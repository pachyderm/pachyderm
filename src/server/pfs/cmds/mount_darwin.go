//go:build darwin

package cmds

import "fmt"

func printWarning() {
	fmt.Println(`WARNING: Download the latest macOS and macFUSE versions or use mount on Linux for the best experience.`)
}
