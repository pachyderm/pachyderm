package clock

import (
	"fmt"
	"strconv"
	"strings"

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
		Clocks: []*persist.Clock{NewClock(branch)},
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

func ClockEq(c1 *persist.Clock, c2 *persist.Clock) bool {
	return c1.Branch == c2.Branch && c1.Clock == c2.Clock
}

func CloneClock(c *persist.Clock) *persist.Clock {
	return &persist.Clock{
		Branch: c.Branch,
		Clock:  c.Clock,
	}
}

func CloneBranchClock(b *persist.BranchClock) *persist.BranchClock {
	res := &persist.BranchClock{}
	if b != nil {
		for _, clock := range b.Clocks {
			res.Clocks = append(res.Clocks, CloneClock(clock))
		}
	}
	return res
}

// GetClockIntervals return an array of clock intervals that describe all clocks
// in between two given clocks.
// Example:
// Given two clocks: [(master, 1)] and [(master, 3), (foo, 2)]
// Return: [[[(master, 2)], [(master, 3)]], [[(master, 3), (foo, 0)], [(master, 3), (foo, 2)]]]
func GetClockIntervals(left *persist.BranchClock, right *persist.BranchClock) ([][]*persist.BranchClock, error) {
	current := CloneBranchClock(left)
	var intervals [][]*persist.BranchClock
	for i := 0; i < len(right.Clocks); i++ {
		var leftClone *persist.BranchClock
		if len(current.Clocks) < i+1 {
			current.Clocks = append(current.Clocks, CloneClock(right.Clocks[i]))
			leftClone = CloneBranchClock(current)
			leftClone.Clocks[i].Clock = 0
		} else if !ClockEq(current.Clocks[i], right.Clocks[i]) || (ClockEq(current.Clocks[i], right.Clocks[i]) && i == len(right.Clocks)-1) {
			if current.Clocks[i].Branch != right.Clocks[i].Branch || current.Clocks[i].Clock > right.Clocks[i].Clock {
				return nil, fmt.Errorf("clocks %s is not an ancestor of %s", left, right)
			}
			leftClone = CloneBranchClock(current)
			leftClone.Clocks = leftClone.Clocks[:i+1]
			leftClone.Clocks[i].Clock += 1
			current.Clocks[i] = right.Clocks[i]
		} else {
			continue
		}
		rightClone := CloneBranchClock(right)
		rightClone.Clocks = rightClone.Clocks[:i+1]
		intervals = append(intervals, []*persist.BranchClock{
			leftClone,
			rightClone,
		})
	}
	return intervals, nil
}

// "master/2"
func StringToClock(s string) (*persist.Clock, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid clock string: %s", s)
	}
	clock, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid clock string: %v", err)
	}
	return &persist.Clock{
		Branch: parts[0],
		Clock:  uint64(clock),
	}, nil
}

// "master/2-foo/3"
func StringToBranchClock(s string) (*persist.BranchClock, error) {
	res := &persist.BranchClock{}
	parts := strings.Split(s, "-")
	for _, part := range parts {
		clock, err := StringToClock(part)
		if err != nil {
			return nil, err
		}
		res.Clocks = append(res.Clocks, clock)
	}
	return res, nil
}
