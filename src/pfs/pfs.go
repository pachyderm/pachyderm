package pfs

import "fmt"

func VersionString(version *Version) string {
	return fmt.Sprintf("%d.%d.%d%s", version.Major, version.Minor, version.Micro, version.Additional)
}
