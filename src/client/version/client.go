package version

import (
	"fmt"

	pb "github.com/pachyderm/pachyderm/src/client/version/versionpb"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 1
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 10
	// MicroVersion is the patch number for pachyderm.
	MicroVersion = 2
)

var (
	// AdditionalVersion is the string provided at release time
	// The value is passed to the linker at build time
	//
	// DO NOT set the value of this variable here. For some reason, if
	// AdditionalVersion is set here, the go linker will not overwrite it.
	AdditionalVersion string

	// Version is the current version for pachyderm.
	Version = &pb.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
	}
)

// PrettyPrintVersion returns a version string optionally tagged with metadata.
// For example: "1.2.3", or "1.2.3rc1" if version.Additional is "rc1".
func PrettyPrintVersion(version *pb.Version) string {
	result := PrettyPrintVersionNoAdditional(version)
	if version.Additional != "" {
		result += version.Additional
	}
	return result
}

// IsAtLeast returns true if Pachyderm is at least at the given version. This
// allows us to gate backwards-incompatible features on release boundaries.
func IsAtLeast(major, minor int) bool {
	return MajorVersion > major || (MajorVersion == major && MinorVersion >= minor)
}

// PrettyVersion calls PrettyPrintVersion on Version and returns the result.
func PrettyVersion() string {
	return PrettyPrintVersion(Version)
}

// PrettyPrintVersionNoAdditional returns a version string without
// version.Additional.
func PrettyPrintVersionNoAdditional(version *pb.Version) string {
	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Micro)
}
