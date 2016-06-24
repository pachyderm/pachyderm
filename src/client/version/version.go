package version

import (
	"fmt"
	"os/exec"
	"strings"

	"go.pedge.io/proto/version"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 1
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 0
	// MicroVersion is the patch number for pachyderm.
	MicroVersion = 1
)

var (
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: getAdditionalVersion(),
	}
	// AdditionalVersion is the string provided at release time
	AdditionalVersion string
)

func PrettyPrintVersion(version *protoversion.Version) string {
	result := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Micro)
	if version.Additional != "" {
		result += fmt.Sprintf("-%s", version.Additional)
	}
	return result
}

func getAdditionalVersion() string {
	value := AdditionalVersion
	if value == "" {
		_, err := exec.LookPath("git")
		if err != nil {
			return "dirty"
		}
		out, err := exec.Command("git", "log", "--pretty=format:%H").Output()
		if err != nil {
			panic(err)
		}
		lines := strings.SplitAfterN(string(out), "\n", 2)
		if len(lines) < 2 {
			panic("Couldn't determine current commit hash")
		}
		value = strings.TrimSpace(lines[0])
	}
	return value
}
