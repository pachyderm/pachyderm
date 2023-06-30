package shell

import (
	"fmt"
	"testing"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestCommitPatternFormat(t *testing.T) {
	var (
		cases = map[string]bool{
			"source-repos/foo/commits":        true,
			"input-repos/foo/commits":         true,
			"input-repos/foo/bar/commits":     true,
			"input-repos/foo/bar/baz/commits": true,
		}
		glob = fmt.Sprintf(commitPatternFormatString, "*")
	)
	for line, match := range cases {
		g := globlib.MustCompile(glob, '/')
		if g.Match(line) != match {
			t.Errorf("expected %q to be a %v match", line, match)
		}
	}
}

func TestPathnameToRepo(t *testing.T) {
	var cases = map[string]*pfs.Repo{
		"source-repos/foo/commits":         {Name: "foo"},
		"source-repos/foo/bar/commits":     {Project: &pfs.Project{Name: "foo"}, Name: "bar"},
		"source-repos/foo/bar/baz/commits": {Project: &pfs.Project{Name: "foo/bar"}, Name: "baz"},
	}
	for pathname, expected := range cases {
		got := pathnameToRepo(pathname)
		if got.Project.GetName() != expected.Project.GetName() || got.Name != expected.Name {
			t.Errorf("%q failed; expected %v but got %v", pathname, expected, got)
		}
	}
}
