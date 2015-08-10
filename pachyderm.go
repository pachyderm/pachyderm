package pachyderm

import "github.com/pachyderm/pachyderm/src/pkg/protoversion"

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 0
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 10
	// MicroVersion is the current micro version for pachyderm.
	MicroVersion = 0
	// AdditionalVersion will be "dev" is this is a development branch, "" otherwise.
	AdditionalVersion = "dev"
)

var (
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
	}
)
