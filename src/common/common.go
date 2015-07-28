package common

import (
	"fmt"

	"go.pedge.io/protolog/logrus"

	stdlogrus "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

const (
	MajorVersion      = 0
	MinorVersion      = 10
	MicroVersion      = 0
	AdditionalVersion = "dev"
)

func init() {
	logrus.Register()
}

func VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", MajorVersion, MinorVersion, MicroVersion, AdditionalVersion)
}

func NewUUID() string {
	return uuid.NewV4().String()
}

func ForceLogColors() {
	logrus.SetPusherOptions(
		logrus.PusherOptions{
			Formatter: &stdlogrus.TextFormatter{
				ForceColors: true,
			},
		},
	)
}
