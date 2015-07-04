package storage

// replica.go contains code for using shards as replicas

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/log"
)

type shardReplica struct {
	url string
}

func newShardReplica(url string) *shardReplica {
	return &shardReplica{url}
}

func (r *shardReplica) Push(diff io.Reader) error {
	resp, err := http.Post(r.url+"/commit", "application/octet-stream", diff)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return fmt.Errorf("Response with status: %s", resp.Status)
	}
	return nil
}

func (r *shardReplica) Pull(from string, cb btrfs.Pusher) error {
	resp, err := http.Get(fmt.Sprintf("%s/pull?from=%s", r.url, from))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return fmt.Errorf("Response with status: %s", resp.Status)
	}
	defer resp.Body.Close()
	return newMultipartPuller(multipart.NewReader(resp.Body, resp.Header.Get("Boundary"))).Pull(from, cb)
}

func (r *shardReplica) From() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/commit", r.url))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var commit CommitMsg
	err = decoder.Decode(&commit)
	if err != nil && err != io.EOF {
		return "", err
	}
	return commit.Name, nil
}
