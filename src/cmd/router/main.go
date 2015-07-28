package main

import (
	"os"

	"go.pedge.io/protolog"

	"github.com/pachyderm/pachyderm/src/router"
)

func main() {
	if err := do(); err != nil {
		protolog.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	protolog.Println("Starting up...")
	r, err := router.RouterFromArgs()
	if err != nil {
		return err
	}
	return r.RunServer()
}
