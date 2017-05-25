// Package input provides helper functions that operate on pps.Input
package input

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// visit each input recursively in ascending order (root last)
func Visit(input *pps.Input, f func(*pps.Input)) {
	switch {
	case input.Cross != nil:
		for _, input := range input.Cross {
			Visit(input, f)
		}
	case input.Union != nil:
		for _, input := range input.Union {
			Visit(input, f)
		}
	}
	f(input)
}

func Name(input *pps.Input) string {
	switch {
	case input.Atom != nil:
		return input.Atom.Name
	case input.Cross != nil:
		if len(input.Cross) > 0 {
			return Name(input.Cross[0])
		}
	case input.Union != nil:
		if len(input.Union) > 0 {
			return Name(input.Union[0])
		}
	}
	return ""
}

func SortInput(input *pps.Input) {
	Visit(input, func(input *pps.Input) {
		sortInputs := func(inputs []*pps.Input) {
			sort.SliceStable(inputs, func(i, j int) bool { return Name(inputs[i]) < Name(inputs[j]) })
		}
		switch {
		case input.Cross != nil:
			sortInputs(input.Cross)
		case input.Union != nil:
			sortInputs(input.Union)
		}
	})
}

func InputCommits(input *pps.Input) []*pfs.Commit {
	var result []*pfs.Commit
	Visit(input, func(input *pps.Input) {
		if input.Atom != nil {
			result = append(result, client.NewCommit(input.Atom.Repo, input.Atom.Commit))
		}
	})
	return result
}
