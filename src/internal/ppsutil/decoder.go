package ppsutil

import (
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	decoder    *yaml.Decoder
	next       func() (*ppsclient.CreatePipelineRequest, error)
	noValidate bool
}

// NewPipelineManifestReader creates a new manifest reader which reads manifests
// from an io.Reader.
func NewPipelineManifestReader(r io.Reader) (result *PipelineManifestReader, retErr error) {
	return &PipelineManifestReader{
		decoder: yaml.NewDecoder(r),
	}, nil
}

// DisableValidation disables pipeline validation.
//
// TODO(INT-1006): this exists only because the implementation of the /_mount_datums
// endpoint in the FUSE server parses its PUT body as a full pipeline spec.
func (r *PipelineManifestReader) DisableValidation() *PipelineManifestReader {
	r.noValidate = true
	return r
}

type unvalidatedCreatePipelineRequest ppsclient.CreatePipelineRequest

var _ proto.Message = new(unvalidatedCreatePipelineRequest)

// ProtoReflect implements proto.Message.
func (ucpr *unvalidatedCreatePipelineRequest) ProtoReflect() protoreflect.Message {
	return (*ppsclient.CreatePipelineRequest)(ucpr).ProtoReflect()
}

func (r *PipelineManifestReader) convertRequest(request interface{}) (*ppsclient.CreatePipelineRequest, error) {
	var result unvalidatedCreatePipelineRequest
	if err := serde.RoundTrip(request, &result); err != nil {
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	if r.noValidate {
		return (*ppsclient.CreatePipelineRequest)(&result), nil
	}
	return r.validateRequest(&result)
}

func (r *PipelineManifestReader) validateRequest(req *unvalidatedCreatePipelineRequest) (*ppsclient.CreatePipelineRequest, error) {
	if req.Pipeline == nil {
		return nil, errors.New("no `pipeline` specified")
	}
	if req.Pipeline.Name == "" {
		return nil, errors.New("no pipeline `name` specified")
	}

	if req.Transform != nil && req.Transform.Image != "" {
		if !strings.Contains(req.Transform.Image, ":") {
			fmt.Fprintf(os.Stderr,
				"WARNING: please specify a tag for the docker image in your transform.image spec.\n"+
					"For example, change 'python' to 'python:3' or 'bash' to 'bash:5'. This improves\n"+
					"reproducibility of your pipelines.\n\n")
		} else if strings.HasSuffix(req.Transform.Image, ":latest") {
			fmt.Fprintf(os.Stderr,
				"WARNING: please do not specify the ':latest' tag for the docker image in your\n"+
					"transform.image spec. For example, change 'python:latest' to 'python:3' or\n"+
					"'bash:latest' to 'bash:5'. This improves reproducibility of your pipelines.\n\n")
		}
	}
	return (*ppsclient.CreatePipelineRequest)(req), nil
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
			return r.convertRequest(document[index-1])
		}
		return r.NextCreatePipelineRequest() // Just load first result from next()
	default:
		// doc is a single request
		return r.convertRequest(document)
	}
}
