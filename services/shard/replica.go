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
	"testing/iotest"

	"github.com/pachyderm/pfs/lib/btrfs"
)

type ShardReplica struct {
	url string
}

func NewShardReplica(url string) ShardReplica {
	return ShardReplica{url: url}
}

func (r ShardReplica) Commit(diff io.Reader) error {
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

func (r ShardReplica) Branch(base, name string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/branch?commit=%s&branch=%s", r.url, base, name), nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return fmt.Errorf("Response with status: %s", resp.Status)
	}
	return nil
}

func (r ShardReplica) Pull(from string, cb btrfs.CommitBrancher) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/pull?from=%s", r.url, from))
	if err != nil {
		return from, err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return from, fmt.Errorf("Response with status: %s", resp.Status)
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

func (m MultiPartCommitBrancher) Commit(diff io.Reader) error {
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

func (m MultiPartCommitBrancher) Branch(base, name string) error {
	h := make(textproto.MIMEHeader)
	h.Set("pfs-diff-type", "branch")
	w, err := m.w.CreatePart(h)
	if err != nil {
		return err
	}
	e := json.NewEncoder(w)
	err = e.Encode(btrfs.BranchRecord{Base: base, Name: name})
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

func (m MultiPartPuller) Pull(from string, cb btrfs.CommitBrancher) (string, error) {
	nextFrom := from
	for {
		part, err := m.r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return nextFrom, err
		}
		switch part.Header.Get("pfs-diff-type") {
		case "commit":
			err := cb.Commit(part)
			if err != nil {
				log.Print(err)
				return nextFrom, err
			}
			// Note we don't update nextFrom here, that's actually ok because
			// Pulls always end in a branch.
		case "branch":
			d := json.NewDecoder(iotest.NewReadLogger("Branch", part))
			var br btrfs.BranchRecord
			err := d.Decode(&br)
			if err != nil {
				log.Print(err)
				return nextFrom, err
			}
			nextFrom := br.Name
			err = cb.Branch(br.Base, br.Name)
			if err != nil {
				log.Print(err)
				return nextFrom, err
			}
		}
	}
	return nextFrom, nil
}

// getFrom is a convenience function to ask a shard what value it would like
// you to use for `from` when pushing to it.
func getFrom(url string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/branch", url))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var branch BranchMsg
	err = decoder.Decode(&branch)
	if err != nil {
		return "", err
	}
	return branch.Name, nil
}

// SyncTo syncs the contents in p to all of the shards in urls
// Returns the first error if there are multiple
func SyncTo(p btrfs.Puller, urls []string) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			from, err := getFrom(url)
			if err != nil {
				log.Print(err)
			}
			r := NewShardReplica(url)
			_, err = p.Pull(from, r)
			if err != nil {
				log.Print(err)
			}
		}(url)
	}
}

// SyncFrom syncs from the most up to date replica in urls
func SyncFrom(dataRepo string, urls []string) error {
	// First we need to figure out what value to use for `from`
	var from string
	err := btrfs.Commits(dataRepo, "", func(c btrfs.CommitInfo) error {
		from = c.Path
		return fmt.Errorf("COMPLETE")
	})
	if err != nil && err.Error() != "COMPLETE" {
		log.Print(err)
		return err
	}

	for _, url := range urls {
		sr := NewShardReplica(url)
		lr := btrfs.NewLocalReplica(dataRepo)

		_from, _err := sr.Pull(from, lr)
		if err != nil {
			err = _err
			log.Print(_err)
		} else {
			from = _from
		}
	}
	return err
}
