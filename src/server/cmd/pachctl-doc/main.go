package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra/doc"
)

type appEnv struct{}

func main() {
	cmdutil.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	// Set 'os.Args[0]' so that examples use the expected command name
	os.Args[0] = "pachctl"

	rootCmd := cmd.PachctlCmd()

	return doc.GenMarkdownTree(rootCmd, "./doc/docs/latest/reference/pachctl/")
}
