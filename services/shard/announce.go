package main

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func (s Shard) Peers() ([]string, error) {
	var peers []string
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	resp, err := client.Get(fmt.Sprintf("/pfs/replica/%d-%d", s.shard, s.modulos), false, true)
	if err != nil {
		return peers, err
	}
	for _, node := range resp.Node.Nodes {
		if node.Value != s.url {
			peers = append(peers, node.Value)
		}
	}
	return peers, err
}

func (s Shard) SyncFromPeers() error {
	peers, err := s.Peers()
	if err != nil {
		return err
	}

	err = SyncFrom(s.dataRepo, peers)
	if err != nil {
		return err
	}

	return nil
}

func (s Shard) SyncToPeers() error {
	peers, err := s.Peers()
	if err != nil {
		return err
	}

	err = SyncTo(s.dataRepo, peers)
	if err != nil {
		return err
	}

	return nil
}

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
			backfillingKey := "[backfilling]" + s.url
			_, err := client.Create(masterKey, backfillingKey, 5*60)
			if err == nil {
				// no error means we succesfully claimed master
				err = s.SyncFromPeers()
				if err != nil {
					log.Print(err)
				}
				// Attempt to finalize ourselves as master
				_, err := client.CompareAndSwap(masterKey, s.url, 60, backfillingKey, 0)
				if err != nil {
					log.Print(err)
				} else {
					// no error means that we succusfully announced ourselves as master
					// Make sure that if we got nothing from the peers we
					// initialize as a writeable repo.
					err = s.EnsureRepos()
					if err != nil {
						log.Print(err)
					}
					// Sync the new data we pulled to peers
					go s.SyncToPeers()
					//Record that we're master
					amMaster = true
				}
			}
		} else {
			// We're already master, renew our lease
			_, err := client.CompareAndSwap(masterKey, s.url, 60, s.url, 0)
			if err != nil { // error means we failed to reclaim master
				amMaster = false
			}
		}

		// We didn't claim master, so we add ourselves as replica instead.
		if replicaKey == "" {
			resp, err := client.CreateInOrder(replicaDir, s.url, 60)
			if err != nil {
				log.Print(err)
			} else {
				replicaKey = resp.Node.Key
				err = s.EnsureRepos()
				if err != nil {
					log.Print(err)
				}
				// Get ourselves up to date
				go s.SyncFromPeers()
			}
		} else {
			_, err := client.CompareAndSwap(replicaKey, s.url, 60, s.url, 0)
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
