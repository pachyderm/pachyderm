package version

import (
	"fmt"

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
	// AdditionalVersion is the string provided at release time
	// The value is passed to the linker at build time
	AdditionalVersion string
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
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
<<<<<<< HEAD
||||||| merged common ancestors

func getAdditionalVersion() string {
	value := os.Getenv("VERSION_ADDITIONAL")
	if value == "" {
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
=======

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
>>>>>>> 800dd5ec21d1eab30ab9dcb101744e46c1887ae5
