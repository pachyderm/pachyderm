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

// BranchClockToArray converts a BranchClock to an array.
// Putting this function here so it stays in sync with ToArray.
func BranchClockToArray(branchClock gorethink.Term) gorethink.Term {
	return branchClock.Field("Clocks").Map(func(clock gorethink.Term) []interface{} {
		return []interface{}{clock.Field("Branch"), clock.Field("Clock")}
	})
}
