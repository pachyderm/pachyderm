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

// ClockHead returns the head of a FullClock
func ClockHead(fc persist.FullClock) *persist.Clock {
	if len(fc) == 0 {
		return nil
	}
	return fc[len(fc)-1]
}

// ClockBranch determines what branch a given FullClock is on.
// It's the branch of the head.
func ClockBranch(fc persist.FullClock) string {
	head := ClockHead(fc)
	if head == nil {
		return ""
	}
	return head.Branch
}

// NewClock returns a new clock for a given branch
func NewClock(branch string) *persist.Clock {
	return &persist.Clock{branch, 0}
}

// NewChild returns the child of a FullClock
// [(master, 0), (foo, 0)] -> [(master, 0), (foo, 1)]
func NewChild(parent persist.FullClock) persist.FullClock {
	if len(parent) == 0 {
		return parent
	} else {
		lastClock := CloneClock(parent.Head())
		lastClock.Clock += 1
		return append(parent[:len(parent)-1], lastClock)
	}
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

// A ClockRangeList is an ordered list of ClockRanges
type ClockRangeList struct {
	ranges []*ClockRange
}

// A ClockRange represents a range of clocks
type ClockRange struct {
	Branch string
	Left   uint64
	Right  uint64
}

// NewClockRangeList creates a ClockRangeList that represents all clock ranges
// in between the two given FullClocks.
func NewClockRangeList(from persist.FullClock, to persist.FullClock) ClockRangeList {
	var crl ClockRangeList
	crl.AddFullClock(to)
	crl.SubFullClock(from)
	return crl
}

func (l *ClockRangeList) AddFullClock(fc persist.FullClock) {
	for _, c := range fc {
		l.AddClock(c)
	}
}

// AddClock adds a range [0, c.Clock]
func (l *ClockRangeList) AddClock(c *persist.Clock) {
	for _, r := range l.ranges {
		if r.Branch == c.Branch {
			if c.Clock > r.Right {
				r.Right = c.Clock
			}
			return
		}
	}
	l.ranges = append(l.ranges, &ClockRange{
		Branch: c.Branch,
		Left:   0,
		Right:  c.Clock,
	})
}

func (l *ClockRangeList) SubFullClock(fc persist.FullClock) {
	for _, c := range fc {
		l.SubClock(c)
	}
}

// SubClock substracts a range [0, c.Clock]
func (l *ClockRangeList) SubClock(c *persist.Clock) {
	// only keep non-empty ranges
	var newRanges []*ClockRange
	for _, r := range l.ranges {
		if r.Branch == c.Branch {
			r.Left = c.Clock
		}
		if r.Left <= r.Right {
			newRanges = append(newRanges, r)
		}
	}
	l.ranges = newRanges
}

func (l *ClockRangeList) Ranges() []*ClockRange {
	return l.ranges
}
