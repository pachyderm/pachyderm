package jobs

import (
	"errors"
	"strings"
)

// Platform is the OS/architecture that a particular artifact applies to; if we are only building an
// image for one platform, artifacts for other platforms don't get fetched or built.
type Platform string

// String implements fmt.Stringer.
func (p Platform) String() string {
	return string(p)
}

var ErrUnsupportedPlatform = errors.New("unsupported platform")

var KnownPlatforms = []Platform{
	"linux/amd64",
	"linux/i386",
	"linux/arm64",
	"linux/ppc64le",
	"linux/s390x",
	"linux/arm32v5",
	"linux/arm32v6",
	"linux/arm32v7",
	"darwin/amd64",
	"darwin/arm64",
	"windows/amd64",
}

// AllPlatforms is a special platform that represents architecture-independent files, like release
// notes or the root TLS cert chain.
const AllPlatforms Platform = "all/all"

// Architecture returns the CPU architecture of the platform.
func (p Platform) Architecture() string {
	parts := strings.SplitN(string(p), "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// Architecture returns the operating system of the platform.
func (p Platform) OS() string {
	parts := strings.SplitN(string(p), "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// GOOS returns the $GOOS environment variable for this platform.
func (p Platform) GOOS() string {
	return p.OS() // Docker and Go use the same OS names.
}

// GOARCH returns the $GOARCH environment variable for this platform.
func (p Platform) GOARCH() string {
	a := p.Architecture()
	if a == "arm32" {
		return "arm"
	}
	return a
}

// GOARM returns the $GOARM environment variable for this platform.
func (p Platform) GOARM() (value string, needed bool) {
	a := p.Architecture()
	if i := strings.Index(a, "arm32"); i > 0 {
		return a[i:], true
	}
	return a, false
}
