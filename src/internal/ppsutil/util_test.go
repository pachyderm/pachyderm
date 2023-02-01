package ppsutil_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestPipelineRcName(t *testing.T) {
	for _, c := range []struct {
		projectName, pipelineName string
		version                   uint64
		name                      string
	}{
		{"default", "foo", 1, "default-foo-v1"},
		{"foo", "bar", 1, "foo-bar-v1"},
	} {
		p := &pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: c.projectName}, Name: c.pipelineName}, Version: c.version}
		name := ppsutil.PipelineRcName(p)
		if name != c.name {
			t.Errorf("case %v: expected %q; got %q", c, c.name, name)
		}
	}
}
