package clock

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/server/pfs/persist"
)

type ErrBranchExists struct {
	error
}

type ErrBranchNotFound struct {
	error
}

// NewBranchClocks creates a new BranchClocks given a branch name
// "master" -> [[(master, 0)]]
func NewBranchClocks(branch string) persist.BranchClocks {
	return persist.BranchClocks{{
		Clocks: []*persist.Clock{{
			Branch: branch,
			Clock:  0,
		}},
	}}
}

// NewChild returns the child of a BranchClock
// [(master, 0), (foo, 0)] -> [(master, 0), (foo, 1)]
func NewChild(parent *persist.BranchClock) *persist.BranchClock {
	if len(parent.Clocks) == 0 {
		return parent
	} else {
		parent.Clocks[len(parent.Clocks)-1].Clock += 1
		return parent
	}
}

// NewBranch takes a BranchClock and a branch name, and returns a new BranchClock
// that contains the new Branch
// Returns an error if the branch already exists on the BranchClock
// [(master, 0), (foo, 1)], "bar" -> [(master, 0), (foo, 1), (bar, 0)]
func NewBranch(parent *persist.BranchClock, branch string) (*persist.BranchClock, error) {
	for _, clock := range parent.Clocks {
		if clock.Branch == branch {
			return nil, ErrBranchExists{
				error: fmt.Errorf("branch %s already exists in the branch clock", branch),
			}
		}
	}
	parent.Clocks = append(parent.Clocks, &persist.Clock{
		Branch: branch,
		Clock:  0,
	})
	return parent, nil
}

// NewChildOfBranchClocks takes a BranchClocks and a branch name, and returns a
// BranchClocks that contains one BranchClock which is the child of the BranchClock
// specified by the given branch name.
//
// Returns an error if the branch is not found
//
// Args: [[(foo, 1), (bar, 1)], [(master, 1)]], "master"
// Return: [[(master, 2)]]
func NewChildOfBranchClocks(parent persist.BranchClocks, branch string) (persist.BranchClocks, error) {
	for _, branchClock := range parent {
		if len(branchClock.Clocks) > 0 && branchClock.Clocks[len(branchClock.Clocks)-1].Branch == branch {
			return persist.BranchClocks{NewChild(branchClock)}, nil
		}
	}
	return nil, ErrBranchNotFound{fmt.Errorf("branch %s not found in branch clocks")}
}

// AddClock adds a BranchClock to a BranchClocks.
// Returns an error if the branch already exists in the BranchClocks
//
// Args: [[(foo, 1), (bar, 1)], [(master, 1)]], [(buzz, 2)]
// Return: [[(foo, 1), (bar, 1)], [(master, 1)], [(buzz, 2)]]
func AddClock(b persist.BranchClocks, c *persist.BranchClock) (persist.BranchClocks, error) {
	if c == nil || len(c.Clocks) == 0 {
		return b, nil
	}

	branch := c.Clocks[len(c.Clocks)-1].Branch

	for _, bc := range b {
		if len(bc.Clocks) > 0 && bc.Clocks[len(bc.Clocks)-1].Branch == branch {
			return nil, ErrBranchExists{fmt.Errorf("branch %s already exists in the branch clock", branch)}
		}
	}

	return append(b, c), nil
}
