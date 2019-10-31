package ppsutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"gopkg.in/pachyderm/yaml.v3"
)

type pipelineDecoder interface {
	// Decode parses the next CreatePipelineRequest in the pipelineDecoder's input
	// stream into 'v'
	Decode(v interface{}) error
}

// jsonpbDecoder implements the 'pipelineDecoder' interface to decode
// JSON-encoded pipelines using the gogoproto json decoder (which is slightly
// more forgiving than go's built-in JSON decoder)
type jsonpbDecoder struct {
	decoder *json.Decoder
}

func newJSONPBDecoder(r io.Reader) *jsonpbDecoder {
	return &jsonpbDecoder{
		decoder: json.NewDecoder(r),
	}
}

func (d *jsonpbDecoder) Decode(v interface{}) error {
	msg, ok := v.(proto.Message)
	if ok {
		return jsonpb.UnmarshalNext(d.decoder, msg)
	}
	return d.decoder.Decode(v)
}

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON. It's used by
// 'create pipeline' and 'update pipeline'
//
// Note that the json decoder is able to parse text that
// gopkg.in/pachyderm/yaml.v3 cannot (multiple json documents) so we currently
// guess whether the document is JSON or not by looking at the first non-space
// character and seeing if it's '{' (we originally tried parsing the pipeline
// spec with both parsers, but that approach made it hard to return sensible
// errors). We may fail to parse valid YAML documents this way, so hopefully the
// yaml parser gains multi-document support and we can rely on it fully.
type PipelineManifestReader struct {
	decoder pipelineDecoder
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
	idx := bytes.IndexFunc(pipelineBytes, func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	if idx >= 0 && pipelineBytes[idx] == '{' {
		return &PipelineManifestReader{
			decoder: newJSONPBDecoder(bytes.NewReader(pipelineBytes)),
		}, nil
	}
	return &PipelineManifestReader{
		decoder: newYAMLDecoder(bytes.NewReader(pipelineBytes)),
	}, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
//
// The implementation of this is currently somewhat wasteful: it parses the
// whole request into a map, modifies the map, serializes the map back into
// JSON, and then deserializes the JSON into a CreatePipelineRequest struct.
// This is to manage embedded data (currently just TFJobs, but could later be
// spark jobs or some such), which are represented as serialized JSON in the
// CreatePipelineRequest struct, but which we parse as structured data.
//
// No effort is made to bypass this parse-serialize-parse process even in the
// common case where the pipeline contains no TFJob, because all of this happens
// only in 'pachctl', and only when a pipeline is created or updated.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	// Parse whole request into a semi-structured map
	holder := make(map[string]interface{})
	if err := r.decoder.Decode(&holder); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("malformed pipeline spec: %s", err)
	}

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
				return nil, errors.New("tf_job must contain a kubernetes manifest for a Kubeflow TFJob")
			}
			tfjobText, err = json.Marshal(tfjob)
		} else {
			err = fmt.Errorf("jsonpb parses TFJob as unexpected type %T", tfjob)
		}
		if err != nil {
			return nil, fmt.Errorf("could not convert TFJob to text: %v", err)
		}
		delete(holder, key)
		holder["tf_job"] = map[string]interface{}{
			"tf_job": string(tfjobText),
		}
	}

	// serialize 'holder' to json, then parse again into a CreatePipelineRequest
	// using jsonpb (which has special handling for timestamps and other types)
	pipelineRequestBytes, err := json.Marshal(holder)
	if err != nil {
		return nil, fmt.Errorf("serialization error while canonicalizing CreatePipelineRequest: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(pipelineRequestBytes))
	var result ppsclient.CreatePipelineRequest
	if err := jsonpb.UnmarshalNext(decoder, &result); err != nil {
		return nil, fmt.Errorf("parse error while canonicalizing CreatePipelineRequest: %v", err)
	}
	return &result, nil
}

// DescribeSyntaxError describes a syntax error encountered parsing json.
func DescribeSyntaxError(originalErr error, parsedBuffer bytes.Buffer) error {
	sErr, ok := originalErr.(*json.SyntaxError)
	if !ok {
		return originalErr
	}

	buffer := make([]byte, sErr.Offset)
	parsedBuffer.Read(buffer)

	lineOffset := strings.LastIndex(string(buffer[:len(buffer)-1]), "\n")
	if lineOffset == -1 {
		lineOffset = 0
	}

	lines := strings.Split(string(buffer[:len(buffer)-1]), "\n")
	lineNumber := len(lines)

	descriptiveErrorString := fmt.Sprintf("Syntax Error on line %v:\n%v\n%v^\n%v\n",
		lineNumber,
		string(buffer[lineOffset:]),
		strings.Repeat(" ", int(sErr.Offset)-2-lineOffset),
		originalErr,
	)

	return errors.New(descriptiveErrorString)
}
