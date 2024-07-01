// Copyright 2022 The Go Authors.
package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/vuln/scan"
)

func main() {
	ctx := context.Background()
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
