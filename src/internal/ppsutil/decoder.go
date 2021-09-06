package ppsutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	decoder *yaml.Decoder
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(path string) (result *PipelineManifestReader, retErr error) {
	var pipelineBytes []byte
	if path == "-" {
		fmt.Print("Reading from stdin.\n")
		var err error
		pipelineBytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
	} else if url, err := url.Parse(path); err == nil && url.Scheme != "" {
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		pipelineBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		pipelineBytes, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	return &PipelineManifestReader{
		decoder: yaml.NewDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	holder := make(map[string]interface{})
	if err := r.decoder.Decode(&holder); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, err
		}
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	var result ppsclient.CreatePipelineRequest
	if err := serde.RoundTrip(holder, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
