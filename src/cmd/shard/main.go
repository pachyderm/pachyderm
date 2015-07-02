package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/shard"
)

func main() {
	if err := do(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	s, err := shard.ShardFromArgs()
	if err != nil {
		return err
	}
	if err := s.EnsureRepos(); err != nil {
		return err
	}

	log.Print("Listening on port 80...")
	cancel := make(chan struct{})
	defer close(cancel)
	go s.FillRole(cancel)
	return s.RunServer()
}
