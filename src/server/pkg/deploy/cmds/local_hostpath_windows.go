// +build windows

package cmds

import (
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func localHostPath() string {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, "AppData", "Local", "pachyderm")

	// Kubernetes expects the path to look like a unix path, so convert here
	// e.g. /D/path/to/file
	path = strings.ReplaceAll(path, "\\", "/")

	re := regexp.MustCompile("^([A-z]):/")
	path = re.ReplaceAllString(path, "/$1/")

	return path
}
