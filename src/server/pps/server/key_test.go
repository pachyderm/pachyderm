package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestKey(t *testing.T) {
	var tt = []struct{ project, pipeline string }{
		{"foo", "bar"},
		{"fooß", "barþ"},
		{"üfoo", "ꙮbar"},
		{"", "bar"},
		{"foo", ""},
		{"", ""},
	}
	for i, c := range tt {
		if p, err := fromKey(toKey(&pps.Pipeline{
			Name: c.pipeline,
		})); err != nil {
			t.Errorf("case %d (%v) failed: %v", i, c, err)
		} else if p.Project.Name != "" {
			t.Errorf("case %d (%v) failed: expected project %q; got %q", i, c, "", p.Project.Name)
		} else if p.Name != c.pipeline {
			t.Errorf("case %d (%v) failed: expected pipeline %q; got %q", i, c, c.pipeline, p.Name)
		}
		if p, err := fromKey(toKey(&pps.Pipeline{
			Project: &pfs.Project{Name: c.project},
			Name:    c.pipeline,
		})); err != nil {
			t.Errorf("case %d (%v) failed: %v", i, c, err)
		} else if p.Project.Name != c.project {
			t.Errorf("case %d (%v) failed: expected project %q; got %q", i, c, c.project, p.Project.Name)
		} else if p.Name != c.pipeline {
			t.Errorf("case %d (%v) failed: expected pipeline %q; got %q", i, c, c.pipeline, p.Name)
		}
	}
}
