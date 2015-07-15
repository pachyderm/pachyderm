package main

import (
	"fmt"
	"net"
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
	address := ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				address = ipnet.IP.String()
				break
			}
		}
	}
	if address == "" {
		return fmt.Errorf("pfs: Couldn't find machine ip.")
	}

	shardNum, modulos, err := route.ParseShard(shardStr)
	if err != nil {
		return err
	}
	shard := storage.NewShard(
		"http://"+address,
		"data-"+shardStr,
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
