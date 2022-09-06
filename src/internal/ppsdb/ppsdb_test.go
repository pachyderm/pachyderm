package ppsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
		"pipeline@0123456789ab40123456789abcdef012": {
			isError:      false,
			projectName:  "",
			pipelineName: "pipeline",
			id:           "0123456789ab40123456789abcdef012",
		},
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

func TestCommitKey(t *testing.T) {
	var cases = []struct {
		isError     bool
		projectName string
		repoName    string
		id          string
		key         string
	}{
		// new style
		{false, "project", "repo", "0123456789ab40123456789abcdef012", "project/repo@0123456789ab40123456789abcdef012"},
		// old style
		{false, "", "repo", "0123456789ab40123456789abcdef012", "repo@0123456789ab40123456789abcdef012"},
	}
	for i, c := range cases {
		if got, err := ppsdb.PipelineCommitKey(&pfs.Commit{
			Branch: &pfs.Branch{
				Repo: &pfs.Repo{
					Project: &pfs.Project{
						Name: c.projectName,
					},
					Name: c.repoName,
					Type: pfs.SpecRepoType,
				},
			},
			ID: c.id,
		}); err != nil {
			t.Errorf("unexpected error with test case %d: %v", i, c)
		} else if got != c.key {
			t.Errorf("expected %q; got %q from case %d: %v", c.key, got, i, c)
		}
	}
}
