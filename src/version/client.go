package version

import (
	"fmt"
	"os"
	"runtime/debug"

	pb "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

var (
	// Overwritten at build time by linker
	AppVersion = "0.0.0"

	MajorVersion, MinorVersion, MicroVersion = getVersions()

	// AdditionalVersion is the string provided at release time
	// The value is passed to the linker at build time
	//
	// DO NOT set the value of this variable here. For some reason, if
	// AdditionalVersion is set here, the go linker will not overwrite it.
	AdditionalVersion string

	B = getBuildInfo()

	// Version is the current version for pachyderm.
	Version = &pb.Version{
		Major:           uint32(MajorVersion),
		Minor:           uint32(MinorVersion),
		Micro:           uint32(MicroVersion),
		Additional:      AdditionalVersion,
		GitCommit:       B.gitCommit,
		GitTreeModified: B.gitTreeModified,
		BuildDate:       B.buildDate,
		GoVersion:       B.goVersion,
		Platform:        B.platform,
	}
)

type buildInfo struct {
	gitCommit       string
	gitTreeModified string
	buildDate       string
	goVersion       string
	platform        string
}

func getBuildInfo() buildInfo {
	info, ok := debug.ReadBuildInfo()
	b := buildInfo{}
	if !ok {
		return b
	} else {
		b.goVersion = info.GoVersion
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				b.gitCommit = kv.Value
			case "vcs.time":
				b.buildDate = kv.Value
			case "vcs.modified":
				b.gitTreeModified = kv.Value
			case "GOARCH":
				b.platform = kv.Value
			}
		}
	}
	return b
}

func getVersions() (int, int, int) {
	var major, minor, micro int
	_, parseError := fmt.Sscanf(AppVersion, "%d.%d.%d", &major, &minor, &micro)
	if parseError != nil {
		fmt.Fprintln(os.Stderr, parseError)
		return 0, 0, 0
	}
	return major, minor, micro
}

// PrettyPrintVersion returns a version string optionally tagged with metadata.
// For example: "1.2.3", or "1.2.3rc1" if version.Additional is "rc1".
func PrettyPrintVersion(version *pb.Version) string {
	return fmt.Sprintf("%d.%d.%d%s", version.Major, version.Minor, version.Micro, version.Additional)
}

// PrettyVersion calls PrettyPrintVersion on Version and returns the result.
func PrettyVersion() string {
	return PrettyPrintVersion(Version)
}
