package ppsutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
)

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON.
// It's used by 'create pipeline' and 'update pipeline'
type PipelineManifestReader struct {
	decoder serde.Decoder
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(path string) (result *PipelineManifestReader, retErr error) {
	var pipelineBytes []byte
	var err error
	if path == "-" {
		fmt.Print("Reading from stdin.\n")
		pipelineBytes, err = ioutil.ReadAll(os.Stdin)
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
	} else {
		pipelineBytes, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	return &PipelineManifestReader{
		decoder: serde.NewYAMLDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	var result ppsclient.CreatePipelineRequest
	r.decoder.DecodeProtoTransform(&result, func(holder map[string]interface{}) error {
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
				err = fmt.Errorf("jsonpb parses TFJob as unexpected type %T", tfjob)
			}
			if err != nil {
				return fmt.Errorf("could not convert TFJob to text: %v", err)
			}
			delete(holder, key)
			holder["tf_job"] = map[string]interface{}{
				"tf_job": string(tfjobText),
			}
		}
		return nil
	})
	return &result, nil
}
