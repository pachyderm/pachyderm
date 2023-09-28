package ppsutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"

	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	specReader *SpecReader
}

// NewPipelineManifestReader creates a new manifest reader which reads manifests
// from an io.Reader.
func NewPipelineManifestReader(r io.Reader) (result *PipelineManifestReader, retErr error) {
	return &PipelineManifestReader{NewSpecReader(r)}, nil
}

// DisableValidation disables pipeline validation.
//
// TODO(INT-1006): this exists only because the implementation of the /_mount_datums
// endpoint in the FUSE server parses its PUT body as a full pipeline spec.
func (r *PipelineManifestReader) DisableValidation() *PipelineManifestReader {
	r.specReader = r.specReader.DisableValidation()
	return r
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*pps.CreatePipelineRequest, error) {
	spec, err := r.specReader.Next()
	if err != nil {
		return nil, err
	}
	var req pps.CreatePipelineRequest
	if err := protojson.Unmarshal([]byte(spec), &req); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal CreatePipelineRequest %s", spec)
	}
	return &req, nil
}

// A SpecReader reads JSON or YAML specs.  There may be a single spec, a series
// of YAML documents and/or an array of specs.  This behavior is copied from
// PipelineManifestReader.
type SpecReader struct {
	decoder    *yaml.Decoder
	next       func() (string, error)
	noValidate bool
}

// NewSpecReader returns a SpecReader which reads from r.
func NewSpecReader(r io.Reader) *SpecReader {
	return &SpecReader{decoder: yaml.NewDecoder(r)}
}

// DisableValidation disables pipeline validation.
//
// TODO(INT-1006): this exists only because the implementation of the /_mount_datums
// endpoint in the FUSE server parses its PUT body as a full pipeline spec.
func (r *SpecReader) DisableValidation() *SpecReader {
	r.noValidate = true
	return r
}

// Next returns the next pipeline spec as a JSON string.  If there are multiple
// YAML documents, it will return them one at a time.  If there is an array of
// specs, it will return them one at a time.
func (r *SpecReader) Next() (string, error) {
	if r.next != nil {
		result, err := r.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", err
			}
			return "", errors.Wrapf(err, "malformed pipeline spec")
		}
		return result, nil
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

	b, err := json.Marshal(object)
	if err != nil {
		return "", errors.Wrapf(err, "could not marshal %v as JSON", object)
	}

	if r.noValidate {
		return string(b), nil
	}

	return r.validateRequest(b)
}

func (r *SpecReader) validateRequest(b []byte) (string, error) {
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

var intRegexp = regexp.MustCompile("^(0|(-?[1-9][0-9]*))$")

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
			if !intRegexp.MatchString(n.Value) {
				return nil, errors.Errorf("integer value %q does not match regexp %s", n.Value, intRegexp)
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
			if keyString != constants.JSONSchemaKey { // Remove JSON schema tags if the editor added them.
				o[keyString] = value
			}
		}
		return o, nil
	default:
		return nil, errors.Errorf("unknown kind %d", int(n.Kind))
	}
}
