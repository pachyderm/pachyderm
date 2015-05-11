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
	log.Print("ShardReplica.Commit")
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
	log.Print("ShardReplica.Branch")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/branch?commit=%s&branch=%s&force=true", r.url, base, name), nil)
	if err != nil {
		log.Print(err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Response with status: %s", resp.Status)
		return fmt.Errorf("Response with status: %s", resp.Status)
	}
	return nil
}

func (r ShardReplica) Pull(from string, cb btrfs.CommitBrancher) error {
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

func (m MultiPartPuller) Pull(from string, cb btrfs.CommitBrancher) error {
	for {
		part, err := m.r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		switch part.Header.Get("pfs-diff-type") {
		case "commit":
			err := cb.Commit(part)
			if err != nil {
				log.Print(err)
				return err
			}
		case "branch":
			d := json.NewDecoder(iotest.NewReadLogger("Branch", part))
			var br btrfs.BranchRecord
			err := d.Decode(&br)
			if err != nil {
				log.Print("Foo: ", err)
				return err
			}
			err = cb.Branch(br.Base, br.Name)
			if err != nil {
				log.Print(err)
				return err
			}
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
	log.Printf("SyncTo: %s, %#v", dataRepo, urls)
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
			log.Print("From in SyncTo: ", from)
			if err != nil {
				addErr(err)
				log.Print("getFrom: ", err)
			}
			log.Printf("Syncing to url: %s from: %s", url, from)
			sr := NewShardReplica(url)
			err = lr.Pull(from, sr)
			if err != nil {
				addErr(err)
				log.Print("Pull: ", err)
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
			return fmt.Errorf("COMPLETE")
		})
		if err != nil && err.Error() != "COMPLETE" {
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
