package persist

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dancannon/gorethink"
)

// Size returns the size of a block ref
func (b *BlockRef) Size() uint64 {
	return b.Upper - b.Lower
}

// ReadableCommitID returns a human-friendly commit ID for
// displaying purposes.
func (c *Clock) ReadableCommitID() string {
	return fmt.Sprintf("%s/%d", c.Branch, c.Clock)
}

// NewCommitID generates a commitID to be used in a database
// from a repo and a clock
func NewCommitID(repo string, clock *Clock) string {
	return fmt.Sprintf("%s:%s:%d", repo, clock.Branch, clock.Clock)
}

// NewClock returns a new clock for a given branch
func NewClock(branch string) *Clock {
	return &Clock{branch, 0}
}

// ClockEq returns if two clocks are equal
func ClockEq(c1 *Clock, c2 *Clock) bool {
	return c1.Branch == c2.Branch && c1.Clock == c2.Clock
}

// CloneClock clones a clock
func CloneClock(c *Clock) *Clock {
	return &Clock{
		Branch: c.Branch,
		Clock:  c.Clock,
	}
}

// CloneFullClock clones a FullClock
func CloneFullClock(fc []*Clock) []*Clock {
	var res []*Clock
	for _, c := range fc {
		res = append(res, CloneClock(c))
	}
	return res
}

// StringToClock converts a string like "master/2" to a clock
func StringToClock(s string) (*Clock, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid clock string: %s", s)
	}
	clock, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid clock string: %v", err)
	}
	return &Clock{
		Branch: parts[0],
		Clock:  uint64(clock),
	}, nil
}

// NewChild returns the child of a FullClock
// [(master, 0), (foo, 0)] -> [(master, 0), (foo, 1)]
func NewChild(parent FullClock) FullClock {
	if len(parent) == 0 {
		return parent
	}
	lastClock := CloneClock(FullClockHead(parent))
	lastClock.Clock++
	return append(parent[:len(parent)-1], lastClock)
}

// FullClockParent returns the parent of a full clock, or nil if the clock has no parent
// [(master, 2), (foo, 1)] -> [(master, 2), (foo, 0)]
// [(master, 2), (foo, 0)] -> [(master, 2)]
func FullClockParent(child FullClock) FullClock {
	clone := CloneFullClock(child)
	if len(clone) > 0 {
		lastClock := FullClockHead(clone)
		if lastClock.Clock > 0 {
			lastClock.Clock--
			return clone
		} else if len(child) > 1 {
			return clone[:len(clone)-1]
		}
	}
	return nil
}

// FullClock is an array of clocks, e.g. [(master, 2), (foo, 3)]
type FullClock []*Clock

// FullClockHead returns the last element of a FullClock
func FullClockHead(fc FullClock) *Clock {
	if len(fc) == 0 {
		return nil
	}
	return fc[len(fc)-1]
}

// FullClockBranch returns the branch of the last element of the FullClock
func FullClockBranch(fc FullClock) string {
	return FullClockHead(fc).Branch
}

// FullClockToArray converts a FullClock to an array.
func FullClockToArray(fullClock gorethink.Term) gorethink.Term {
	return fullClock.Map(func(clock gorethink.Term) []interface{} {
		return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
	})
}

// ToArray converts a clock to an array
func (c *Clock) ToArray() []interface{} {
	return []interface{}{c.Branch, c.Clock}
}

// ClockToArray is the same as Clock.ToArray except that it operates on a
// gorethink Term
func ClockToArray(clock gorethink.Term) []interface{} {
	return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
}

// CommitID returns the CommitID of the clock associated with the diff
func (d *Diff) CommitID() string {
	return NewCommitID(d.Repo, d.Clock)
}

// ClockRangeList is an ordered list of ClockRanges
type ClockRangeList struct {
	ranges []*ClockRange
}

// ClockRange represents a range of clocks
type ClockRange struct {
	Branch string
	Left   uint64
	Right  uint64
}

// NewClockRangeList creates a ClockRangeList that represents all clock ranges
// in between the two given FullClocks.
func NewClockRangeList(from FullClock, to FullClock) ClockRangeList {
	var crl ClockRangeList
	crl.AddFullClock(to)
	crl.SubFullClock(from)
	return crl
}

// AddFullClock adds a FullClock to the ClockRange
func (l *ClockRangeList) AddFullClock(fc FullClock) {
	for _, c := range fc {
		l.AddClock(c)
	}
}

// AddClock adds a range [0, c.Clock]
func (l *ClockRangeList) AddClock(c *Clock) {
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

// SubFullClock subtracts a FullClock from the ClockRange
func (l *ClockRangeList) SubFullClock(fc FullClock) {
	for _, c := range fc {
		l.SubClock(c)
	}
}

// SubClock substracts a range [0, c.Clock]
func (l *ClockRangeList) SubClock(c *Clock) {
	// only keep non-empty ranges
	var newRanges []*ClockRange
	for _, r := range l.ranges {
		if r.Branch == c.Branch {
			r.Left = c.Clock + 1
		}
		if r.Left <= r.Right {
			newRanges = append(newRanges, r)
		}
	}
	l.ranges = newRanges
}

// Ranges return the clock ranges stored in a ClockRangeList
func (l *ClockRangeList) Ranges() []*ClockRange {
	return l.ranges
}

// DBClockDescendent returns whether one FullClock is the descendent of the other,
// assuming both are rethinkdb terms
func DBClockDescendent(child, parent gorethink.Term) gorethink.Term {
	return gorethink.Branch(
		gorethink.Or(child.Count().Lt(parent.Count()), parent.Count().Eq(0)),
		gorethink.Expr(false),
		gorethink.Branch(
			child.Count().Eq(parent.Count()),
			gorethink.And(child.Slice(0, -1).Eq(parent.Slice(0, -1)), gorethink.And(child.Nth(-1).Field("Branch").Eq(parent.Nth(-1).Field("Branch")), child.Nth(-1).Gt(parent.Nth(-1)))),
			child.Slice(0, parent.Count()).Eq(parent),
		),
	)
}
