package main

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"

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
	rootCmd.DisableAutoGenTag = true
	return doc.GenMarkdownTree(rootCmd, "./doc/docs/master/reference/pachctl/")
}
