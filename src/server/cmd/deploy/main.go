package main

import (
	"go.pedge.io/env"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
)

type appEnv struct{}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	return cmds.DeployCmd().Execute()
}
