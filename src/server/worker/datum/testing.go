package datum

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

// A MockIterator can be used in tests to fulfill the Iterator interface without
// having PFS running.  It returns Inputs based on the number range selected.
type MockIterator struct {
	start  int
	end    int
	index  int
	inputs int
}

// MockIteratorOptions specifies the behavior of the MockIterator.
type MockIteratorOptions struct {
	Start  uint
	Length uint
	Inputs uint
}

// NewMockIterator constructs a mock iterator that will consistently return
// Inputs based on the current index.
func NewMockIterator(options *MockIteratorOptions) *MockIterator {
	result := &MockIterator{
		start:  int(options.Start),
		end:    int(options.Start + options.Length),
		index:  int(options.Start) - 1,
		inputs: int(options.Inputs),
	}

	if result.inputs == 0 {
		result.inputs = 1
	}

	return result
}

// Reset returns the current index to the start of the Iterator.
func (mi *MockIterator) Reset() {
	mi.index = mi.start - 1
}

// Len returns the number of items in the Iterator.
func (mi *MockIterator) Len() int {
	return int(mi.end - mi.start)
}

// Next steps the Iterator to the next item.
func (mi *MockIterator) Next() bool {
	if mi.index < mi.end {
		mi.index++
	}
	return mi.index < mi.end
}

// Datum returns the set of Inputs for the current index of the Iterator.
func (mi *MockIterator) Datum() []*common.Input {
	return mi.DatumN(mi.index - mi.start)
}

// DatumN returns the set of Inputs for the selected index in the Iterator.
func (mi *MockIterator) DatumN(index int) []*common.Input {
	index = mi.start + index
	if index < mi.start || index > mi.end {
		panic(fmt.Sprintf("out of bounds access in MockIterator: index %d of range %d-%d", index, mi.start, mi.end))
	}

	result := []*common.Input{}
	for i := 0; i < mi.inputs; i++ {
		// Warning: this might break some assumptions about the format of this data
		result = append(result, &common.Input{
			FileInfo: []*pfs.FileInfo{{
				File:      client.NewFile("dummy-repo", "dummy-commit", "/path"),
				FileType:  pfs.FileType_FILE,
				SizeBytes: uint64(index),
				Hash:      []byte(fmt.Sprintf("%d", index)),
			}},
			Name: fmt.Sprintf("source-%d", i),
		})
	}
	return result
}
