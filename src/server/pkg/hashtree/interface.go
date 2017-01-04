package hashtree

import "github.com/pachyderm/pachyderm/src/client/pfs"

type HashTree interface {
	GetFile(path string) *FileNode
	PutFile(path string, blockRefs []*pfs.BlockRef) error
	ListFile(path string) []*FileNode
	GlobFile(pattern string) []*Node
}
