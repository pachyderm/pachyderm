package files

import "github.com/pachyderm/pachyderm/src/server/pkg/storage/obj"

func Merge(in, out obj.Storage, ids []string) {
	// Walk through file/chunk index.
	// When implementing sliding window hashing,
	// we need to logically pad zeroes at file boundaries to avoid
	// earlier files from effecting the hash of later files (within the window).
}
