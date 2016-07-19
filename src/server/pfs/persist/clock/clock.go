package clock

import (
	"github.com/pachyderm/pachyderm/src/server/pfs/persist"
)

// NewBranchClocks creates a new BranchClocks given a branch name
// "master" -> [[(master, 0)]]
func NewBranchClocks(branch string) persist.BranchClocks {
	return nil
}

// NewChild returns the child of a BranchClock
// [(master, 0), (foo, 0)] -> [(master, 0), (foo, 1)]
func NewChild(parent *persist.BranchClock) *persist.BranchClock {
	return nil
}

// NewBranch takes a BranchClock and a branch name, and returns a new BranchClock
// that contains the new Branch
// Returns an error if the branch already exists on the BranchClock
// [(master, 0), (foo, 1)], "bar" -> [(master, 0), (foo, 1), (bar, 0)]
func NewBranch(parent *persist.BranchClock, branch string) (*persist.BranchClock, error) {
	return nil, nil
}

// NewChildOfBranchClocks takes a BranchClocks and a branch name, and returns a
// BranchClocks that contains one BranchClock which is the child of the BranchClock
// specified by the given branch name.
//
// Returns an error if the branch is not found
//
// Args: [[(foo, 1), (bar, 1)], [(master, 1)]], "master"
// Return: [[(master, 2)]], true
func NewChildOfBranchClocks(parent persist.BranchClocks, branch string) (persist.BranchClocks, error) {
	return nil, nil
}

// AddClock adds a BranchClock to a BranchClocks.
// Returns an error if the BranchClock already exists in the BranchClocks
func AddClock(b persist.BranchClocks, c *persist.BranchClock) (persist.BranchClocks, error) {
	return nil, nil
}
