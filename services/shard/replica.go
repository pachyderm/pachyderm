package main

// replica.go contains code for using shards as replicas

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sync"

	"github.com/pachyderm/pfs/lib/btrfs"
)

type ShardReplica struct {
	url string
}

func NewShardReplica(url string) ShardReplica {
	return ShardReplica{url: url}
}

func (r ShardReplica) Push(diff io.Reader) error {
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

func (r ShardReplica) Pull(from string, cb btrfs.Pusher) error {
	resp, err := http.Get(fmt.Sprintf("%s/pull?from=%s", r.url, from))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return fmt.Errorf("Response with status: %s", resp.Status)
	}
	defer resp.Body.Close()
	m := NewMultiPartPuller(multipart.NewReader(resp.Body, resp.Header.Get("Boundary")))
	return m.Pull(from, cb)
}

type MultiPartCommitBrancher struct {
	w *multipart.Writer
}

func NewMultiPartCommitBrancher(w *multipart.Writer) MultiPartCommitBrancher {
	return MultiPartCommitBrancher{w: w}
}

func (m MultiPartCommitBrancher) Push(diff io.Reader) error {
	h := make(textproto.MIMEHeader)
	h.Set("pfs-diff-type", "commit")
	w, err := m.w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, diff)
	if err != nil {
		return err
	}
	return nil
}

type MultiPartPuller struct {
	r *multipart.Reader
}

func NewMultiPartPuller(r *multipart.Reader) MultiPartPuller {
	return MultiPartPuller{r: r}
}

func (m MultiPartPuller) Pull(from string, cb btrfs.Pusher) error {
	for {
		part, err := m.r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		err = cb.Push(part)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// getFrom is a convenience function to ask a shard what value it would like
// you to use for `from` when pushing to it.
func getFrom(url string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/commit", url))
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

// SyncTo syncs the contents in p to all of the shards in urls
// Returns the first error if there are multiple
func SyncTo(dataRepo string, urls []string) error {
	var errs []error
	var lock sync.Mutex
	addErr := func(err error) {
		lock.Lock()
		errs = append(errs, err)
		lock.Unlock()
	}
	lr := btrfs.NewLocalReplica(dataRepo)
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			from, err := getFrom(url)
			if err != nil {
				addErr(err)
			}
			sr := NewShardReplica(url)
			err = lr.Pull(from, sr)
			if err != nil {
				addErr(err)
			}
		}(url)
	}
	wg.Wait()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

// SyncFrom syncs from the most up to date replica in urls
func SyncFrom(dataRepo string, urls []string) error {
	for _, url := range urls {
		// First we need to figure out what value to use for `from`
		var from string
		err := btrfs.Commits(dataRepo, "", btrfs.Desc, func(c btrfs.CommitInfo) error {
			from = c.Path
			return btrfs.Complete
		})
		if err != nil && err != btrfs.Complete {
			log.Print(err)
			return err
		}

		sr := NewShardReplica(url)
		lr := btrfs.NewLocalReplica(dataRepo)

		err = sr.Pull(from, lr)
		if err != nil {
			log.Print(err)
		}
	}
	//TODO(jd) we need to figure out under what conditions this function should
	//return an error
	return nil
}
