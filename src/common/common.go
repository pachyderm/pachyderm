package common

import "fmt"

const (
	MajorVersion      = 0
	MinorVersion      = 10
	MicroVersion      = 0
	AdditionalVersion = "dev"
)

func VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", MajorVersion, MinorVersion, MicroVersion, AdditionalVersion)
}
