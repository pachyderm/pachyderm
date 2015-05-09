package main

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// FillRole attempts to find a role in the cluster. Once on is found it
// prepares the local storage for the role and announces the shard to the rest
// of the cluster. This function will loop until `cancel` is closed.
func (s Shard) FillRole(cancel chan struct{}) error {
	shard := fmt.Sprintf("%d-%d", s.shard, s.modulos)
	masterKey := path.Join("/pfs/master", shard)
	replicaDir := path.Join("/pfs/replica", shard)

	amMaster := false //true if we're master
	replicaKey := ""
	for {
		client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
		// First we attempt to become the master for this shard
		if !amMaster {
			// We're not master, so we attempt to claim it, this will error if
			// another shard is already master
			_, err := client.Create(masterKey, s.url, 60)
			if err == nil {
				// no error means we succesfully claimed master
				amMaster = true
				err = s.EnsureRepos()
				if err != nil {
					log.Print(err)
				}
			}
		} else {
			_, err := client.CompareAndSwap(masterKey, s.url, 60, s.url, 0)
			if err != nil { // error means we failed to reclaim master
				amMaster = false
			}
		}

		if !amMaster {
			// We didn't claim master, so we add ourselves as replica instead.
			if replicaKey == "" {
				resp, err := client.CreateInOrder(replicaDir, s.url, 60)
				if err != nil {
					log.Print(err)
				} else {
					replicaKey = resp.Node.Key
					err = s.EnsureReplicaRepos()
					if err != nil {
						log.Print(err)
					}
				}
			} else {
				_, err := client.CompareAndSwap(replicaKey, s.url, 60, s.url, 0)
				if err != nil {
					replicaKey = ""
				}
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
