package main

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type appEnv struct{}

func main() {
	cmdutil.Main(do, &appEnv{})
}

// Walk the command tree, wrap any examples in a block-quote with shell highlighting
func recursiveBlockQuoteExamples(parent *cobra.Command) {
	if parent.Example != "" {
		parent.Example = fmt.Sprintf("```sh\n%s\n```", parent.Example)
	}

	for _, cmd := range parent.Commands() {
		recursiveBlockQuoteExamples(cmd)
	}
}

func do(appEnvObj interface{}) error {
	// Set 'os.Args[0]' so that examples use the expected command name
	os.Args[0] = "pachctl"

	rootCmd := cmd.PachctlCmd()

	recursiveBlockQuoteExamples(rootCmd)
	return doc.GenMarkdownTree(rootCmd, "./doc/pachctl/")
}
