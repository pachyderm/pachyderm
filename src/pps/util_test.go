package pps_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestPipelineNameValidation(t *testing.T) {
	var cases = map[string]bool{
		"pIpeline_0": true,
		"pipe line":  false,
		"pipelinesß": false,
		"pipe/line":  false,
	}
	for name, expected := range cases {
		if (pps.ValidatePipelineName(name) == nil) != expected {
			t.Errorf("incorrect validation of pipeline name %q (expected %v)", name, expected)
		}
	}
}
