package storage

// replica.go contains code for using shards as replicas

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sync"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/log"
)

type shardReplica struct {
	url string
}

func NewShardReplica(url string) btrfs.Replica {
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
	m := NewMultipartPuller(multipart.NewReader(resp.Body, resp.Header.Get("Boundary")))
	return m.Pull(from, cb)
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

type multipartPusher struct {
	w *multipart.Writer
}

func NewMultipartPusher(w *multipart.Writer) btrfs.Pusher {
	return &multipartPusher{w}
}

func (m *multipartPusher) From() (string, error) {
	return "", nil
}

func (m *multipartPusher) Push(diff io.Reader) error {
	h := make(textproto.MIMEHeader)
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

type multipartPuller struct {
	r *multipart.Reader
}

func NewMultipartPuller(r *multipart.Reader) btrfs.Puller {
	return &multipartPuller{r}
}

func (m *multipartPuller) Pull(from string, cb btrfs.Pusher) error {
	for {
		part, err := m.r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = cb.Push(part)
		if err != nil {
			return err
		}
	}
	return nil
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
			sr := NewShardReplica(url)
			from, err := sr.From()
			if err != nil {
				addErr(err)
			}
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
// Returns the first error if ALL urls error.
func SyncFrom(dataRepo string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	errCount := 0
	for _, url := range urls {
		if err := syncFromUrl(dataRepo, url); err != nil {
			errCount++
		}
	}
	if errCount == len(urls) {
		return fmt.Errorf("all urls %v had errors", urls)
	}
	return nil
}

func syncFromUrl(dataRepo string, url string) error {
	sr := NewShardReplica(url)
	lr := btrfs.NewLocalReplica(dataRepo)
	from, err := lr.From()
	if err != nil {
		return err
	}
	return sr.Pull(from, lr)
}
