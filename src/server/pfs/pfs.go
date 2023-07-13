package pfs

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
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

// ErrBranchNotFound represents a branch-not-found error.
type ErrBranchNotFound struct {
	Branch *pfs.Branch
}

// ErrBranchExists represents a branch-exists error.
type ErrBranchExists struct {
	Branch *pfs.Branch
}

// ErrProjectNotFound represents a project-not-found error.
type ErrProjectNotFound struct {
	Project *pfs.Project
}

// ErrProjectExists represents a project-exists error.
type ErrProjectExists struct {
	Project *pfs.Project
}

// ErrCommitNotFound represents a commit-not-found error.
type ErrCommitNotFound struct {
	Commit *pfs.Commit
}

// ErrCommitSetNotFound represents a commitset-not-found error.
type ErrCommitSetNotFound struct {
	CommitSet *pfs.CommitSet
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

// ErrCommitError represents an error where the commit has been finished with an error.
type ErrCommitError struct {
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

// ErrBaseCommitNotFinished represents an error where a base commit has not been finished.
type ErrBaseCommitNotFinished struct {
	BaseCommit *pfs.Commit
	Commit     *pfs.Commit
}

// ErrInconsistentCommit represents an error where a transaction attempts to
// create a CommitSet with multiple commits in the same branch, which would
// result in inconsistent data dependencies.
type ErrInconsistentCommit struct {
	Commit *pfs.Commit
}

func (e ErrInconsistentCommit) Is(other error) bool {
	_, ok := other.(ErrInconsistentCommit)
	return ok
}

// ErrCommitOnOutputBranch represents an error where an attempt was made to start
// a commit on an output branch (a branch that is provenant on other branches).
// Users should not manually try to start a commit in an output branch, this
// should only be done internally in PFS.
type ErrCommitOnOutputBranch struct {
	Branch *pfs.Branch
}

// ErrSquashWithoutChildren represents an error when attempting to squash a
// commit that has no children.  Since squash works by removing a commit and
// leaving its data in any child commits, a squash would result in data loss in
// this situation, so it is not allowed.  To proceed anyway, use the
// DropCommitSet operation, which implies data loss.
type ErrSquashWithoutChildren struct {
	Commit *pfs.Commit
}

// ErrDropWithChildren represents an error when attempting to drop a commit that
// has children.  Because proper datum removal semantics have not been
// implemented in the middle of a commit chain, this operation is unsupported.
// However, a drop is still allowed for a commit with no children as there is no
// cleanup needed for child commits.
type ErrDropWithChildren struct {
	Commit *pfs.Commit
}

// ErrInvalidProvenanceStructure represents an error where more than one branch
// from a repo is reachable in a DAG. Such a DAG structure is unsupported.
type ErrInvalidProvenanceStructure struct {
	Branch *pfs.Branch
}

// ErrSquashWithSubvenance represents an error where a commit set cannot be deleted
// because it contains commits that are depended on by commits in another commit set.
// The dependent commit sets can be deleted by passing a force flag.
type ErrSquashWithSubvenance struct {
	CommitSet           *pfs.CommitSet
	SubvenantCommitSets []*pfs.CommitSet
}

const GetFileTARSuggestion = "Use GetFileTAR instead"

var (
	ErrMatchedNonFile       = fmt.Errorf("cannot get directory or non-regular file. %s", GetFileTARSuggestion)
	ErrMatchedMultipleFiles = fmt.Errorf("matched multiple files. %s", GetFileTARSuggestion)
)

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("file %v not found in repo %v at commit %v", e.File.Path, e.File.Commit.Repo, e.File.Commit.Id)
}

func (e ErrFileNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrRepoNotFound) Error() string {
	return fmt.Sprintf("repo %v not found", e.Repo)
}

func (e ErrRepoNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrRepoExists) Error() string {
	return fmt.Sprintf("repo %v already exists", e.Repo)
}

func (e ErrRepoExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

func (e ErrBranchNotFound) Error() string {
	return fmt.Sprintf("branch %q not found in repo %v", e.Branch.Name, e.Branch.Repo)
}

func (e ErrBranchNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrBranchExists) Error() string {
	return fmt.Sprintf("branch %q already exists in repo %v", e.Branch.Name, e.Branch.Repo)
}

func (e ErrBranchExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

func (e ErrProjectNotFound) Error() string {
	return fmt.Sprintf("project %q not found", e.Project.Name)
}

func (e ErrProjectNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrProjectExists) Error() string {
	return fmt.Sprintf("project %s already exists", e.Project.Name)
}

func (e ErrProjectExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

func (e ErrCommitNotFound) Error() string {
	return fmt.Sprintf("commit %v not found", e.Commit)
}

func (e ErrCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrCommitSetNotFound) Error() string {
	return fmt.Sprintf("no commits found for commitset %v", e.CommitSet.Id)
}

func (e ErrCommitSetNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrCommitExists) Error() string {
	return fmt.Sprintf("commit %v already exists", e.Commit)
}

func (e ErrCommitExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

func (e ErrCommitFinished) Error() string {
	return fmt.Sprintf("commit %v has already finished", e.Commit)
}

func (e ErrCommitError) Error() string {
	return fmt.Sprintf("commit %v finished with an error", e.Commit)
}

func (e ErrCommitDeleted) Error() string {
	return fmt.Sprintf("commit %v was deleted", e.Commit)
}

func (e ErrParentCommitNotFound) Error() string {
	return fmt.Sprintf("parent commit %v not found", e.Commit)
}

func (e ErrParentCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

func (e ErrOutputCommitNotFinished) Error() string {
	return fmt.Sprintf("output commit %v not finished", e.Commit)
}

func (e ErrCommitNotFinished) Error() string {
	return fmt.Sprintf("commit %v not finished", e.Commit)
}

func (err ErrCommitNotFinished) GRPCStatus() *status.Status {
	s, sErr := status.New(codes.Unavailable, fmt.Sprintf("commit %v not finished", err.Commit)).WithDetails(&epb.ResourceInfo{
		ResourceType: "pfs:commit",
		ResourceName: err.Commit.Id,
		Description:  "commit not finished",
	})
	if sErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		return status.New(codes.Internal, fmt.Sprintf("internal server error in %s:%d: %v", filename, line, sErr))
	}
	return s
}

func (e ErrBaseCommitNotFinished) Error() string {
	return fmt.Sprintf("base commit %v for commit %v unfinished", e.BaseCommit, e.Commit)
}

func (e ErrInconsistentCommit) Error() string {
	return fmt.Sprintf("inconsistent dependencies: cannot create commit from %s; repo (%s) already has a commit in this transaction", e.Commit, e.Commit.Repo.Name)
}

func (e ErrCommitOnOutputBranch) Error() string {
	return fmt.Sprintf("cannot start a commit on an output branch: %s", e.Branch)
}

func (e ErrSquashWithoutChildren) Error() string {
	return fmt.Sprintf("cannot squash a commit that has no children as that would cause data loss, use the drop operation instead: %s", e.Commit)
}

func (e ErrDropWithChildren) Error() string {
	return fmt.Sprintf("cannot drop a commit that has children: %s", e.Commit)
}

func (e ErrInvalidProvenanceStructure) Error() string {
	return fmt.Sprintf("multiple branches from the same repo, %q, cannot participate in a DAG", e.Branch.Repo.String())
}

func (e ErrSquashWithSubvenance) Error() string {
	return fmt.Sprintf("commit set %q cannot be dropped because it has subvenant commit sets: %v", e.CommitSet, e.SubvenantCommitSets)
}

var (
	commitNotFoundRe          = regexp.MustCompile("commit [^ ]+ not found")
	commitsetNotFoundRe       = regexp.MustCompile("no commits found for commitset")
	commitDeletedRe           = regexp.MustCompile("commit [^ ]+ was deleted")
	commitFinishedRe          = regexp.MustCompile("commit [^ ]+ has already finished")
	commitErrorRe             = regexp.MustCompile("commit [^ ]+ finished with an error")
	repoNotFoundRe            = regexp.MustCompile(`repos ([a-zA-Z0-9.\-_]{1,255}/)?[a-zA-Z0-9.\-_]{1,255} not found`)
	repoExistsRe              = regexp.MustCompile(`repo ?([a-zA-Z0-9.\-_]{1,255}/)?[a-zA-Z0-9.\-_]{1,255} already exists`)
	branchNotFoundRe          = regexp.MustCompile(`branch [^ ]+ not found in repo [^ ]*/[^ ]+`)
	fileNotFoundRe            = regexp.MustCompile(`file .+ not found`)
	outputCommitNotFinishedRe = regexp.MustCompile("output commit .+ not finished")
	commitNotFinishedRe       = regexp.MustCompile("commit .+ not finished")
	baseCommitNotFinishedRe   = regexp.MustCompile("base commit .+ unfinished")
	ambiguousCommitRe         = regexp.MustCompile("commit .+ is ambiguous")
	inconsistentCommitRe      = regexp.MustCompile("branch already has a commit in this transaction")
	commitOnOutputBranchRe    = regexp.MustCompile("cannot start a commit on an output branch")
	squashWithoutChildrenRe   = regexp.MustCompile("cannot squash a commit that has no children")
	dropWithChildrenRe        = regexp.MustCompile("cannot drop a commit that has children")
	invalidBranchStructureRe  = regexp.MustCompile("multiple branches from the same repo, .+, cannot participate in a DAG")
	squashWithSubvenanceRe    = regexp.MustCompile("commit set .+ cannot be dropped because it has subvenant commit sets: .+")
)

// IsCommitNotFoundErr returns true if 'err' has an error message that matches
// ErrCommitNotFound
func IsCommitNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return commitNotFoundRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}

// IsCommitSetNotFoundErr returns true if 'err' has an error message that matches
// ErrCommitSetNotFound
func IsCommitSetNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return commitsetNotFoundRe.MatchString(grpcutil.ScrubGRPC(err).Error())
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

// IsCommitError returns true of 'err' has an error message that matches
// ErrCommitError
func IsCommitErrorErr(err error) bool {
	if err == nil {
		return false
	}
	return commitErrorRe.MatchString(grpcutil.ScrubGRPC(err).Error())
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
	if fileNotFoundRe.MatchString(err.Error()) {
		return true
	}
	if status.Code(err) == codes.NotFound && strings.Contains(err.Error(), "commit") {
		return true
	}
	return false
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

func IsBaseCommitNotFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return baseCommitNotFinishedRe.MatchString(err.Error())
}

// IsAmbiguousCommitErr returns true if the err is due to attempting to resolve
// a commit without specifying a branch when it is required to uniquely identify
// the commit.
func IsAmbiguousCommitErr(err error) bool {
	if err == nil {
		return false
	}
	return ambiguousCommitRe.MatchString(err.Error())
}

// IsInconsistentCommitErr returns true if the err is due to an attempt to have
// multiple commits in a single branch within a transaction.
func IsInconsistentCommitErr(err error) bool {
	if err == nil {
		return false
	}
	return inconsistentCommitRe.MatchString(err.Error())
}

// IsCommitOnOutputBranchErr returns true if the err is due to an attempt to
// start a commit on an output branch.
func IsCommitOnOutputBranchErr(err error) bool {
	if err == nil {
		return false
	}
	return commitOnOutputBranchRe.MatchString(err.Error())
}

func IsSquashWithoutChildrenErr(err error) bool {
	if err == nil {
		return false
	}
	return squashWithoutChildrenRe.MatchString(err.Error())
}

func IsDropWithChildrenErr(err error) bool {
	if err == nil {
		return false
	}
	return dropWithChildrenRe.MatchString(err.Error())
}

func IsInvalidBranchStructureErr(err error) bool {
	if err == nil {
		return false
	}
	return invalidBranchStructureRe.MatchString(err.Error())
}

func IsSquashWithSuvenanceErr(err error) bool {
	if err == nil {
		return false
	}
	return squashWithSubvenanceRe.MatchString(err.Error())
}

func ValidateSQLDatabaseEgress(sql *pfs.SQLDatabaseEgress) error {
	if sql == nil {
		return nil
	}
	secret := sql.GetSecret()
	if secret == nil {
		return errors.Errorf("egress.sql_database.secret is required")
	}
	if secret.Name == "" || secret.Key == "" {
		return errors.Errorf("egress.sql_database.secret.name and egress.sql_database.secret.key are required")
	}
	return nil
}
