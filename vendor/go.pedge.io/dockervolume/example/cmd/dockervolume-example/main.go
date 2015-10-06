package main

import (
	"fmt"
	"os"

	"go.pedge.io/dockervolume"
)

func main() {
	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	return dockervolume.NewTCPServer(
		newVolumeDriver("/tmp/dockervolume-example-mount"),
		"dockervolume-example",
		":6789",
		dockervolume.ServerOptions{},
	).Serve()
}
