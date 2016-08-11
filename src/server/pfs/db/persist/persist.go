package persist

import "github.com/dancannon/gorethink"

type BranchClocks []*BranchClock

func (b *BlockRef) Size() uint64 {
	return b.Upper - b.Lower
}

// ToArray converts a BranchClock to an array.
// This is useful in indexing BranchClocks in RethinkDB.
func (b *BranchClock) ToArray() (res []interface{}) {
	for _, clock := range b.Clocks {
		res = append(res, []interface{}{clock.Branch, clock.Clock})
	}
	return res
}

func (b *BranchClock) Head() *Clock {
	if len(b.Clocks) == 0 {
		return nil
	}
	return b.Clocks[len(b.Clocks)-1]
}

func Heads(bs BranchClocks) []*Clock {
	var res []*Clock
	for _, b := range bs {
		res = append(res, b.Head())
	}
	return res
}

// BranchClockToArray converts a BranchClock to an array.
// Putting this function here so it stays in sync with ToArray.
func BranchClockToArray(branchClock gorethink.Term) gorethink.Term {
	return branchClock.Field("Clocks").Map(func(clock gorethink.Term) []interface{} {
		return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
	})
}

func (c *Clock) ToArray() []interface{} {
	return []interface{}{c.Branch, c.Clock}
}

func ClockToArray(clock gorethink.Term) []interface{} {
	return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
}
