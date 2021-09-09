package ppsutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	decoder *serde.YAMLDecoder
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
		decoder: serde.NewYAMLDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
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
