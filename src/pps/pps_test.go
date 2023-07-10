package pps_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPipelineSpec(t *testing.T) {
	var (
		ps         pps.PipelineSpec
		openCVSpec = `{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "images"
    }
  },
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv:1.0"
  }
}`
		jsonPBSpec = `{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "images"
    }
  },
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv:1.0"
  },
  "resource_limits": {
    "cpu": "2"
  }
}`
	)
	if err := protojson.Unmarshal([]byte(openCVSpec), &ps); err != nil {
		t.Error("could not unmarshal OpenCV spec")
	}
	if err := protojson.Unmarshal([]byte(jsonPBSpec), &ps); err != nil {
		t.Errorf("could not unmarshal OpenCV spec: %v", err)
	}
}
