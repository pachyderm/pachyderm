package persist

import (
	"fmt"

	"github.com/dancannon/gorethink"
)

func (b *BlockRef) Size() uint64 {
	return b.Upper - b.Lower
}

// FullClock is an array of clocks, e.g. [(master, 2), (foo, 3)]
type FullClock []*Clock

func (fc FullClock) Size() int {
	return len(fc)
}

// ToArray converts a FullClock to an array of arrays.
// This is useful in indexing BranchClocks in RethinkDB.
func (fc FullClock) ToArray() (res []interface{}) {
	for _, clock := range fc {
		res = append(res, []interface{}{clock.Branch, clock.Clock})
	}
	return res
}

func (fc FullClock) Head() *Clock {
	if len(fc) == 0 {
		return nil
	}
	return fc[len(fc)-1]
}

// BranchClockToArray converts a BranchClock to an array.
// Putting this function here so it stays in sync with ToArray.
func FullClockToArray(fullClock gorethink.Term) gorethink.Term {
	return fullClock.Map(func(clock gorethink.Term) []interface{} {
		return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
	})
}

func (c *Clock) ToArray() []interface{} {
	return []interface{}{c.Branch, c.Clock}
}

func ClockToArray(clock gorethink.Term) []interface{} {
	return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
}

func (c *Clock) ToCommitID() string {
	return fmt.Sprintf("%s/%d", c.Branch, c.Clock)
}

func (d *Diff) CommitID() string {
	return d.Clock.ToCommitID()
}
