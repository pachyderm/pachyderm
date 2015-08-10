package mainutil

import (
	"fmt"
	"os"

	"github.com/peter-edge/go-env"
)

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
