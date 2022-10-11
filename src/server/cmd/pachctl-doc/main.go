package main

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl-doc/docgen"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"
	"os"
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
	rootCmd := cmd.PachctlCmd()
	return errors.EnsureStack(docgen.GenMarkdownTree(rootCmd, path))
}
