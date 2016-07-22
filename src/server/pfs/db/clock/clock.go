package clock

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
)

type ErrBranchExists struct {
	error
}

type ErrBranchNotFound struct {
	error
}

// NewClock returns a new clock for a given branch
func NewClock(branch string) *persist.Clock {
	return &persist.Clock{branch, 0}
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
//
// Returns an error if the branch already exists on the BranchClock
//
// Args: [(master, 0), (foo, 1)], "bar"
// Return: [(master, 0), (foo, 1), (bar, 0)]
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

// NewBranchOffBranchClocks takes a BranchClocks, a branch name A, a branch name B,
// and returns a BranchClocks that contains one BranchClock corresponding to A
// branching off to B.

// Returns an error if the branch is not found
//
// Args: [[(foo, 1), (bar, 1)], [(master, 1)]], "master", "buzz"
// Return: [[(master, 2), (buzz, 0)]]
func NewBranchOffBranchClocks(parent persist.BranchClocks, branchA string, branchB string) (persist.BranchClocks, error) {
	for _, branchClock := range parent {
		if len(branchClock.Clocks) > 0 && branchClock.Clocks[len(branchClock.Clocks)-1].Branch == branchA {
			branchClock.Clocks = append(branchClock.Clocks, NewClock(branchB))
			return persist.BranchClocks{branchClock}, nil
		}
	}
	return nil, ErrBranchNotFound{fmt.Errorf("branch %s not found in branch clocks %v", branchA, parent)}
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
	return nil, ErrBranchNotFound{fmt.Errorf("branch %s not found in branch clocks %v", branch, parent)}
}

// AddClock adds a BranchClock to a BranchClocks.
//
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

// GetClockForBranch takes a BranchClocks and a branch name, and return the clock
// that specifies the position on that branch.
//
// Returns an error if the branch is not found
//
// Args: [[(foo, 1), (bar, 1)], [(master, 1)]], "bar"
// Return: (bar, 1)
func GetClockForBranch(clocks persist.BranchClocks, branch string) (*persist.Clock, error) {
	for _, bc := range clocks {
		clock := lastComponent(bc)
		if clock.Branch == branch {
			return clock, nil
		}
	}
	return nil, ErrBranchNotFound{fmt.Errorf("branch %s not found in branch clocks")}
}

func lastComponent(bc *persist.BranchClock) *persist.Clock {
	if len(bc.Clocks) == 0 {
		return nil
	}
	return bc.Clocks[len(bc.Clocks)-1]
}

// GetBranchNameFromBranchClock takes a BranchClock and returns the branch name
// Args: [(foo,0), (bar, 1)]
// Return: bar
func GetBranchNameFromBranchClock(b *persist.BranchClock) string {
	clock := lastComponent(b)
	return clock.Branch
}
