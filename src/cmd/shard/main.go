package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"go.pedge.io/protolog"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/storage"
)

func main() {
	if err := do(); err != nil {
		protolog.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	if err := btrfs.CheckVersion(); err != nil {
		return err
	}

	shardNum := flag.Int("shard", -1, "Optional. The shard to service.")
	modulos := flag.Int("modulos", 4, "The total number of shards.")
	address := flag.String("address", "", "Optional. The address to advertise for this node.")
	flag.Parse()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	if *address == "" {
		// No address, we'll try to use our ip addr instead
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					*address = ipnet.IP.String()
					break
				}
			}
		}
	}
	if *address == "" {
		return fmt.Errorf("pfs: Couldn't find machine ip.")
	}

	shard := storage.NewShard(
		"http://"+*address,
		fmt.Sprintf("data-%d-%d", *shardNum, *modulos),
		fmt.Sprintf("pipe-%d-%d", *shardNum, *modulos),
		uint64(*shardNum),
		uint64(*modulos),
		etcache.NewCache(),
	)
	if *shardNum == -1 {
		go shard.FindRole()
	} else {
		if err := shard.EnsureRepos(); err != nil {
			return err
		}
		go shard.FillRole()
	}
	protolog.Println("Listening on port 80...")
	return http.ListenAndServe(":80", storage.NewShardHTTPHandler(shard))
}
