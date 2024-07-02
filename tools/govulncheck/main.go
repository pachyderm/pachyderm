package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/testing/govulncheck"
	"golang.org/x/vuln/scan"
)

func main() {
	ctx := context.Background()

	root := govulncheck.SetupEnv()
	if err := os.Chdir(root); err != nil {
		fmt.Fprintf(os.Stderr, "Problem switching to Bazel module root: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Beginning scan relative to %v...\n", root)

	cmd := scan.Command(ctx, os.Args[1:]...)
	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	switch err := err.(type) {
	case nil:
	case interface{ ExitCode() int }:
		os.Exit(err.ExitCode())
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
