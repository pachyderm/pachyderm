package s3

import (
	"bytes"
	"strings"
	"encoding/json"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
)

// ObjectMeta is JSON-serializable metadata about an object
type ObjectMeta struct {
	MD5 string `json:"md5"`
}

func isMeta(file string) bool {
	return strings.HasSuffix(file, ".s3g.json")
}

func getMeta(client *client.APIClient, repo, branch, file string) (*ObjectMeta, error) {
	metaReader, err := client.GetFileReader(repo, branch, fmt.Sprintf("%s.s3g.json", file), 0, 0)
	if err != nil {
		if !fileNotFoundMatcher.MatchString(err.Error()) {
			return nil, err
		}
		return nil, nil
	}

	meta := new(ObjectMeta)
	if err = json.NewDecoder(metaReader).Decode(&meta); err != nil {
		if !fileNotFoundMatcher.MatchString(err.Error()) {
			return nil, err
		}
		return nil, nil
	}

	return meta, nil
}

func putMeta(client client.PutFileClient, repo, branch, file string, meta *ObjectMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}
	metaReader := bytes.NewReader(metaBytes)
	metaFile := fmt.Sprintf("%s.s3g.json", file)
	_, err = client.PutFileOverwrite(repo, branch, metaFile, metaReader, 0)
	return err
}

func delMeta(client *client.APIClient, repo, branch, file string) error {
	if err := client.DeleteFile(repo, branch, fmt.Sprintf("%s.s3g.json", file)); err != nil {
		// ignore errors related to the metadata file not being found, since
		// it may validly not exist
		if !fileNotFoundMatcher.MatchString(err.Error()) {
			return err
		}
	}
	return nil
}
