package main

import (
	"log"

	"github.com/pachyderm/pfs/src/router"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Starting up...")
	r, err := router.RouterFromArgs()
	if err != nil {
		log.Fatal(err)
	}
	r.RunServer()
}
