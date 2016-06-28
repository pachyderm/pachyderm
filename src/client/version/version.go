package version

import (
	"fmt"

	"go.pedge.io/proto/version"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 0
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 0
	// MicroVersion is the patch number for pachyderm.
	MicroVersion = 99999
)

var (
	// AdditionalVersion is the string provided at release time
	// The value is passed to the linker at build time
	// DO NOT set the value of this variable here
	AdditionalVersion string
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
	}
)

func PrettyPrintVersion(version *protoversion.Version) string {
	result := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Micro)
	if version.Additional != "" {
		result += fmt.Sprintf("-%s", version.Additional)
	}
	return result
}
