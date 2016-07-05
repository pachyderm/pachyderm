package pfs

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type ErrFileNotFound struct {
	error
}

type ErrRepoNotFound struct {
	error
}

type ErrCommitNotFound struct {
	error
}

type ErrParentCommitNotFound struct {
	error
}

func NewErrFileNotFound(file string, repo string, commitID string) *ErrFileNotFound {
	return &ErrFileNotFound{
		error: fmt.Errorf("file %v not found in repo %v at commit %v", file, repo, commitID),
	}
}

func NewErrRepoNotFound(repo string) *ErrRepoNotFound {
	return &ErrRepoNotFound{
		error: fmt.Errorf("repo %v not found", repo),
	}
}

func NewErrCommitNotFound(repo string, commitID string) *ErrCommitNotFound {
	return &ErrCommitNotFound{
		error: fmt.Errorf("commit %v not found in repo %v", commitID, repo),
	}
}

func NewErrParentCommitNotFound(repo string, commitID string) *ErrParentCommitNotFound {
	return &ErrParentCommitNotFound{
		error: fmt.Errorf("parent commit %v not found in repo %v", commitID, repo),
	}
}

func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
