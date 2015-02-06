package pfsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pachyderm/pfs/lib/mapreduce"
)

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func (client *Client) Commit(branch, commit string, run bool) error {
	url := fmt.Sprintf("%s/commit?branch=%s", client.url, branch)
	if commit != "" {
		url += fmt.Sprintf("&commit=%s", commit)
	}
	if run {
		url += "&run"
	}
	resp, err := http.Post(url, "application/text", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Got status code: %d", resp.StatusCode)
	}
	return nil
}

func (client *Client) Branch(commit, branch string) error {
	url := fmt.Sprintf("%s/branch?commit=%s&branch=%s", client.url, commit, branch)
	resp, err := http.Post(url, "application/text", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Got status code: %d", resp.StatusCode)
	}
	return nil
}

func (client *Client) MakeJob(branch, name string, job mapreduce.Job) error {
	jobJson, err := json.Marshal(job)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/job/%s?branch=%s", client.url, name, branch)
	resp, err := http.Post(url, "application/text", bytes.NewReader(jobJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Got status code: %d", resp.StatusCode)
	}
	return nil
}
