package pfs

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

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
		error: grpc.Errorf(codes.NotFound, "File %v not found in repo %v at commit %v", file, repo, commitID),
	}
}

func NewErrRepoNotFound(repo string) *ErrRepoNotFound {
	return &ErrRepoNotFound{
		error: grpc.Errorf(codes.NotFound, "repo %v not found", repo),
	}
}

func NewErrCommitNotFound(repo string, commitID string) *ErrCommitNotFound {
	return &ErrCommitNotFound{
		error: grpc.Errorf(codes.NotFound, "commit %v not found in repo %v", commitID, repo),
	}
}

func NewErrParentCommitNotFound(repo string, commitID string) *ErrParentCommitNotFound {
	return &ErrParentCommitNotFound{
		error: grpc.Errorf(codes.NotFound, "parent commit %v not found in repo %v", commitID, repo),
	}
}

func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
