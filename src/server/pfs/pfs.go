package pfs

import (
	"fmt"
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

// Retuns nothing / doesn't modify if it isn't one of the types listed above
func ErrToCodedGrpcError(e error) error {
	if e == nil {
		return nil
	}
	switch recognizedErr := e.(type) {
	case ErrFileNotFound:
		return grpc.Errorf(codes.NotFound, recognizedErr.Error())
	case ErrRepoNotFound:
		return grpc.Errorf(codes.NotFound, recognizedErr.Error())
	case ErrCommitNotFound:
		return grpc.Errorf(codes.NotFound, recognizedErr.Error())
	case ErrParentCommitNotFound:
		return grpc.Errorf(codes.NotFound, recognizedErr.Error())
	}
	return e
}

func NewErrFileNotFound(file string, repo string, commitID string) *ErrFileNotFound {
	return &ErrFileNotFound{
		error: fmt.Errorf("File %v not found in repo %v at commit %v", file, repo, commitID),
	}
}

func NewErrRepoNotFound(repo string) *ErrRepoNotFound {
	return &ErrRepoNotFound{
		error: fmt.Errorf("Repo %v not found", repo),
	}
}

func NewErrCommitNotFound(repo string, commitID string) *ErrCommitNotFound {
	return &ErrCommitNotFound{
		error: fmt.Errorf("Commit %v not found in repo %v", commitID, repo),
	}
}

func NewErrParentCommitNotFound(repo string, commitID string) *ErrParentCommitNotFound {
	return &ErrParentCommitNotFound{
		error: fmt.Errorf("Parent commit %v not found in repo %v", commitID, repo),
	}
}

func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
