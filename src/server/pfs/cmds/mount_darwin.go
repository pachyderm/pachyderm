// +build darwin

package cmds

import "fmt"

func printWarning() {
	fmt.Println(`WARNING: Mount is supported on macOS versions 1.10.5 and earlier. Mount is implemented using FUSE which isn't supported in macOS 1.11. FUSE on macOS differs from Linux FUSE which causes some issues known issues. For the best experience we recommend using mount on Linux.`)
}
