package pps

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"

	"gopkg.in/src-d/go-git.v4"
)

// VisitInput visits each input recursively in ascending order (root last)
func VisitInput(input *Input, f func(*Input)) {
	switch {
	case input.Cross != nil:
		for _, input := range input.Cross {
			VisitInput(input, f)
		}
	case input.Union != nil:
		for _, input := range input.Union {
			VisitInput(input, f)
		}
	}
	f(input)
}

// InputName computes the name of an Input.
func InputName(input *Input) string {
	switch {
	case input.Atom != nil:
		return input.Atom.Name
	case input.Cross != nil:
		if len(input.Cross) > 0 {
			return InputName(input.Cross[0])
		}
	case input.Union != nil:
		if len(input.Union) > 0 {
			return InputName(input.Union[0])
		}
	}
	return ""
}

// SortInput sorts an Input.
func SortInput(input *Input) {
	VisitInput(input, func(input *Input) {
		SortInputs := func(inputs []*Input) {
			sort.SliceStable(inputs, func(i, j int) bool { return InputName(inputs[i]) < InputName(inputs[j]) })
		}
		switch {
		case input.Cross != nil:
			SortInputs(input.Cross)
		case input.Union != nil:
			SortInputs(input.Union)
		}
	})
}

// InputCommits returns the commits in an Input.
func InputCommits(input *Input) []*pfs.Commit {
	var result []*pfs.Commit
	VisitInput(input, func(input *Input) {
		if input.Atom != nil {
			result = append(result, &pfs.Commit{
				Repo: &pfs.Repo{input.Atom.Repo},
				ID:   input.Atom.Commit,
			})
		}
		if input.Cron != nil {
			result = append(result, &pfs.Commit{
				Repo: &pfs.Repo{input.Cron.Repo},
				ID:   input.Cron.Commit,
			})
		}
		if input.Git != nil {
			result = append(result, &pfs.Commit{
				Repo: &pfs.Repo{input.Git.Name},
				ID:   input.Git.Commit,
			})
		}
	})
	return result
}

// ValidateGitCloneURL returns an error if the provided URL is invalid
func ValidateGitCloneURL(url string) error {
	exampleURL := "https://github.com/org/foo.git"
	if url == "" {
		return fmt.Errorf("clone URL is missing (example clone URL %v)", exampleURL)
	}
	// Use the git client's validator to make sure its a valid URL
	o := &git.CloneOptions{
		URL: url,
	}
	if err := o.Validate(); err != nil {
		return err
	}

	// Make sure its the type that we want. Of the following we
	// only accept the 'clone' type of url:
	//     git_url: "git://github.com/sjezewski/testgithook.git",
	//     ssh_url: "git@github.com:sjezewski/testgithook.git",
	//     clone_url: "https://github.com/sjezewski/testgithook.git",
	//     svn_url: "https://github.com/sjezewski/testgithook",
	invalidErr := fmt.Errorf("clone URL is missing .git suffix (example clone URL %v)", exampleURL)

	if !strings.HasSuffix(url, ".git") {
		// svn_url case
		return invalidErr
	}
	if !strings.HasPrefix(url, "https://") {
		// git_url or ssh_url cases
		return invalidErr
	}

	return nil
}
