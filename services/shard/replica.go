package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"

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
			continue
		}
		if err != nil {
			return nextFrom, nil
		}
		switch part.Header.Get("pfs-diff-type") {
		case "commit":
			err := cb.Commit(part)
			if err != nil {
				return nextFrom, err
			}
			// TODO(jd) we don't update nextFrom here, that's mostly ok because
			// pulls always end in a branch but ideall we'd update here as well
		case "branch":
			d := json.NewDecoder(part)
			var br btrfs.BranchRecord
			err := d.Decode(&br)
			if err != nil {
				return nextFrom, err
			}
			nextFrom := br.Name
			err = cb.Branch(br.Base, br.Name)
			if err != nil {
				return nextFrom, err
			}
		}
	}
	return nextFrom, nil
}
