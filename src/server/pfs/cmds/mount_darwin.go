// +build darwin

package cmds

import "fmt"

func printWarning() {
	fmt.Println(`WARNING: FUSE is not officially supported on macOS.
You may encounter issues depending on which version of macOS you're on.`)
}
