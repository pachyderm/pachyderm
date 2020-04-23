// +build !windows

package cmds

import (
	"fmt"
)

// The delimiter for multiple paths in the the KUBECONFIG environment variable differs between unix and windows.
func concatKubeConfig(a string, b string) string {
	return fmt.Sprintf("%s:%s", a, b)
}