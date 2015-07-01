package main

import (
	stdlog "log"
	"os"
	"path"

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
	if err := os.MkdirAll("/var/lib/pfs/log", 0777); err != nil {
		return err
	}
	logF, err := os.Create(path.Join("/var/lib/pfs/log", "log-"+os.Args[1]))
	if err != nil {
		return err
	}
	defer logF.Close()

	// TODO(pedge)
	log.SetLogger(stdlog.New(logF, "", stdlog.Lshortfile))

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
