package storage

// replica.go contains code for using shards as replicas

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pachyderm/pachyderm/src/btrfs"
)

type shardReplica struct {
	url string
}

func newShardReplica(url string) *shardReplica {
	return &shardReplica{url}
}

func (r *shardReplica) Push(diff io.Reader) error {
	_, err := httpPost(r.url+"/commit", "application/octet-stream", diff)
	return err
}

func (r *shardReplica) Pull(from string, pusher btrfs.Pusher) (retErr error) {
	response, err := httpGet(fmt.Sprintf("%s/pull?from=%s", r.url, from))
	if err != nil {
		return err
	}
	defer func() {
		if err := response.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return newMultipartPuller(
		multipart.NewReader(
			response.Body,
			response.Header.Get("Boundary"),
		),
	).Pull(from, pusher)
}

func (r *shardReplica) From() (retVal string, retErr error) {
	response, err := httpGet(fmt.Sprintf("%s/commit", r.url))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := response.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var commitMsg CommitMsg
	if err = json.NewDecoder(response.Body).Decode(&commitMsg); err != nil {
		return "", err
	}
	return commitMsg.Name, nil
}

func httpGet(url string) (*http.Response, error) {
	return httpDo(url, http.Get)
}

func httpPost(url string, bodyType string, body io.Reader) (*http.Response, error) {
	return httpDo(url, func(url string) (*http.Response, error) { return http.Post(url, bodyType, body) })
}

func httpDo(url string, f func(string) (*http.Response, error)) (*http.Response, error) {
	response, err := f(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response for %s had status %s", url, response.Status)
	}
	return response, nil
}
