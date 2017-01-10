package pfs

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ErrFileNotFound represents a file-not-found error.
type ErrFileNotFound struct {
	File *pfs.File
}

// ErrRepoNotFound represents a repo-not-found error.
type ErrRepoNotFound struct {
	Repo *pfs.Repo
}

// ErrRepoNotFound represents a repo-not-found error.
type ErrRepoExists struct {
	Repo *pfs.Repo
}

// ErrCommitNotFound represents a commit-not-found error.
type ErrCommitNotFound struct {
	Commit *pfs.Commit
}

// ErrCommitExists represents an error where the commit already exists.
type ErrCommitExists struct {
	Commit *pfs.Commit
}

// ErrCommitFinished represents an error where the commit has been finished.
type ErrCommitFinished struct {
	Commit *pfs.Commit
}

// ErrParentCommitNotFound represents a parent-commit-not-found error.
type ErrParentCommitNotFound struct {
	Commit *pfs.Commit
}

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("file %v not found in repo %v at commit %v", e.File.Path, e.File.Commit.Repo.Name, e.File.Commit.ID)
}

func (e ErrRepoNotFound) Error() string {
	return fmt.Sprintf("repo %v not found", e.Repo.Name)
}

func (e ErrRepoExists) Error() string {
	return fmt.Sprintf("repo %v already exists", e.Repo.Name)
}

func (e ErrCommitNotFound) Error() string {
	return fmt.Sprintf("commit %v not found in repo %v", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrCommitExists) Error() string {
	return fmt.Sprintf("commit %v already exists in repo %v", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrCommitFinished) Error() string {
	return fmt.Sprintf("commit %v in repo %v has already finished", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrParentCommitNotFound) Error() string {
	return fmt.Sprintf("parent commit %v not found in repo %v", e.Commit.ID, e.Commit.Repo.Name)
}

// ByteRangeSize returns byteRange.Upper - byteRange.Lower.
func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
