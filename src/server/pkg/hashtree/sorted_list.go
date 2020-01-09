package hashtree

import "sort"

// This is not in "math", unfortunately
func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

// insertStr inserts 's' into 'ss', preserving sorting (it assumes that 'ss' is
// sorted). If a copy is necessary (because cap(ss) is too small), only does
// one copy (unlike `append(append(ss[:idx], newS), ss[idx:]...)`). This is
// because directory nodes may include a large number of children.
//
// This is used to preserve the order in DirectoryNode.Children, which must
// be maintained so that equivalent directories have the same hash.
//
// Returns 'true' if newS was added to 'ss' and 'false' otherwise (if newS is
// already in 'ss').
func insertStr(ss *[]string, newS string) bool {
	sz := cap(*ss)
	idx := sort.SearchStrings(*ss, newS)
	if idx >= len(*ss) || (*ss)[idx] != newS {
		// Need to insert new element
		if sz >= (len(*ss) + 1) {
			*ss = (*ss)[:len(*ss)+1]
			copy((*ss)[idx+1:], (*ss)[idx:])
			(*ss)[idx] = newS
		} else {
			// Need to grow ss (i.e. make a copy)
			// - Using a factor (instead of always adding a constant number of
			//   elements) ensures amortized constant time for insertions, and keeping
			//   it reasonably low avoids wasting too much space. Current value of
			//   1.33 is arbitrary.
			// - If sz is small, grow sz by at least a constant amount (must be >=1,
			//   currently 10)
			cap1, cap2 := int(float64(sz)*1.33), sz+10
			newSs := make([]string, len(*ss)+1, max(cap1, cap2))
			copy(newSs, (*ss)[:idx])
			copy(newSs[idx+1:], (*ss)[idx:])
			newSs[idx] = newS
			*ss = newSs
		}
		return true
	}
	return false
}
