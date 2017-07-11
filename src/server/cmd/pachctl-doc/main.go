package main

import (
	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra/doc"
)

type appEnv struct{}

func main() {
	cmdutil.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	rootCmd, err := cmd.PachctlCmd()
	if err != nil {
		return err
	}
	return doc.GenMarkdownTree(rootCmd, "./doc/pachctl/")
}
