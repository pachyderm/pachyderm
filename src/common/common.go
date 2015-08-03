/*
Package common contains common functions for pachyderm.

Nothing in src/pkg should rely on this as these packages are meant to be extractable.
*/
package common

import (
	"fmt"
	"os"

	"go.pedge.io/protolog/logrus"

	stdlogrus "github.com/Sirupsen/logrus"
	"github.com/peter-edge/go-env"
	"github.com/satori/go.uuid"
)

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

func init() {
	logrus.Register()
}

// VersionString returns the current version for pachyderm in the format MAJOR.MINOR.MICRO[dev].
func VersionString() string {
	return fmt.Sprintf("%d.%d.%d%s", MajorVersion, MinorVersion, MicroVersion, AdditionalVersion)
}

// NewUUID returns a new unique ID as a string. This wraps the construction of UUIDs
// for all pachyderm code so we can switch libraries quickly.
func NewUUID() string {
	return uuid.NewV4().String()
}

// ForceLogColors will register the logrus protolog.Pusher as the global protolog.Logger,
// and force terminal colors. Only use this in tests.
func ForceLogColors() {
	logrus.SetPusherOptions(
		logrus.PusherOptions{
			Formatter: &stdlogrus.TextFormatter{
				ForceColors: true,
			},
		},
	)
}

func Main(do func(interface{}) error, appEnv interface{}, defaultEnv map[string]string) {
	if err := env.Populate(appEnv, env.PopulateOptions{Defaults: defaultEnv}); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	if err := do(appEnv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
