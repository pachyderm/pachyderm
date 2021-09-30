package ppsutil

import (
	"bytes"
	"io"
	"unicode"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	decoder serde.Decoder
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(pipelineBytes []byte) *PipelineManifestReader {
	// TODO(msteffen): if we can get the yaml decoder to handle leading tabs, as
	// in pps/cmds/cmds_test.go, then we can get rid of this
	idx := bytes.IndexFunc(pipelineBytes, func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	if idx >= 0 && pipelineBytes[idx] == '{' {
		return &PipelineManifestReader{
			decoder: serde.NewJSONDecoder(bytes.NewReader(pipelineBytes)),
		}
	}
	return &PipelineManifestReader{
		decoder: serde.NewYAMLDecoder(bytes.NewReader(pipelineBytes)),
	}
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	var result ppsclient.CreatePipelineRequest
	err := r.decoder.DecodeProtoTransform(&result, func(holder map[string]interface{}) error {
		var key string
		var ok bool
		var tfjob interface{}
		if tfjob, ok = holder["TFJob"]; ok {
			key = "TFJob" // go json default munging
		} else if tfjob, ok = holder["tf_job"]; ok {
			key = "tf_job" // protobuf-generated 'json' tag
		}
		if key != "" {
			var err error
			var tfjobText []byte
			if tfjob, ok := tfjob.(map[string]interface{}); ok {
				// tiny validation--make sure "kind" is "TFJob" (or is unset)
				if tfjob["kind"] == "" {
					tfjob["kind"] = "TFJob"
				} else if tfjob["kind"] != "TFJob" {
					return errors.New("tf_job must contain a kubernetes manifest for a Kubeflow TFJob")
				}
				tfjobText, err = serde.EncodeJSON(tfjob)
			} else {
				err = errors.Errorf("jsonpb parses TFJob as unexpected type %T", tfjob)
			}
			if err != nil {
				return errors.Wrapf(err, "could not convert TFJob to text")
			}
			delete(holder, key)
			holder["tf_job"] = map[string]interface{}{
				"tf_job": string(tfjobText),
			}
		}
		return nil
	})
	switch {
	case errors.Is(err, io.EOF):
		return nil, err
	case err != nil:
		return nil, errors.Wrapf(err, "malformed pipeline spec")
	default:
		return &result, nil
	}
}
