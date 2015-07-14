package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/pachyderm/pachyderm/src/storage"
)

func main() {
	if err := do(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("unknown args: %v", os.Args)
	}
	shardStr := os.Args[1]
	address, err := os.Hostname()
	if err != nil {
		return err
	}

	shardNum, modulos, err := route.ParseShard(shardStr)
	if err != nil {
		return err
	}
	shard := storage.NewShard(
		"http://"+address,
		"data-"+shardStr,
		"comp-"+shardStr,
		"pipe-"+shardStr,
		shardNum,
		modulos,
		etcache.NewCache(),
	)
	if err := shard.EnsureRepos(); err != nil {
		return err
	}
	log.Print("Listening on port 80...")
	go shard.FindRole()
	return http.ListenAndServe(":80", storage.NewShardHTTPHandler(shard))
}
