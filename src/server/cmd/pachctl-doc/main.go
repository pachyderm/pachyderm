package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"

	"github.com/spf13/cobra/doc"
)

type appEnv struct{}

func main() {
	cmdutil.Main(context.Background(), do, &appEnv{})
}

func do(ctx context.Context, appEnvObj interface{}) error {
	// Set 'os.Args[0]' so that examples use the expected command name
	os.Args[0] = "pachctl"

	path := "./doc/docs/master/reference/pachctl/"
	if len(os.Args) == 2 {
		path = os.Args[1]
	}

	rootCmd, err := cmd.PachctlCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not generate pachctl command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.DisableAutoGenTag = true
	return errors.EnsureStack(doc.GenMarkdownTree(rootCmd, path))
}
