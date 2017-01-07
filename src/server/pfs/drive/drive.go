/*
Package drive provides the definitions for the low-level pfs storage drivers.
*/
package drive

import (
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ListFileMode specifies how ListFile executes.
type ListFileMode int

const (
	// ListFileNORMAL computes sizes for files but not for directories
	ListFileNORMAL ListFileMode = iota
	// ListFileFAST does not compute sizes for files or directories
	ListFileFAST
	// ListFileRECURSE computes sizes for files and directories
	ListFileRECURSE
)

// IsPermissionError returns true if a given error is a permission error.
func IsPermissionError(err error) bool {
	return strings.Contains(err.Error(), "has already finished")
}

// Driver represents a low-level pfs storage driver.
type Driver interface {
	CreateRepo(repo *pfs.Repo, provenance []*pfs.Repo) error
	InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error)
	ListRepo(provenance []*pfs.Repo) ([]*pfs.RepoInfo, error)
	DeleteRepo(repo *pfs.Repo, force bool) error

	StartCommit(parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error)
	FinishCommit(commit *pfs.Commit, cancel bool) error
	// Squash merges the content of fromCommits into a single commit with
	// the given parent.
	SquashCommit(fromCommits []*pfs.Commit, parent *pfs.Commit) (*pfs.Commit, error)
	InspectCommit(commit *pfs.Commit) (*pfs.CommitInfo, error)
	ListCommit(from *pfs.Commit, to *pfs.Commit) ([]*pfs.CommitInfo, error)
	FlushCommit(fromCommits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error)
	DeleteCommit(commit *pfs.Commit) error

	ListBranch(repo *pfs.Repo) ([]string, error)
	MakeBranch(repo *pfs.Repo, commit *pfs.Commit, name string) error
	RenameBranch(repo *pfs.Repo, from string, to string)

	PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) error
	MakeDirectory(file *pfs.File) error
	GetFile(file *pfs.File, offset int64, size int64) (io.ReadCloser, error)
	InspectFile(file *pfs.File) (*pfs.FileInfo, error)
	ListFile(file *pfs.File) ([]*pfs.FileInfo, error)
	DeleteFile(file *pfs.File) error

	DeleteAll() error
	Dump()
}
