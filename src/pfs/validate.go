package pfs

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type LiveValidator interface {
	ValidateRepoExists(*Repo) error
	ValidateProjectExists(*Project) error
}

type LiveValidationError struct {
	Path string
	Err  error
}

var _ error = new(LiveValidationError)

func (err *LiveValidationError) Error() string {
	return "field " + err.Path + ": " + err.Err.Error()
}

func (err *LiveValidationError) Unwrap() error {
	return err.Err
}

func validationError(path string, err error) error {
	if err == nil {
		return nil
	}
	return &LiveValidationError{Path: path, Err: err}
}

func (x *CreateRepoRequest) ValidatePFS(v LiveValidator) error {
	return validationError("repo.project", v.ValidateProjectExists(x.GetRepo().GetProject()))
}

func (x *InspectRepoRequest) ValidatePFS(v LiveValidator) error {
	return validationError("repo.project", v.ValidateProjectExists(x.GetRepo().GetProject()))
}

func (x *ListRepoRequest) ValidatePFS(v LiveValidator) error {
	var errs error
	for i, p := range x.GetProjects() {
		errors.JoinInto(&errs, validationError(fmt.Sprintf("projects[%d]", i), v.ValidateProjectExists(p)))
	}
	return errs
}

func (x *DeleteRepoRequest) ValidatePFS(v LiveValidator) error {
	return validationError("repo.project", v.ValidateProjectExists(x.GetRepo().GetProject()))
}

func (x *DeleteReposRequest) ValidatePFS(v LiveValidator) error {
	var errs error
	for i, p := range x.GetProjects() {
		errors.JoinInto(&errs, validationError(fmt.Sprintf("projects[%d]", i), v.ValidateProjectExists(p)))
	}
	return errs
}

func (x *StartCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("branch.repo", v.ValidateRepoExists(x.GetBranch().GetRepo()))
}

func (x *FinishCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *InspectCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *ListCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("repo", v.ValidateRepoExists(x.GetRepo()))
}

func (x *InspectCommitSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ListCommitSetRequest) ValidatePFS(v LiveValidator) error {
	if p := x.Project; p != nil {
		return validationError("project", v.ValidateProjectExists(p))
	}
	return nil
}

func (x *SquashCommitSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *DropCommitSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *SubscribeCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("repo", v.ValidateRepoExists(x.GetRepo()))
}

func (x *ClearCommitRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *CreateBranchRequest) ValidatePFS(v LiveValidator) error {
	var errs error
	errors.JoinInto(&errs, validationError("branch.repo", v.ValidateRepoExists(x.GetBranch().GetRepo())))
	for i, b := range x.GetProvenance() {
		errors.JoinInto(&errs, validationError(fmt.Sprintf("provenance[%d].repo", i), v.ValidateRepoExists(b.GetRepo())))
	}
	return errs
}

func (x *FindCommitsRequest) ValidatePFS(v LiveValidator) error {
	return validationError("start.repo", v.ValidateRepoExists(x.GetStart().GetRepo()))
}

func (x *InspectBranchRequest) ValidatePFS(v LiveValidator) error {
	return validationError("branch.repo", v.ValidateRepoExists(x.GetBranch().GetRepo()))
}

func (x *ListBranchRequest) ValidatePFS(v LiveValidator) error {
	if x.Repo != nil {
		return validationError("repo", v.ValidateRepoExists(x.GetRepo()))
	}
	return nil
}

func (x *DeleteBranchRequest) ValidatePFS(v LiveValidator) error {
	return validationError("branch.repo.project", v.ValidateProjectExists(x.GetBranch().GetRepo().GetProject()))
}

func (x *CreateProjectRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *InspectProjectRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ListProjectRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *DeleteProjectRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *GetFileRequest) ValidatePFS(v LiveValidator) error {
	return validationError("file.commit.repo", v.ValidateRepoExists(x.GetFile().GetCommit().GetRepo()))
}

func (x *InspectFileRequest) ValidatePFS(v LiveValidator) error {
	return validationError("file.commit.repo", v.ValidateRepoExists(x.GetFile().GetCommit().GetRepo()))
}

func (x *ListFileRequest) ValidatePFS(v LiveValidator) error {
	return validationError("file.commit.repo", v.ValidateRepoExists(x.GetFile().GetCommit().GetRepo()))
}

func (x *WalkFileRequest) ValidatePFS(v LiveValidator) error {
	return validationError("file.commit.repo", v.ValidateRepoExists(x.GetFile().GetCommit().GetRepo()))
}

func (x *GlobFileRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *DiffFileRequest) ValidatePFS(v LiveValidator) error {
	var errs error
	errors.JoinInto(&errs, validationError("new_file.commit.repo", v.ValidateRepoExists(x.GetNewFile().GetCommit().GetRepo())))
	errors.JoinInto(&errs, validationError("old_file.commit.repo", v.ValidateRepoExists(x.GetOldFile().GetCommit().GetRepo())))
	return errs
}

func (x *FsckRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *CreateFileSetResponse) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *GetFileSetRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *AddFileSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *RenewFileSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ComposeFileSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ShardFileSetRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *CheckStorageRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *PutCacheRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *GetCacheRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ClearCacheRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *ActivateAuthRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *RunLoadTestRequest) ValidatePFS(v LiveValidator) error {
	return nil
}

func (x *EgressRequest) ValidatePFS(v LiveValidator) error {
	return validationError("commit.repo", v.ValidateRepoExists(x.GetCommit().GetRepo()))
}

func (x *ModifyFileRequest) ValidatePFS(v LiveValidator) error {
	switch x := x.GetBody().(type) {
	case *ModifyFileRequest_SetCommit:
		return validationError("body.set_commit.repo", v.ValidateRepoExists(x.SetCommit.GetRepo()))
	}
	return nil
}
