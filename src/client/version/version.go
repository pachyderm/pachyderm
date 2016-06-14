package version

import (
	"fmt"
	"os"

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
		Additional: getBuildNumber(),
	}
)

func PrettyPrintVersion(version *protoversion.Version) string {
	result := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Micro)
	if version.Additional != "" {
		result += fmt.Sprintf("(%s)", version.Additional)
	}
	return result
}

func getBuildNumber() string {
	value := os.Getenv("PACH_BUILD_NUMBER")
	if value == "" {
		value = "dirty"
	}
	return value
}
