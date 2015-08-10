package pachyderm

import "fmt"

const (
	// MajorVersion is the major version for pachyderm.
	MajorVersion = 0
	// MinorVersion is the minor version for pachyderm.
	MinorVersion = 10
	// MicroVersion is the micro version for pachyderm.
	MicroVersion = 0
	// AdditionalVersion will be "dev" is this is a development branch, "" otherwise.
	AdditionalVersion = "dev"
)

var (
	// Version returns the current version for pachyderm in the format MAJOR.MINOR.MICRO[dev].
	Version = fmt.Sprintf("%d.%d.%d%s", MajorVersion, MinorVersion, MicroVersion, AdditionalVersion)
)
