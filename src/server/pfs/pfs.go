package pfs

import (
	"fmt"
	"regexp"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// ErrFileNotFound represents a file-not-found error.
type ErrFileNotFound struct {
	File *pfs.File
}

// ErrRepoNotFound represents a repo-not-found error.
type ErrRepoNotFound struct {
	Repo *pfs.Repo
}

// ErrRepoExists represents a repo-exists error.
type ErrRepoExists struct {
	Repo *pfs.Repo
}

// ErrCommitNotFound represents a commit-not-found error.
type ErrCommitNotFound struct {
	Commit *pfs.Commit
}

// ErrNoHead represents an error encountered because a branch has no head (e.g.
// inspectCommit(master) when 'master' has no commits)
type ErrNoHead struct {
	Branch *pfs.Branch
}

// ErrCommitExists represents an error where the commit already exists.
type ErrCommitExists struct {
	Commit *pfs.Commit
}

// ErrCommitFinished represents an error where the commit has been finished
// (e.g from PutFile or DeleteFile)
type ErrCommitFinished struct {
	Commit *pfs.Commit
}

// ErrCommitDeleted represents an error where the commit has been deleted (e.g.
// from InspectCommit)
type ErrCommitDeleted struct {
	Commit *pfs.Commit
}

// ErrParentCommitNotFound represents a parent-commit-not-found error.
type ErrParentCommitNotFound struct {
	Commit *pfs.Commit
}

// ErrOutputCommitNotFinished represents an error where the commit has not
// been finished
type ErrOutputCommitNotFinished struct {
	Commit *pfs.Commit
}

// ErrCommitNotFinished represents an error where the commit has not been finished.
type ErrCommitNotFinished struct {
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

func (e ErrNoHead) Error() string {
	// the dashboard is matching on this message in stats. Please open an issue on the dash before changing this
	return fmt.Sprintf("the branch \"%s\" has no head (create one with 'start commit')", e.Branch.Name)
}

func (e ErrCommitExists) Error() string {
	return fmt.Sprintf("commit %v already exists in repo %v", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrCommitFinished) Error() string {
	return fmt.Sprintf("commit %v in repo %v has already finished", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrCommitDeleted) Error() string {
	return fmt.Sprintf("commit %v/%v was deleted", e.Commit.Repo.Name, e.Commit.ID)
}

func (e ErrParentCommitNotFound) Error() string {
	return fmt.Sprintf("parent commit %v not found in repo %v", e.Commit.ID, e.Commit.Repo.Name)
}

func (e ErrOutputCommitNotFinished) Error() string {
	return fmt.Sprintf("output commit %v not finished", e.Commit.ID)
}

func (e ErrCommitNotFinished) Error() string {
	return fmt.Sprintf("commit %v not finished", e.Commit.ID)
}

// ByteRangeSize returns byteRange.Upper - byteRange.Lower.
func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}

var (
	commitNotFoundRe          = regexp.MustCompile("commit [^ ]+ not found in repo [^ ]+")
	commitDeletedRe           = regexp.MustCompile("commit [^ ]+/[^ ]+ was deleted")
	commitFinishedRe          = regexp.MustCompile("commit [^ ]+ in repo [^ ]+ has already finished")
	repoNotFoundRe            = regexp.MustCompile(`repos/ ?[a-zA-Z0-9.\-_]{1,255} not found`)
	repoExistsRe              = regexp.MustCompile(`repo ?[a-zA-Z0-9.\-_]{1,255} already exists`)
	branchNotFoundRe          = regexp.MustCompile(`branches/[a-zA-Z0-9.\-_]{1,255}/ [^ ]+ not found`)
	fileNotFoundRe            = regexp.MustCompile(`file .+ not found`)
	hasNoHeadRe               = regexp.MustCompile(`the branch .+ has no head \(create one with 'start commit'\)`)
	outputCommitNotFinishedRe = regexp.MustCompile("output commit .+ not finished")
	commitNotFinishedRe       = regexp.MustCompile("commit .+ not finished")
)

// IsCommitNotFoundErr returns true if 'err' has an error message that matches
// ErrCommitNotFound
func IsCommitNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return commitNotFoundRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}

// IsCommitDeletedErr returns true if 'err' has an error message that matches
// ErrCommitDeleted
func IsCommitDeletedErr(err error) bool {
	if err == nil {
		return false
	}
	return commitDeletedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}

// IsCommitFinishedErr returns true of 'err' has an error message that matches
// ErrCommitFinished
func IsCommitFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return commitFinishedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}

// IsRepoNotFoundErr returns true if 'err' is an error message about a repo
// not being found
func IsRepoNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return repoNotFoundRe.MatchString(err.Error())
}

// IsRepoExistsErr returns true if 'err' is an error message about a repo
// existing
func IsRepoExistsErr(err error) bool {
	if err == nil {
		return false
	}
	return repoExistsRe.MatchString(err.Error())
}

// IsBranchNotFoundErr returns true if 'err' is an error message about a
// branch not being found
func IsBranchNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return branchNotFoundRe.MatchString(err.Error())
}

// IsFileNotFoundErr returns true if 'err' is an error message about a PFS
// file not being found
func IsFileNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return fileNotFoundRe.MatchString(err.Error())
}

// IsNoHeadErr returns true if the err is due to an operation that cannot be
// performed on a headless branch
func IsNoHeadErr(err error) bool {
	if err == nil {
		return false
	}
	return hasNoHeadRe.MatchString(err.Error())
}

// IsOutputCommitNotFinishedErr returns true if the err is due to an operation
// that cannot be performed on an unfinished output commit
func IsOutputCommitNotFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return outputCommitNotFinishedRe.MatchString(err.Error())
}

// IsCommitNotFinishedErr returns true if the err is due to an attempt at performing
// an operation that only applies to finished commits on an unfinished commit.
func IsCommitNotFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return commitNotFinishedRe.MatchString(err.Error())
}
