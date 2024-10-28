package server

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

func (a *apiServer) createRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, description string, update bool) error {
	if repo == nil {
		return errors.New("repo cannot be nil")
	}
	repo.EnsureProject()
	if repo.Type == "" {
		// default to user type
		repo.Type = pfs.UserRepoType
	}

	if err := ancestry.ValidateName(repo.Name); err != nil {
		return err
	}

	// Check that the repo’s project exists; if not, return an error.
	if _, err := pfsdb.GetProjectInfoByName(ctx, txnCtx.SqlTx, repo.Project.Name); err != nil {
		return errors.Wrapf(err, "cannot find project %q", repo.Project)
	}

	// Check whether the user has the permission to create a repo in this project.
	if err := a.env.Auth.CheckProjectIsAuthorizedInTransaction(ctx, txnCtx, repo.Project, auth.Permission_PROJECT_CREATE_REPO); err != nil {
		return errors.Wrap(err, "failed to create repo")
	}

	// Check if 'repo' already exists. If so, return that error.  Otherwise,
	// proceed with ACL creation (avoids awkward "access denied" error when
	// calling "createRepo" on a repo that already exists).
	existingRepoInfo, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type)
	if err != nil {
		if !pfsdb.IsErrRepoNotFound(err) {
			return errors.Wrapf(err, "error checking whether repo %q exists", repo)
		}
		// if this is a system repo, make sure the corresponding user repo already exists
		if repo.Type != pfs.UserRepoType {
			baseRepo := &pfs.Repo{Project: &pfs.Project{Name: repo.Project.GetName()}, Name: repo.GetName(), Type: pfs.UserRepoType}
			_, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, baseRepo.Project.Name, baseRepo.Name, baseRepo.Type)
			if err != nil {
				if !pfsdb.IsErrRepoNotFound(err) {
					return errors.Wrapf(err, "error checking whether user repo for %q exists", baseRepo)
				}
				return errors.Wrap(err, "cannot create a system repo without a corresponding 'user' repo")
			}
		}

		// New repo case
		if whoAmI, err := txnCtx.WhoAmI(); err != nil {
			if !errors.Is(err, auth.ErrNotActivated) {
				return errors.Wrap(err, "error authenticating (must log in to create a repo)")
			}
			// ignore ErrNotActivated because we simply don't need to create the role bindings.
		} else {
			// Create ACL for new repo. Make caller the sole owner. If this is a user repo,
			// and the ACL already exists with a different owner, this will fail.
			// For now, we expect system repos to share auth info with their corresponding
			// user repo, so the role binding should exist
			if err := a.env.Auth.CreateRoleBindingInTransaction(ctx, txnCtx, whoAmI.Username, []string{auth.RepoOwnerRole}, repo.AuthResource()); err != nil && (!col.IsErrExists(err) || repo.Type == pfs.UserRepoType) {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not create role binding for new repo %q", repo)
			}
		}
		_, err := pfsdb.UpsertRepo(ctx, txnCtx.SqlTx, &pfs.RepoInfo{Repo: repo, Description: description})
		return errors.Wrapf(err, "could not create repo %q", repo)
	}

	// Existing repo case--just update the repo description.
	if !update {
		return pfsserver.ErrRepoExists{
			Repo: repo,
		}
	}
	if existingRepoInfo.Description == description {
		// Don't overwrite the stored proto with an identical value. This
		// optimization is impactful because pps will frequently update the spec
		// repo to make sure it exists.
		return nil
	}

	// Check if the caller is authorized to modify this repo
	// Note, we don't do this before checking if the description changed because
	// there is client code that calls CreateRepo(R, update=true) as an
	// idempotent way to ensure that R exists. By permitting these calls when
	// they don't actually change anything, even if the caller doesn't have
	// WRITER access, we make the pattern more generally useful.
	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.Wrapf(err, "could not update description of %q", repo)
	}
	existingRepoInfo.Description = description
	_, err = pfsdb.UpsertRepo(ctx, txnCtx.SqlTx, existingRepoInfo)
	return errors.Wrapf(err, "could not update description of %q", repo)
}

func (a *apiServer) deleteRepos(ctx context.Context, projects []*pfs.Project, force bool) ([]*pfs.Repo, error) {
	var deleted []*pfs.Repo
	if err := a.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		deleted, err = a.deleteReposInTransaction(ctx, txnCtx, projects, force)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "delete repos")
	}
	return deleted, nil
}

func (a *apiServer) deleteReposInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, projects []*pfs.Project, force bool) ([]*pfs.Repo, error) {
	repos, err := a.listRepoInTransaction(ctx, txnCtx, false /* includeAuth */, "", projects, nil)
	if err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	return a.deleteReposHelper(ctx, txnCtx, repos, force)
}

func (a *apiServer) deleteReposHelper(ctx context.Context, txnCtx *txncontext.TransactionContext, ris []*pfs.RepoInfo, force bool) ([]*pfs.Repo, error) {
	if len(ris) == 0 {
		return nil, nil
	}
	// filter out repos that cannot be deleted
	var filter []*pfs.RepoInfo
	for _, ri := range ris {
		ok, err := a.canDeleteRepo(ctx, txnCtx, ri.Repo)
		if err != nil {
			return nil, err
		} else if ok {
			filter = append(filter, ri)
		}
	}
	ris = filter
	var deleted []*pfs.Repo
	var bis []*pfs.BranchInfo
	for _, ri := range ris {
		bs, err := a.listRepoBranches(ctx, txnCtx, ri)
		if err != nil {
			return nil, err
		}
		bis = append(bis, bs...)
	}
	if err := a.deleteBranches(ctx, txnCtx, bis, force); err != nil {
		return nil, err
	}
	for _, ri := range ris {
		if err := a.deleteRepoInfo(ctx, txnCtx, ri, force); err != nil {
			return nil, err
		}
		deleted = append(deleted, ri.Repo)
	}
	return deleted, nil
}

func (a *apiServer) deleteRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repoHandle *pfs.Repo, force bool) (bool, error) {
	if _, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repoHandle.Project.Name, repoHandle.Name, repoHandle.Type); err != nil {
		if !pfsdb.IsErrRepoNotFound(err) {
			return false, errors.Wrapf(err, "error checking whether %q exists", repoHandle)
		}
		return false, nil
	}
	if ok, err := a.canDeleteRepo(ctx, txnCtx, repoHandle); err != nil {
		return false, errors.Wrapf(err, "error checking whether repo %q can be deleted", repoHandle.String())
	} else if !ok {
		return false, nil
	}
	related, err := a.relatedRepos(ctx, txnCtx, repoHandle)
	if err != nil {
		return false, err
	}
	var bis []*pfs.BranchInfo
	for _, repo := range related {
		bs, err := a.listRepoBranches(ctx, txnCtx, repo.RepoInfo)
		if err != nil {
			return false, err
		}
		bis = append(bis, bs...)
	}
	// this is where the deletion is happening.
	if err := a.deleteBranches(ctx, txnCtx, bis, force); err != nil {
		return false, err
	}
	for _, repo := range related {
		if err := a.deleteRepoInfo(ctx, txnCtx, repo.RepoInfo, force); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *apiServer) relatedRepos(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]pfsdb.Repo, error) {
	if repo.Type != pfs.UserRepoType {
		ri, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type)
		if err != nil {
			return nil, err
		}
		return []pfsdb.Repo{{RepoInfo: ri}}, nil
	}
	filter := &pfsdb.RepoFilter{
		RepoTemplate: &pfs.Repo{Name: repo.Name, Project: repo.Project},
	}
	related, err := pfsdb.ListRepo(ctx, txnCtx.SqlTx, filter, nil)
	if err != nil {
		return nil, errors.Wrap(err, "list repo by name")
	}
	return related, nil
}

func (a *apiServer) canDeleteRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo) (bool, error) {
	userRepo := proto.Clone(repo).(*pfs.Repo)
	userRepo.Type = pfs.UserRepoType
	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, userRepo, auth.Permission_REPO_DELETE); err != nil {
		if auth.IsErrNotAuthorized(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "check repo %q is authorized for deletion", userRepo.String())
	}
	if _, err := a.env.GetPipelineInspector().InspectPipelineInTransaction(ctx, txnCtx, pps.RepoPipeline(repo)); err == nil {
		return false, errors.Errorf("cannot delete a repo associated with a pipeline - delete the pipeline instead")
	} else if !errutil.IsNotFoundError(err) {
		return false, errors.Wrapf(err, "inspect pipeline %q", pps.RepoPipeline(repo).String())
	}
	return true, nil
}

// before this method is called, a caller should make sure this repo can be deleted with a.canDeleteRepo() and that
// all of the repo's branches are deleted using a.deleteBranches()
func (a *apiServer) deleteRepoInfo(ctx context.Context, txnCtx *txncontext.TransactionContext, ri *pfs.RepoInfo, force bool) error {
	var nonCtxCommitInfos []*pfs.CommitInfo
	if err := pfsdb.ForEachCommit(ctx, a.env.DB, &pfs.Commit{Repo: ri.Repo}, func(commit pfsdb.Commit) error {
		nonCtxCommitInfos = append(nonCtxCommitInfos, commit.CommitInfo)
		return nil
	}); err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	commits, err := pfsdb.ListCommitTxByFilter(ctx, txnCtx.SqlTx, &pfs.Commit{Repo: ri.Repo})
	if err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	// and then delete their file sets.
	for _, commit := range commits {
		if err := a.commitStore.DropFileSetsTx(txnCtx.SqlTx, commit); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Despite the fact that we already deleted each branch with
	// deleteBranch, we also do branches.DeleteAll(), this insulates us
	// against certain corruption situations where the RepoInfo doesn't
	// exist in postgres but branches do.
	err = pfsdb.ForEachBranch(ctx, txnCtx.SqlTx, &pfs.Branch{Repo: ri.Repo}, func(branch pfsdb.Branch) error {
		return pfsdb.DeleteBranch(ctx, txnCtx.SqlTx, &branch, force)
	}, pfsdb.OrderByBranchColumn{Column: pfsdb.BranchColumnID, Order: pfsdb.SortOrderAsc})
	if err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	// Similarly with commits
	for _, commitInfo := range commits {
		if err := pfsdb.DeleteCommit(ctx, txnCtx.SqlTx, commitInfo.Commit); err != nil {
			return errors.Wrap(err, "delete repo info")
		}
	}
	if err := pfsdb.DeleteRepo(ctx, txnCtx.SqlTx, ri.Repo.Project.Name, ri.Repo.Name, ri.Repo.Type); err != nil && !pfsdb.IsErrRepoNotFound(err) {
		return errors.Wrapf(err, "repos.Delete")
	}
	// since system repos share a role binding, only delete it if this is the user repo, in which case the other repos will be deleted anyway
	if ri.Repo.Type == pfs.UserRepoType {
		if err := a.env.Auth.DeleteRoleBindingInTransaction(ctx, txnCtx, ri.Repo.AuthResource()); err != nil && !auth.IsErrNotActivated(err) {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

func (a *apiServer) inspectRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
	// Validate arguments
	if repo == nil {
		return nil, errors.New("repo cannot be nil")
	}
	repoInfo, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type)
	if err != nil {
		if pfsdb.IsErrRepoNotFound(err) {
			return nil, pfsserver.ErrRepoNotFound{Repo: repo}
		}
		return nil, errors.EnsureStack(err)
	}
	if includeAuth {
		resp, err := a.env.Auth.GetPermissionsInTransaction(ctx, txnCtx, &auth.GetPermissionsRequest{
			Resource: repo.AuthResource(),
		})
		if err != nil {
			if auth.IsErrNotActivated(err) {
				return repoInfo, nil
			}
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error getting access level for %q", repo)
		}
		repoInfo.AuthInfo = &pfs.AuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
	}
	return repoInfo, nil
}

func (a *apiServer) listRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, includeAuth bool, repoType string, projects []*pfs.Project, page *pfs.RepoPage) ([]*pfs.RepoInfo, error) {
	filter := &pfsdb.RepoFilter{
		RepoTemplate: &pfs.Repo{},
		Projects:     projects,
	}
	if repoType != "" { // blank type means return all, otherwise return a specific type
		filter.RepoTemplate.Type = repoType
	}
	var authActive bool
	if includeAuth {
		if _, err := txnCtx.WhoAmI(); err == nil {
			authActive = true
		} else if !errors.Is(err, auth.ErrNotActivated) {
			return nil, errors.Wrap(err, "check if auth is active")
		}
	}
	// Cache the project level access because it applies to every repo within the same project.
	checkProjectAccess := miscutil.CacheFunc(func(project string) error {
		return a.env.Auth.CheckProjectIsAuthorizedInTransaction(ctx, txnCtx, &pfs.Project{Name: project}, auth.Permission_PROJECT_LIST_REPO)
	}, 100)
	var repos []*pfs.RepoInfo
	if err := pfsdb.ForEachRepo(ctx, txnCtx.SqlTx, filter, page, func(repo pfsdb.Repo) error {
		if authActive {
			hasAccess, err := a.hasProjectAccess(ctx, txnCtx, repo.RepoInfo, checkProjectAccess)
			if err != nil {
				return errors.Wrap(err, "checking project access")
			}
			if !hasAccess {
				return nil
			}
			permissions, roles, err := a.getPermissionsInTransaction(ctx, txnCtx, repo.RepoInfo.Repo)
			if err != nil {
				return errors.Wrapf(err, "error getting access level for %q", repo.RepoInfo.Repo)
			}
			repo.RepoInfo.AuthInfo = &pfs.AuthInfo{Permissions: permissions, Roles: roles}
		}
		size, err := a.repoSize(ctx, txnCtx, repo.RepoInfo)
		if err != nil {
			return errors.Wrapf(err, "getting repo size for %q", repo.RepoInfo.Repo)
		}
		repo.RepoInfo.SizeBytesUpperBound = size
		repos = append(repos, repo.RepoInfo)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "for each repo")
	}
	return repos, nil
}

// NOTE: repoSize() is only calculated based on the "master" branch
func (a *apiServer) repoSize(ctx context.Context, txnCtx *txncontext.TransactionContext, repoInfo *pfs.RepoInfo) (int64, error) {
	for _, branch := range repoInfo.Branches {
		if branch.Name == "master" {
			branchInfo, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branch)
			if err != nil {
				return 0, errors.Wrap(err, "repo size")
			}
			commit := branchInfo.Head
			for commit != nil {
				c, err := a.pickCommitTx(ctx, txnCtx.SqlTx, commit)
				if err != nil {
					return 0, err
				}
				if c.Details != nil {
					return c.Details.SizeBytes, nil
				}
				commit = c.ParentCommit
			}
		}
	}
	return 0, nil
}
