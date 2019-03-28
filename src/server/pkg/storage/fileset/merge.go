package fileset

import "github.com/pachyderm/pachyderm/src/server/pkg/obj"

const (
	MergeLog   = "mergelog"
	Checkpoint = "checkpoint"
)

// CompactMergeLog will compact the merge log of path based on heuristics.
// This may involve overwriting path with the compacted fileset and writing a new merge log with
// the filesets that were not compacted.
// - May not want to overwrite path and instead store it as a checkpoint, that can easily be cleaned
// up later, this would make getting the content of later commits a longer set of merge/un-merge operations.
func CompactMergeLog(client *obj.Client, path, parentPath string) {
	// Check merge log of parent and apply heuristics to determine if we should merge anything.
	// These heuristics should be a function of number of objects and size of objects.
	// We should attempt to merge the smallest objects first when there are many objects.
}

func Merge(w Writer, rs []Reader) {
	// Walk through file index, each file header will have chunk ref (may move to only the first file header in chunk having a chunk ref)
	// When implementing sliding window hashing,
	// we need to logically pad zeroes at file boundaries to avoid
	// earlier files from effecting the hash of later files (within the window).
	// Apply operations to files in the order in which they are passed into the merge.
	//   Will need to decide what this means in terms of datums/reverse index (maybe store identifier
	//   with DataOp that is used as precedence when determining order)
}
