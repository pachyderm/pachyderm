package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/router"
)

func main() {
	if err := do(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	log.Print("Starting up...")
	r, err := router.RouterFromArgs()
	if err != nil {
		return err
	}
	return r.RunServer()
}
