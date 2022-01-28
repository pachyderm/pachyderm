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
	next    func() (*ppsclient.CreatePipelineRequest, error)
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(pipelineBytes []byte) (result *PipelineManifestReader, retErr error) {
	return &PipelineManifestReader{
		decoder: yaml.NewDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
}

func convertRequest(request interface{}) (*ppsclient.CreatePipelineRequest, error) {
	var result ppsclient.CreatePipelineRequest
	if err := serde.RoundTrip(request, &result); err != nil {
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	return &result, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	// return 2nd or later pipeline spec in a list (from a template)
	if r.next != nil {
		result, err := r.next()
		switch {
		case errors.Is(err, io.EOF):
			return nil, err
		case err != nil:
			return nil, errors.Wrapf(err, "malformed pipeline spec")
		default:
			return result, nil
		}
	}

	// No list is in progress--parse next doc (either a single spec or a
	// list)
	var holder interface{}
	if err := r.decoder.Decode(&holder); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.EnsureStack(err)
		}
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	switch document := holder.(type) {
	case []interface{}:
		// doc is a list of requests--return elements one by one
		index := 0 // captured in r.next(), below
		r.next = func() (*ppsclient.CreatePipelineRequest, error) {
			index++
			if index >= len(document) {
				r.next = nil // last request in the list--reset r.next
			}
			return convertRequest(document[index-1])
		}
		return r.NextCreatePipelineRequest() // Just load first result from next()
	default:
		// doc is a single request
		return convertRequest(document)
	}
}
