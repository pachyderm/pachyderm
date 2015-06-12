package main

import (
	"log"
	"os"
	"path"

	"github.com/pachyderm/pfs/lib/shard"
)

func main() {
	log.SetFlags(log.Lshortfile)
	if err := os.MkdirAll("/var/lib/pfs/log", 0777); err != nil {
		log.Fatal(err)
	}
	logF, err := os.Create(path.Join("/var/lib/pfs/log", "log-"+os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}
	defer logF.Close()
	log.SetOutput(logF)

	s, err := shard.ShardFromArgs()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.EnsureRepos(); err != nil {
		log.Fatal(err)
	}

	log.Print("Listening on port 80...")
	cancel := make(chan struct{})
	defer close(cancel)
	go s.FillRole(cancel)
	s.RunServer()
}
