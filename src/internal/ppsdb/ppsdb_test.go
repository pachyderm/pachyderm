package ppsdb

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestParsePipelineKey(t *testing.T) {
	var cases = map[string]struct {
		isError      bool
		projectName  string
		pipelineName string
		id           string
	}{
		"": {isError: true},
		// new format
		"project/pipeline@0123456789ab40123456789abcdef012": {
			isError:      false,
			projectName:  "project",
			pipelineName: "pipeline",
			id:           "0123456789ab40123456789abcdef012",
		},
		// old format
		//
		// TODO: this should be removed after migration is completed in
		// CORE-93, perhaps as part of CORE-1046
		"pipeline@0123456789ab40123456789abcdef012": {
			isError:      false,
			projectName:  "",
			pipelineName: "pipeline",
			id:           "0123456789ab40123456789abcdef012",
		},
	}
	for k, expected := range cases {
		projectName, pipelineName, id, err := ParsePipelineKey(k)
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

func TestCommitKey(t *testing.T) {
	var cases = []struct {
		isError     bool
		projectName string
		repoName    string
		id          string
		key         string
	}{
		// new format
		{false, "project", "repo", "0123456789ab40123456789abcdef012", "project/repo@0123456789ab40123456789abcdef012"},
		// old format
		//
		// TODO: this should be removed after migration is completed in
		// CORE-93, perhaps as part of CORE-1046
		{false, "", "repo", "0123456789ab40123456789abcdef012", "repo@0123456789ab40123456789abcdef012"},
	}
	for i, c := range cases {
		if got, err := pipelineCommitKey(&pfs.Commit{
			Repo: &pfs.Repo{
				Project: &pfs.Project{
					Name: c.projectName,
				},
				Name: c.repoName,
				Type: pfs.SpecRepoType,
			},
			Id: c.id,
		}); err != nil {
			t.Errorf("unexpected error with test case %d: %v", i, c)
		} else if got != c.key {
			t.Errorf("expected %q; got %q from case %d: %v", c.key, got, i, c)
		}
	}
}

func TestVersionKey(t *testing.T) {
	var cases = map[string]struct {
		projectName, pipelineName string
		version                   uint64
	}{
		// pre-projects key
		"foo@00001234": {
			pipelineName: "foo",
			version:      1234,
		},
		// post-projects key
		"foo/bar@00001234": {
			projectName:  "foo",
			pipelineName: "bar",
			version:      1234,
		},
	}
	for expected, c := range cases {
		p := &pps.Pipeline{
			Project: &pfs.Project{Name: c.projectName},
			Name:    c.pipelineName,
		}
		if got := VersionKey(p, c.version); expected != got {
			t.Errorf("expected %q but got %q (%v)", expected, got, c)
		}
	}
}

func TestJobsPipelineKey(t *testing.T) {
	var cases = map[string]struct {
		projectName, pipelineName string
	}{
		"foo": {
			pipelineName: "foo",
		},
		"foo/bar": {
			projectName:  "foo",
			pipelineName: "bar",
		},
	}
	for expected, c := range cases {
		var p = &pps.Pipeline{
			Project: &pfs.Project{Name: c.projectName},
			Name:    c.pipelineName,
		}
		if got := JobsPipelineKey(p); expected != got {
			t.Errorf("expected %q but got %q (%v)", expected, got, c)
		}
	}
}

func TestJobTerminalKey(t *testing.T) {
	var cases = map[string]struct {
		projectName, pipelineName string
		isTerminal                bool
	}{
		"foo_true": {
			pipelineName: "foo",
			isTerminal:   true,
		},
		"foo/bar_false": {
			projectName:  "foo",
			pipelineName: "bar",
		},
	}
	for expected, c := range cases {
		var p = &pps.Pipeline{
			Project: &pfs.Project{Name: c.projectName},
			Name:    c.pipelineName,
		}
		if got := JobsTerminalKey(p, c.isTerminal); expected != got {
			t.Errorf("expected %q but got %q (%v)", expected, got, c)
		}
	}
}

func TestJobKey(t *testing.T) {
	var cases = map[string]struct {
		projectName, pipelineName, id string
	}{
		"foo@3333484b130143e586d08e30a8a977ae": {
			pipelineName: "foo",
			id:           "3333484b130143e586d08e30a8a977ae",
		},
		"foo/bar@99b5a418091643a0993785b1df999522": {
			projectName:  "foo",
			pipelineName: "bar",
			id:           "99b5a418091643a0993785b1df999522",
		},
	}
	for expected, c := range cases {
		var j = &pps.Job{
			Pipeline: &pps.Pipeline{
				Project: &pfs.Project{Name: c.projectName},
				Name:    c.pipelineName,
			},
			Id: c.id,
		}
		if got := JobKey(j); expected != got {
			t.Errorf("expected %q but got %q (%v)", expected, got, c)
		}
	}
}

func TestPipelinesNameKey(t *testing.T) {
	var cases = map[string]struct {
		projectName, pipelineName string
	}{
		"foo": {
			pipelineName: "foo",
		},
		"foo/bar": {
			projectName:  "foo",
			pipelineName: "bar",
		},
	}
	for expected, c := range cases {
		var p = &pps.Pipeline{
			Project: &pfs.Project{Name: c.projectName},
			Name:    c.pipelineName,
		}
		if got := PipelinesNameKey(p); expected != got {
			t.Errorf("expected %q but got %q (%v)", expected, got, c)
		}
	}
}
