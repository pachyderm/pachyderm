package ppsutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	// Do first round of parsing (yaml -> holder) here (not serde), in case of
	// multi-pipeline document
	decoder    *yaml.Decoder
	next       func() (*pps.CreatePipelineRequest, error)
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

type unvalidatedCreatePipelineRequest pps.CreatePipelineRequest

var _ proto.Message = new(unvalidatedCreatePipelineRequest)

// ProtoReflect implements proto.Message.
func (ucpr *unvalidatedCreatePipelineRequest) ProtoReflect() protoreflect.Message {
	return (*pps.CreatePipelineRequest)(ucpr).ProtoReflect()
}

func (r *PipelineManifestReader) convertRequest(request interface{}) (*pps.CreatePipelineRequest, error) {
	var result unvalidatedCreatePipelineRequest
	if err := serde.RoundTrip(request, &result); err != nil {
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	}
	if r.noValidate {
		return (*pps.CreatePipelineRequest)(&result), nil
	}
	return r.validateRequest(&result)
}

func (r *PipelineManifestReader) validateRequest(req *unvalidatedCreatePipelineRequest) (*pps.CreatePipelineRequest, error) {
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
	return (*pps.CreatePipelineRequest)(req), nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*pps.CreatePipelineRequest, error) {
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
		r.next = func() (*pps.CreatePipelineRequest, error) {
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

// A SpecReader reads JSON or YAML specs.  There may be a single spec, a series
// of YAML documents and/or an array of specs.  This behavior is copied from
// PipelineManifestReader.
type SpecReader struct {
	decoder *yaml.Decoder
	next    func() (string, error)
}

// NewSpecReader returns a SpecReader which reads from r.
func NewSpecReader(r io.Reader) *SpecReader {
	return &SpecReader{decoder: yaml.NewDecoder(r)}
}

// Next returns the next pipeline spec as a JSON string.  If there are multiple
// YAML documents, it will return them one at a time.  If there is an array of
// specs, it will return them one at a time.
func (r *SpecReader) Next() (string, error) {
	if r.next != nil {
		result, err := r.next()
		switch {
		case errors.Is(err, io.EOF):
			return "", err
		case err != nil:
			return "", errors.Wrapf(err, "malformed pipeline spec")
		default:
			return result, nil
		}
	}
	// no list is in progress: parse the next document as either a spec or a
	// list of specs.
	var holder yaml.Node
	if err := r.decoder.Decode(&holder); err != nil {
		if errors.Is(err, io.EOF) {
			return "", io.EOF
		}
		return "", errors.Wrapf(err, "malformed pipeline spec")
	}
	if holder.Kind != yaml.DocumentNode {
		return "", errors.Errorf("unexpected YAML kind %v", holder.Kind)
	}
	content := holder.Content
	if len(content) != 1 {
		return "", errors.Errorf("expected a single YAML document; got %d", len(content))
	}
	switch content[0].Kind {
	case yaml.SequenceNode:
		index := 0
		r.next = func() (string, error) {
			index++
			if index >= len(content[0].Content) {
				r.next = nil
			}
			return r.convertRequest(content[0].Content[index-1])
		}
		return r.Next()
	case yaml.MappingNode:
		return r.convertRequest(holder.Content[0])
	default:
		return "", errors.Errorf("expected a mapping or a sequence; got %v", content[0].Kind)
	}
}

func (r *SpecReader) convertRequest(n *yaml.Node) (string, error) {
	object, err := yamlToJSON(n)
	if err != nil {
		return "", errors.Wrap(err, "could not convert YAML to JSON object")
	}
	return r.validateRequest(object)
}

func (r *SpecReader) validateRequest(jsObj any) (string, error) {
	b, err := json.Marshal(jsObj)
	if err != nil {
		return "", errors.Wrapf(err, "could not marshal %v as JSON", jsObj)
	}
	var req pps.CreatePipelineRequest
	if err := protojson.Unmarshal(b, &req); err != nil {
		return "", errors.Wrapf(err, "could not unmarshal %s as CreatePipelineRequest", string(b))
	}

	if req.Pipeline == nil {
		return "", errors.New("no `pipeline` specified")
	}
	if req.Pipeline.Name == "" {
		return "", errors.New("no pipeline `name` specified")
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
	return string(b), nil
}

// yamlToJSON converts a YAML node into an any value which represents a JSON
// value.  Strings are strings; numbers are json.Number; booleans are bools;
// nulls are nil; arrays are []any; and maps are map[string]any.
func yamlToJSON(n *yaml.Node) (any, error) {
	switch n.Kind {
	case yaml.DocumentNode:
		// documents are handled at a higher level
		return nil, errors.New("document may not be converted to JSON")
	case yaml.AliasNode:
		return nil, errors.Errorf("alias nodes not supported")
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str":
			return n.Value, nil
		case "!!int":
			if strings.HasPrefix(n.Value, "0") {
				return nil, errors.Errorf("number %s has a leading zero", n.Value)
			}
			return json.Number(n.Value), nil
		case "!!float":
			n := json.Number(n.Value)
			if n[0] == '.' {
				n = "0" + n
			}
			return n, nil
		case "!!bool":
			switch strings.ToLower(n.Value) {
			case "true":
				return true, nil
			case "false":
				return false, nil
			default:
				return nil, errors.Errorf("invalid boolean %q", n.Value)
			}
		case "!!null":
			return nil, nil
		default:
			return nil, errors.Errorf("tag %q unsupported", n.Tag)
		}
	case yaml.SequenceNode:
		var o = make([]any, len(n.Content))
		for i, n := range n.Content {
			var err error
			if o[i], err = yamlToJSON(n); err != nil {
				return nil, errors.Wrapf(err, "bad item %d in sequence", i)
			}
		}
		return o, nil
	case yaml.MappingNode:
		// mappings are represented as a list of interleaved keys and
		// values; if the length of the list is odd, something is broken
		if len(n.Content)%2 == 1 {
			return nil, errors.Errorf("odd number of values %d in mapping content node", len(n.Content))
		}
		var o = make(map[string]any, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			key, err := yamlToJSON(n.Content[i])
			if err != nil {
				return nil, errors.Wrapf(err, "bad key %v in mapping", n.Content[0])
			}
			value, err := yamlToJSON(n.Content[i+1])
			if err != nil {
				return nil, errors.Wrapf(err, "bad key %v in mapping", n.Content[0])
			}
			keyString, ok := key.(string)
			if !ok {
				return nil, errors.Errorf("mapping keys must be strings, not %T", key)
			}
			o[keyString] = value
		}
		return o, nil
	default:
		return nil, errors.Errorf("unknown kind %d", int(n.Kind))
	}
}
