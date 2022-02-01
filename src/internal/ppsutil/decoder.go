package ppsutil

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	// Do first round of parsing (yaml -> holder) here (not serde), in case of
	// multi-pipeline document
	decoder *yaml.Decoder
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(pipelineBytes []byte) (result *PipelineManifestReader, retErr error) {
	return &PipelineManifestReader{
		decoder: yaml.NewDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	holder := make(map[string]interface{})
	if err := r.decoder.Decode(&holder); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.EnsureStack(err)
		}
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	var result ppsclient.CreatePipelineRequest
	if err := serde.RoundTrip(holder, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
