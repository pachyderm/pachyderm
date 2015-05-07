package main

import (
	"log"
	"path"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// AnnounceShard announces the shard to the rest of the cluster. This function
// will loop until `cancel` is closed.
// Announcing has 2 parts:
// - Announcing the shard as a replica which every shard does.
// - Announcing the shard as a master, which ever shard tries to do but only
//   one shard in the group succeeds at. AnnounceShard will continually try to
//   become the master and if the current master goes down it may succeed (or a
//   third shard may succeed).
func AnnounceShard(shard, url string, cancel chan struct{}) error {
	masterKey := path.Join("/pfs/master", shard)
	replicaDir := path.Join("/pfs/replica", shard)

	amMaster := false //true if we're master
	replicaKey := ""
	for {
		client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
		if !amMaster {
			_, err := client.Create(masterKey, url, 60)
			if err == nil { // no error means we succesfully claimed master
				amMaster = true
			}
		} else {
			_, err := client.CompareAndSwap(masterKey, url, 60, url, 0)
			if err != nil { // error means we failed to reclaim master
				amMaster = false
			}
		}
		if replicaKey == "" {
			resp, err := client.CreateInOrder(replicaDir, url, 60)
			if err != nil {
				log.Print(err)
			} else {
				replicaKey = resp.Node.Key
			}
		} else {
			_, err := client.CompareAndSwap(replicaKey, url, 60, url, 0)
			if err != nil {
				replicaKey = ""
			}
		}

		select {
		case <-time.After(time.Second * 45):
			continue
		case <-cancel:
			break
		}
	}
}
