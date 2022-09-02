package ppsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
)

func TestParsePipelineKey(t *testing.T) {
	var cases = map[string]struct {
		isError      bool
		projectName  string
		pipelineName string
		id           string
	}{
		"": {isError: true},
	}
	for k, expected := range cases {
		projectName, pipelineName, id, err := ppsdb.ParsePipelineKey(k)
		if (err != nil) != expected.isError {
			t.Errorf("ParsePipelineKey(%q): expected error %v", k, expected.isError)
		}
		if err != nil {
			continue
		}
		if expected.projectName != projectName {
			t.Errorf("ParsePipelineKey(%q): expected project name %q; got %q", k, expected.projectName, projectName)
		}
		if expected.pipelineName != pipelineName {
			t.Errorf("ParsePipelineKey(%q): expected pipeline name %q; got %q", k, expected.pipelineName, pipelineName)
		}
		if expected.id != id {
			t.Errorf("ParsePipelineKey(%q): expected id  %q; got %q", k, expected.id, id)
		}
	}
}
