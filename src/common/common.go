package common

import (
	"fmt"

	"github.com/satori/go.uuid"
)

const (
	MajorVersion      = 0
	MinorVersion      = 10
	MicroVersion      = 0
	AdditionalVersion = "dev"
)

func VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", MajorVersion, MinorVersion, MicroVersion, AdditionalVersion)
}

func NewUUID() string {
	return uuid.NewV4().String()
}
