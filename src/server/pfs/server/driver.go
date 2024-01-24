package server

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

const (
	StorageTaskNamespace = "storage"
	fileSetsRepo         = client.FileSetsRepoName
	defaultTTL           = client.DefaultTTL
	maxTTL               = 30 * time.Minute
)

// IsPermissionError returns true if a given error is a permission error.
func IsPermissionError(err error) bool {
	return strings.Contains(err.Error(), "has already finished")
}

// CommitEvent is an event that contains a CommitInfo or an error
type CommitEvent struct {
	Err   error
	Value *pfs.CommitInfo
}

// CommitStream is a stream of CommitInfos
type CommitStream interface {
	Stream() <-chan CommitEvent
	Close()
}

type driver struct {
	env Env

	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	txnEnv     *txnenv.TransactionEnv
	prefix     string

	// collections
	commits  col.PostgresCollection
	branches col.PostgresCollection

	storage     *storage.Server
	commitStore commitStore

	cache *fileset.Cache
}

func newDriver(env Env) (*driver, error) {
	// test object storage.
	if err := func() error {
		ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
		defer cf()
		return obj.TestStorage(ctx, env.Bucket, env.ObjectClient)
	}(); err != nil {
		return nil, err
	}
	commits := pfsdb.Commits(env.DB, env.Listener)
	branches := pfsdb.Branches(env.DB, env.Listener)

	// Setup driver struct.
	d := &driver{
		env:        env,
		etcdClient: env.EtcdClient,
		txnEnv:     env.TxnEnv,
		prefix:     env.EtcdPrefix,
		commits:    commits,
		branches:   branches,
	}
	storageEnv := storage.Env{DB: env.DB}
	if env.Bucket != nil {
		storageEnv.Bucket = env.Bucket
	} else {
		storageEnv.ObjectStore = env.ObjectClient
	}
	storageSrv, err := storage.New(storageEnv, env.StorageConfig)
	if err != nil {
		return nil, err
	}
	d.storage = storageSrv
	d.commitStore = newPostgresCommitStore(env.DB, storageSrv.Tracker, storageSrv.Filesets)
	// TODO: Make the cache max size configurable.
	d.cache = fileset.NewCache(env.DB, storageSrv.Tracker, 10000)
	return d, nil
}

func (d *driver) createRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, description string, update bool) error {
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
	if _, err := pfsdb.GetProjectByName(ctx, txnCtx.SqlTx, repo.Project.Name); err != nil {
		return errors.Wrapf(err, "cannot find project %q", repo.Project)
	}

	// Check whether the user has the permission to create a repo in this project.
	if err := d.env.Auth.CheckProjectIsAuthorizedInTransaction(txnCtx, repo.Project, auth.Permission_PROJECT_CREATE_REPO); err != nil {
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
			if err := d.env.Auth.CreateRoleBindingInTransaction(txnCtx, whoAmI.Username, []string{auth.RepoOwnerRole}, repo.AuthResource()); err != nil && (!col.IsErrExists(err) || repo.Type == pfs.UserRepoType) {
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
	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.Wrapf(err, "could not update description of %q", repo)
	}
	existingRepoInfo.Description = description
	_, err = pfsdb.UpsertRepo(ctx, txnCtx.SqlTx, existingRepoInfo)
	return errors.Wrapf(err, "could not update description of %q", repo)
}

func (d *driver) inspectRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
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
		resp, err := d.env.Auth.GetPermissionsInTransaction(txnCtx, &auth.GetPermissionsRequest{
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

func (d *driver) getPermissionsInTransaction(txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]auth.Permission, []string, error) {
	resp, err := d.env.Auth.GetPermissionsInTransaction(txnCtx, &auth.GetPermissionsRequest{Resource: repo.AuthResource()})
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}

	return resp.Permissions, resp.Roles, nil
}

func (d *driver) hasProjectAccess(
	txnCtx *txncontext.TransactionContext,
	repoInfo *pfs.RepoInfo,
	checkProjectAccess func(string) error) (bool, error) {
	if err := checkProjectAccess(repoInfo.Repo.Project.GetName()); err != nil {
		if !errors.As(err, &auth.ErrNotAuthorized{}) {
			return false, err
		}
		// Allow access if user has the right permissions at the individual Repo-level.
		if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, repoInfo.Repo, auth.Permission_REPO_READ); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return false, errors.Wrapf(err, "could not check user is authorized to list repo, problem with repo %s", repoInfo.Repo)
			}
			// User does not have permissions, so we should skip this repo.
			return false, nil
		}
	}
	return true, nil
}

func (d *driver) listRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, includeAuth bool, repoType string, projects []*pfs.Project) ([]*pfs.RepoInfo, error) {
	projectNames := make(map[string]bool, 0)
	for _, project := range projects {
		projectNames[project.GetName()] = true
	}
	filter := &pfs.Repo{}
	if repoType != "" { // blank type means return all, otherwise return a specific type
		filter.Type = repoType
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
		return d.env.Auth.CheckProjectIsAuthorizedInTransaction(txnCtx, &pfs.Project{Name: project}, auth.Permission_PROJECT_LIST_REPO)
	}, 100)
	var repos []*pfs.RepoInfo
	if err := pfsdb.ForEachRepo(ctx, txnCtx.SqlTx, filter, func(repoWithID pfsdb.RepoInfoWithID) error {
		if _, ok := projectNames[repoWithID.RepoInfo.Repo.Project.GetName()]; !ok && len(projectNames) > 0 {
			return nil // project doesn't match filter.
		}
		if authActive {
			hasAccess, err := d.hasProjectAccess(txnCtx, repoWithID.RepoInfo, checkProjectAccess)
			if err != nil {
				return errors.Wrap(err, "checking project access")
			}
			if !hasAccess {
				return nil
			}
			permissions, roles, err := d.getPermissionsInTransaction(txnCtx, repoWithID.RepoInfo.Repo)
			if err != nil {
				return errors.Wrapf(err, "error getting access level for %q", repoWithID.RepoInfo.Repo)
			}
			repoWithID.RepoInfo.AuthInfo = &pfs.AuthInfo{Permissions: permissions, Roles: roles}
		}
		size, err := d.repoSize(ctx, txnCtx, repoWithID.RepoInfo)
		if err != nil {
			return errors.Wrapf(err, "getting repo size for %q", repoWithID.RepoInfo.Repo)
		}
		repoWithID.RepoInfo.SizeBytesUpperBound = size
		repos = append(repos, repoWithID.RepoInfo)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "for each repo")
	}
	return repos, nil
}

func (d *driver) deleteRepos(ctx context.Context, projects []*pfs.Project, force bool) ([]*pfs.Repo, error) {
	var deleted []*pfs.Repo
	if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		deleted, err = d.deleteReposInTransaction(ctx, txnCtx, projects, force)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "delete repos")
	}
	return deleted, nil
}

func (d *driver) deleteReposInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, projects []*pfs.Project, force bool) ([]*pfs.Repo, error) {
	repos, err := d.listRepoInTransaction(ctx, txnCtx, false /* includeAuth */, "", projects)
	if err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	return d.deleteReposHelper(ctx, txnCtx, repos, force)
}

func (d *driver) deleteReposHelper(ctx context.Context, txnCtx *txncontext.TransactionContext, ris []*pfs.RepoInfo, force bool) ([]*pfs.Repo, error) {
	if len(ris) == 0 {
		return nil, nil
	}
	// filter out repos that cannot be deleted
	var filter []*pfs.RepoInfo
	for _, ri := range ris {
		ok, err := d.canDeleteRepo(ctx, txnCtx, ri.Repo)
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
		bs, err := d.listRepoBranches(ctx, txnCtx, ri)
		if err != nil {
			return nil, err
		}
		bis = append(bis, bs...)
	}
	if err := d.deleteBranches(ctx, txnCtx, bis, force); err != nil {
		return nil, err
	}
	for _, ri := range ris {
		if err := d.deleteRepoInfo(ctx, txnCtx, ri, force); err != nil {
			return nil, err
		}
		deleted = append(deleted, ri.Repo)
	}
	return deleted, nil
}

func (d *driver) deleteRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, force bool) (bool, error) {
	if _, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type); err != nil {
		if !pfsdb.IsErrRepoNotFound(err) {
			return false, errors.Wrapf(err, "error checking whether %q exists", repo)
		}
		return false, nil
	}
	if ok, err := d.canDeleteRepo(ctx, txnCtx, repo); err != nil {
		return false, errors.Wrapf(err, "error checking whether repo %q can be deleted", repo.String())
	} else if !ok {
		return false, nil
	}
	related, err := d.relatedRepos(ctx, txnCtx, repo)
	if err != nil {
		return false, err
	}
	var bis []*pfs.BranchInfo
	for _, repoInfoWithID := range related {
		bs, err := d.listRepoBranches(ctx, txnCtx, repoInfoWithID.RepoInfo)
		if err != nil {
			return false, err
		}
		bis = append(bis, bs...)
	}
	// this is where the deletion is happening.
	if err := d.deleteBranches(ctx, txnCtx, bis, force); err != nil {
		return false, err
	}
	for _, repoInfoWithID := range related {
		if err := d.deleteRepoInfo(ctx, txnCtx, repoInfoWithID.RepoInfo, force); err != nil {
			return false, err
		}
	}
	return true, nil
}

// delete branches from most provenance to least, that way if one
// branch is provenant on another (which is likely the case when
// multiple repos are provided) we delete them in the right order.
func (d *driver) deleteBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branchInfos []*pfs.BranchInfo, force bool) error {
	sort.Slice(branchInfos, func(i, j int) bool { return len(branchInfos[i].Provenance) > len(branchInfos[j].Provenance) })
	for _, bi := range branchInfos {
		if err := d.deleteBranch(ctx, txnCtx, bi.Branch, force); err != nil {
			return errors.Wrapf(err, "delete branch %s", bi.Branch)
		}
	}
	return nil
}

func (d *driver) listRepoBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.RepoInfo) ([]*pfs.BranchInfo, error) {
	var bis []*pfs.BranchInfo
	for _, branch := range repo.Branches {
		bi, err := d.inspectBranchInTransaction(ctx, txnCtx, branch)
		if err != nil {
			return nil, errors.Wrapf(err, "error inspecting branch %s", branch)
		}
		bis = append(bis, bi)
	}
	return bis, nil
}

// before this method is called, a caller should make sure this repo can be deleted with d.canDeleteRepo() and that
// all of the repo's branches are deleted using d.deleteBranches()
func (d *driver) deleteRepoInfo(ctx context.Context, txnCtx *txncontext.TransactionContext, ri *pfs.RepoInfo, force bool) error {
	var nonCtxCommitInfos []*pfs.CommitInfo
	if err := pfsdb.ForEachCommit(ctx, d.env.DB, &pfs.Commit{Repo: ri.Repo}, func(commitWithID pfsdb.CommitWithID) error {
		nonCtxCommitInfos = append(nonCtxCommitInfos, commitWithID.CommitInfo)
		return nil
	}); err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	commitInfos, err := pfsdb.ListCommitTxByFilter(ctx, txnCtx.SqlTx, &pfs.Commit{Repo: ri.Repo})
	if err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	// and then delete their file sets.
	for _, ci := range commitInfos {
		if err := d.commitStore.DropFileSetsTx(txnCtx.SqlTx, ci.Commit); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Despite the fact that we already deleted each branch with
	// deleteBranch, we also do branches.DeleteAll(), this insulates us
	// against certain corruption situations where the RepoInfo doesn't
	// exist in postgres but branches do.
	err = pfsdb.ForEachBranch(ctx, txnCtx.SqlTx, &pfs.Branch{Repo: ri.Repo}, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
		return pfsdb.DeleteBranch(ctx, txnCtx.SqlTx, &branchInfoWithID, force)
	}, pfsdb.OrderByBranchColumn{Column: pfsdb.BranchColumnID, Order: pfsdb.SortOrderAsc})
	if err != nil {
		return errors.Wrap(err, "delete repo info")
	}
	// Similarly with commits
	for _, commitInfo := range commitInfos {
		if err := pfsdb.DeleteCommit(ctx, txnCtx.SqlTx, commitInfo.Commit); err != nil {
			return errors.Wrap(err, "delete repo info")
		}
	}
	if err := pfsdb.DeleteRepo(ctx, txnCtx.SqlTx, ri.Repo.Project.Name, ri.Repo.Name, ri.Repo.Type); err != nil && !pfsdb.IsErrRepoNotFound(err) {
		return errors.Wrapf(err, "repos.Delete")
	}
	// since system repos share a role binding, only delete it if this is the user repo, in which case the other repos will be deleted anyway
	if ri.Repo.Type == pfs.UserRepoType {
		if err := d.env.Auth.DeleteRoleBindingInTransaction(txnCtx, ri.Repo.AuthResource()); err != nil && !auth.IsErrNotActivated(err) {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

func (d *driver) relatedRepos(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]pfsdb.RepoInfoWithID, error) {
	if repo.Type != pfs.UserRepoType {
		ri, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type)
		if err != nil {
			return nil, err
		}
		return []pfsdb.RepoInfoWithID{{RepoInfo: ri}}, nil
	}
	filter := &pfs.Repo{
		Name:    repo.Name,
		Project: repo.Project,
	}
	related, err := pfsdb.ListRepo(ctx, txnCtx.SqlTx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "list repo by name")
	}
	if err != nil && !pfsdb.IsErrRepoNotFound(err) { // TODO(acohen4): !RepoNotFound may be unnecessary
		return nil, errors.Wrapf(err, "error finding dependent repos for %q", repo.Name)
	}
	return related, nil
}

func (d *driver) canDeleteRepo(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo) (bool, error) {
	userRepo := proto.Clone(repo).(*pfs.Repo)
	userRepo.Type = pfs.UserRepoType
	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, userRepo, auth.Permission_REPO_DELETE); err != nil {
		if auth.IsErrNotAuthorized(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "check repo %q is authorized for deletion", userRepo.String())
	}
	if _, err := d.env.GetPipelineInspector().InspectPipelineInTransaction(ctx, txnCtx, pps.RepoPipeline(repo)); err == nil {
		return false, errors.Errorf("cannot delete a repo associated with a pipeline - delete the pipeline instead")
	} else if err != nil && !errutil.IsNotFoundError(err) {
		return false, errors.Wrapf(err, "inspect pipeline %q", pps.RepoPipeline(repo).String())
	}
	return true, nil
}

func (d *driver) createProject(ctx context.Context, req *pfs.CreateProjectRequest) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return d.createProjectInTransaction(ctx, txnCtx, req)
	})
}

func (d *driver) createProjectInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, req *pfs.CreateProjectRequest) error {
	if err := req.Project.ValidateName(); err != nil {
		return errors.Wrapf(err, "invalid project name")
	}
	if req.Update {
		return errors.Wrapf(pfsdb.UpsertProject(ctx, txnCtx.SqlTx,
			&pfs.ProjectInfo{Project: req.Project, Description: req.Description}),
			"failed to update project %s", req.Project.Name)
	}
	// If auth is active, make caller the owner of this new project.
	if whoAmI, err := txnCtx.WhoAmI(); err == nil {
		if err := d.env.Auth.CreateRoleBindingInTransaction(
			txnCtx,
			whoAmI.Username,
			[]string{auth.ProjectOwnerRole},
			&auth.Resource{Type: auth.ResourceType_PROJECT, Name: req.Project.GetName()},
		); err != nil && !errors.Is(err, col.ErrExists{}) {
			return errors.Wrapf(err, "could not create role binding for new project %s", req.Project.GetName())
		}
	} else if !errors.Is(err, auth.ErrNotActivated) {
		return errors.Wrap(err, "could not get caller's username")
	}
	if err := pfsdb.CreateProject(ctx, txnCtx.SqlTx, &pfs.ProjectInfo{
		Project:     req.Project,
		Description: req.Description,
		CreatedAt:   timestamppb.Now(),
	}); err != nil {
		if errors.As(err, &pfsdb.ProjectAlreadyExistsError{}) {
			return errors.Join(err, pfsserver.ErrProjectExists{Project: req.Project})
		}
		return errors.Wrap(err, "could not create project")
	}
	return nil
}

func (d *driver) inspectProject(ctx context.Context, project *pfs.Project) (*pfs.ProjectInfo, error) {
	var pi *pfs.ProjectInfo
	if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		pi, err = pfsdb.GetProjectByName(ctx, tx, pfsdb.ProjectKey(project))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "error getting project %q", project)
	}
	resp, err := d.env.Auth.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: project.AuthResource()})
	if err != nil {
		if errors.Is(err, auth.ErrNotActivated) {
			return pi, nil
		}
		return nil, errors.Wrapf(err, "error getting permissions for project %q", project)
	}
	pi.AuthInfo = &pfs.AuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
	return pi, nil
}

func (d *driver) findCommits(ctx context.Context, request *pfs.FindCommitsRequest, cb func(response *pfs.FindCommitsResponse) error) error {
	commit := request.Start
	foundCommits, commitsSearched := uint32(0), uint32(0)
	searchDone := false
	var found *pfs.Commit
	makeResp := func(found *pfs.Commit, commitsSearched uint32, lastSearchedCommit *pfs.Commit) *pfs.FindCommitsResponse {
		resp := &pfs.FindCommitsResponse{}
		if found != nil {
			resp.Result = &pfs.FindCommitsResponse_FoundCommit{FoundCommit: found}
		} else {
			resp.Result = &pfs.FindCommitsResponse_LastSearchedCommit{LastSearchedCommit: commit}
		}
		resp.CommitsSearched = commitsSearched
		return resp
	}
	for {
		if searchDone {
			return errors.EnsureStack(cb(makeResp(nil, commitsSearched, commit)))
		}
		logFields := []zap.Field{
			zap.String("commit", commit.Id),
			zap.String("repo", commit.Repo.String()),
			zap.String("target", request.FilePath),
		}
		if commit.Branch != nil {
			logFields = append(logFields, zap.String("branch", commit.Branch.String()))
		}
		if err := log.LogStep(ctx, "searchingCommit", func(ctx context.Context) error {
			inspectCommitResp, err := d.resolveCommitWithAuth(ctx, commit)
			if err != nil {
				return err
			}
			if inspectCommitResp.Finished == nil {
				return pfsserver.ErrCommitNotFinished{Commit: commit}
			}
			commit = inspectCommitResp.Commit
			inCommit, err := d.isPathModifiedInCommit(ctx, commit, request.FilePath)
			if err != nil {
				return err
			}
			if inCommit {
				log.Info(ctx, "found target", zap.String("commit", commit.Id), zap.String("repo", commit.Repo.String()), zap.String("target", request.FilePath))
				found = commit
				foundCommits++
				if err := cb(makeResp(found, commitsSearched, nil)); err != nil {
					return err
				}
			}
			commitsSearched++
			if (foundCommits == request.Limit && request.Limit != 0) || inspectCommitResp.ParentCommit == nil {
				searchDone = true
				return nil
			}
			commit = inspectCommitResp.ParentCommit
			return nil
		}, logFields...); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return errors.EnsureStack(cb(makeResp(nil, commitsSearched, commit)))
			}
			return errors.EnsureStack(err)
		}
	}
}

func (d *driver) isPathModifiedInCommit(ctx context.Context, commit *pfs.Commit, filePath string) (bool, error) {
	diffID, err := d.getCompactedDiffFileSet(ctx, commit)
	if err != nil {
		return false, err
	}
	diffFileSet, err := d.storage.Filesets.Open(ctx, []fileset.ID{*diffID})
	if err != nil {
		return false, err
	}
	found := false
	pathOption := []index.Option{
		index.WithRange(&index.PathRange{
			Lower: filePath,
		}),
	}
	if err = diffFileSet.Iterate(ctx, func(file fileset.File) error {
		if file.Index().Path == filePath {
			found = true
			return errutil.ErrBreak
		}
		return nil
	}, pathOption...); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return false, err
	}
	// we don't care about the file operation, so if a file was already found, skip iterating over deletive set.
	if found {
		return found, nil
	}
	if err = diffFileSet.IterateDeletes(ctx, func(file fileset.File) error {
		if file.Index().Path == filePath {
			found = true
			return errutil.ErrBreak
		}
		return nil
	}, pathOption...); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return false, err
	}
	return found, nil
}

func (d *driver) getCompactedDiffFileSet(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	diff, err := d.commitStore.GetDiffFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	isCompacted, err := d.storage.Filesets.IsCompacted(ctx, *diff)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if isCompacted {
		return diff, nil
	}
	if err := d.storage.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		compactor := newCompactor(d.storage.Filesets, d.env.StorageConfig.StorageCompactionMaxFanIn)
		taskDoer := d.env.TaskService.NewDoer(StorageTaskNamespace, commit.Id, nil)
		diff, err = d.compactDiffFileSet(ctx, compactor, taskDoer, renewer, commit)
		return err
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return diff, nil
}

// The ProjectInfo provided to the closure is repurposed on each invocation, so it's the client's responsibility to clone the ProjectInfo if desired
func (d *driver) listProject(ctx context.Context, cb func(*pfs.ProjectInfo) error) error {
	authIsActive := true
	return errors.Wrap(dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		return d.txnEnv.WithWriteContext(ctx, func(txnCxt *txncontext.TransactionContext) error {
			return d.listProjectInTransaction(ctx, txnCxt, func(proj *pfs.ProjectInfo) error {
				if authIsActive {
					resp, err := d.env.Auth.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: proj.GetProject().AuthResource()})
					if err != nil {
						if errors.Is(err, auth.ErrNotActivated) {
							// Avoid unnecessary subsequent Auth API calls.
							authIsActive = false
							return cb(proj)
						}
						return errors.Wrapf(err, "getting permissions for project %s", proj.Project)
					}
					proj.AuthInfo = &pfs.AuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
				}
				return cb(proj)
			})
		})
	}), "list projects")
}

// The ProjectInfo provided to the closure is repurposed on each invocation, so it's the client's responsibility to clone the ProjectInfo if desired
func (d *driver) listProjectInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, cb func(*pfs.ProjectInfo) error) error {
	return errors.Wrap(pfsdb.ForEachProject(ctx, txnCtx.SqlTx, func(project pfsdb.ProjectWithID) error {
		return cb(project.ProjectInfo)
	}), "list projects in transaction")
}

// TODO: delete all repos and pipelines within project
func (d *driver) deleteProject(ctx context.Context, txnCtx *txncontext.TransactionContext, project *pfs.Project, force bool) error {
	if err := project.ValidateName(); err != nil {
		return errors.Wrap(err, "invalid project name")
	}
	if err := d.env.Auth.CheckProjectIsAuthorizedInTransaction(txnCtx, project, auth.Permission_PROJECT_DELETE, auth.Permission_PROJECT_MODIFY_BINDINGS); err != nil {
		return errors.Wrapf(err, "user is not authorized to delete project %q", project)
	}
	var errs error
	repos, err := d.listRepoInTransaction(ctx, txnCtx, false, "", []*pfs.Project{project})
	if err != nil {
		return errors.Wrap(err, "list repos to determine if any still exist")
	}
	for _, repoInfo := range repos {
		errs = errors.Join(errs, fmt.Errorf("repo %v still exists", repoInfo.GetRepo()))
	}
	if errs != nil && !force {
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("cannot delete project %s: %v", project.Name, errs))
	}
	if err := pfsdb.DeleteProject(ctx, txnCtx.SqlTx, pfsdb.ProjectKey(project)); err != nil {
		return errors.Wrapf(err, "delete project %q", project)
	}
	if err := d.env.Auth.DeleteRoleBindingInTransaction(txnCtx, project.AuthResource()); err != nil {
		if !errors.Is(err, auth.ErrNotActivated) {
			return errors.Wrapf(err, "delete role binding for project %q", project)
		}
	}
	return nil
}

// Set child.ParentCommit (if 'parent' has been determined).
func (d *driver) linkParent(ctx context.Context, txnCtx *txncontext.TransactionContext, child *pfs.CommitInfo, parent *pfs.Commit, needsFinishedParent bool) error {
	if parent == nil {
		return nil
	}
	// Resolve 'parent' if it's a branch that isn't 'branch' (which can
	// happen if 'branch' is new and diverges from the existing branch in
	// 'parent').
	parentCommitInfo, err := d.resolveCommit(ctx, txnCtx.SqlTx, parent)
	if err != nil {
		return errors.Wrapf(err, "parent commit not found")
	}
	// fail if the parent commit has not been finished
	if needsFinishedParent && parentCommitInfo.Finishing == nil {
		return errors.Errorf("parent commit %s has not been finished", parentCommitInfo.Commit)
	}
	child.ParentCommit = parentCommitInfo.Commit
	return nil
}

// creates a new commit, and adds both commit ancestry, and commit provenance pointers
//
// NOTE: Requiring source commits to have finishing / finished parents ensures that the commits are not compacted
// in a pathological order (finishing later commits before earlier commits will result with us compacting
// the earlier commits multiple times).
func (d *driver) addCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, newCommitInfo *pfs.CommitInfo, parent *pfs.Commit, directProvenance []*pfs.Branch, needsFinishedParent bool) (pfsdb.CommitID, error) {
	if err := d.linkParent(ctx, txnCtx, newCommitInfo, parent, needsFinishedParent); err != nil {
		return 0, err
	}
	for _, prov := range directProvenance {
		b, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, prov)
		if err != nil {
			if pfsdb.IsNotFoundError(err) {
				return 0, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: prov})
			}
			return 0, errors.Wrap(err, "add commit")
		}
		newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, b.BranchInfo.Head)
	}
	commitID, err := pfsdb.CreateCommit(ctx, txnCtx.SqlTx, newCommitInfo)
	if err != nil {
		if errors.As(err, &pfsdb.CommitAlreadyExistsError{}) {
			return 0, errors.Join(err, pfsserver.ErrInconsistentCommit{Commit: newCommitInfo.Commit})
		}
		return 0, errors.EnsureStack(err)
	}
	for _, p := range newCommitInfo.DirectProvenance {
		if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommitInfo.Commit, p); err != nil {
			return 0, err
		}
	}
	return commitID, nil
}

// startCommit makes a new commit in 'branch', with the parent 'parent':
//   - 'parent' may be omitted, in which case the parent commit is inferred
//     from 'branch'.
//   - If 'parent' is set, it determines the parent commit, but 'branch' is
//     still moved to point at the new commit
func (d *driver) startCommit(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	parent *pfs.Commit,
	branch *pfs.Branch,
	description string,
) (*pfs.Commit, error) {
	// Validate arguments:
	if branch == nil || branch.Name == "" {
		return nil, errors.Errorf("branch must be specified")
	}
	// Check that caller is authorized
	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_WRITE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// New commit and commitInfo
	newCommitInfo := newUserCommitInfo(txnCtx, branch)
	newCommitInfo.Description = description
	if err := ancestry.ValidateName(branch.Name); err != nil {
		return nil, err
	}
	// Check if repo exists
	_, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, branch.Repo.Project.Name, branch.Repo.Name, branch.Repo.Type)
	if err != nil {
		if pfsdb.IsErrRepoNotFound(err) {
			return nil, pfsserver.ErrRepoNotFound{Repo: branch.Repo}
		}
		return nil, errors.EnsureStack(err)
	}
	updateCommitBranchField := false
	// update 'branch' (which must always be set) and set parent.ID (if 'parent'
	// was not set)
	b, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, branch)
	if err != nil {
		if !pfsdb.IsNotFoundError(err) {
			return nil, errors.Wrap(err, "start commit")
		}
		updateCommitBranchField = true // the branch field on the commit will have to be updated after the branch is created.
	}
	branchInfo := &pfs.BranchInfo{}
	if b != nil && b.BranchInfo != nil {
		branchInfo = b.BranchInfo
	}
	if branchInfo.Branch == nil {
		// New branch
		branchInfo.Branch = branch
	}
	// If the parent is unspecified, use the current head of the branch
	if parent == nil {
		parent = branchInfo.Head
	}
	commitID, err := d.addCommit(ctx, txnCtx, newCommitInfo, parent, branchInfo.DirectProvenance, true)
	if err != nil {
		return nil, err
	}
	// Point 'branch' at the new commit
	branchInfo.Head = newCommitInfo.Commit
	if _, err = pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, branchInfo); err != nil {
		return nil, errors.Wrap(err, "start commit")
	}
	// update branch field after a new branch is created.
	if updateCommitBranchField {
		newCommitInfo.Commit.Branch = branchInfo.Branch
		if err := pfsdb.UpdateCommit(ctx, txnCtx.SqlTx, commitID, newCommitInfo, pfsdb.AncestryOpt{SkipParent: true, SkipChildren: true}); err != nil {
			return nil, errors.Wrap(err, "start commit: update branch field after new branch is created")
		}
	}
	// check if this is happening in a spout pipeline, and alias the spec commit
	_, ok1 := os.LookupEnv(client.PPSPipelineNameEnv)
	_, ok2 := os.LookupEnv("PPS_SPEC_COMMIT")
	if !(ok1 && ok2) && len(branchInfo.Provenance) > 0 {
		// Otherwise, we don't allow user code to start commits on output branches
		return nil, pfsserver.ErrCommitOnOutputBranch{Branch: branch}
	}
	// Defer propagation of the commit until the end of the transaction so we can
	// batch downstream commits together if there are multiple changes.
	if err := txnCtx.PropagateBranch(branch); err != nil {
		return nil, err
	}
	return newCommitInfo.Commit, nil
}

func (d *driver) finishCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commit *pfs.Commit, description, commitError string, force bool) error {
	commitWithID, err := d.resolveCommitWithID(ctx, txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	commitInfo := commitWithID.CommitInfo
	if commitInfo.Finishing != nil {
		return pfsserver.ErrCommitFinished{
			Commit: commitInfo.Commit,
		}
	}
	if !force && len(commitInfo.DirectProvenance) > 0 {
		if info, err := d.env.GetPipelineInspector().InspectPipelineInTransaction(ctx, txnCtx, pps.RepoPipeline(commitInfo.Commit.Repo)); err != nil && !errutil.IsNotFoundError(err) {
			return errors.EnsureStack(err)
		} else if err == nil && info.Type == pps.PipelineInfo_PIPELINE_TYPE_TRANSFORM {
			return errors.Errorf("cannot finish a pipeline output or meta commit, use 'stop job' instead")
		}
		// otherwise, this either isn't a pipeline at all, or is a spout or service for which we should allow finishing
	}
	if description != "" {
		commitInfo.Description = description
	}
	commitInfo.Finishing = txnCtx.Timestamp
	commitInfo.Error = commitError

	return pfsdb.UpdateCommit(ctx, txnCtx.SqlTx, commitWithID.ID, commitInfo, pfsdb.AncestryOpt{SkipParent: true, SkipChildren: true})
}

func (d *driver) repoSize(ctx context.Context, txnCtx *txncontext.TransactionContext, repoInfo *pfs.RepoInfo) (int64, error) {
	for _, branch := range repoInfo.Branches {
		if branch.Name == "master" {
			branchInfo, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, branch)
			if err != nil {
				return 0, errors.Wrap(err, "repo size")
			}
			commit := branchInfo.Head
			for commit != nil {
				commitInfo, err := d.resolveCommit(ctx, txnCtx.SqlTx, commit)
				if err != nil {
					return 0, err
				}
				if commitInfo.Details != nil {
					return commitInfo.Details.SizeBytes, nil
				}
				commit = commitInfo.ParentCommit
			}
		}
	}
	return 0, nil
}

// propagateBranches starts commits downstream of 'branches'
// in order to restore the invariant that branch provenance matches HEAD commit
// provenance:
//
//	B.Head is provenant on A.Head <=>
//	branch B is provenant on branch A
//
// The implementation assumes that the invariant already holds for all branches
// upstream of 'branches', but not necessarily for each 'branch' itself.
//
// In other words, propagateBranches scans all branches b_downstream that are
// equal to or downstream of 'branches', and if the HEAD of b_downstream isn't
// provenant on the HEADs of b_downstream's provenance, propagateBranches starts
// a new HEAD commit in b_downstream that is. For example, propagateBranches
// starts downstream output commits (which trigger PPS jobs) when new input
// commits arrive on 'branch', when 'branches's HEAD is deleted, or when
// 'branches' are newly created (i.e. in CreatePipeline).
func (d *driver) propagateBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branches []*pfs.Branch) error {
	if len(branches) == 0 {
		return nil
	}
	var propagatedBranches []*pfs.BranchInfo
	seen := make(map[string]*pfs.BranchInfo)
	for _, b := range branches {
		biWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, b)
		if err != nil {
			return errors.Wrapf(err, "propagate branches: get branch %q", pfsdb.BranchKey(b))
		}
		for _, subvenantB := range biWithID.Subvenance {
			if _, ok := seen[pfsdb.BranchKey(subvenantB)]; !ok {
				subvenantBiWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, subvenantB)
				if err != nil {
					return errors.Wrapf(err, "propagate branches: get subvenant branch %q", pfsdb.BranchKey(subvenantB))
				}
				branchInfo := subvenantBiWithID.BranchInfo
				seen[pfsdb.BranchKey(subvenantB)] = branchInfo
				propagatedBranches = append(propagatedBranches, branchInfo)
			}
		}
	}
	sort.Slice(propagatedBranches, func(i, j int) bool {
		return len(propagatedBranches[i].Provenance) < len(propagatedBranches[j].Provenance)
	})
	// add new commits, set their ancestry + provenance pointers, and advance branch heads
	for _, bi := range propagatedBranches {
		// TODO(acohen4): can we just make calls to addCommit() here?
		// Do not propagate an open commit onto spout output branches (which should
		// only have a single provenance on a spec commit)
		if len(bi.Provenance) == 1 && bi.Provenance[0].Repo.Type == pfs.SpecRepoType {
			continue
		}
		// necessary when upstream and downstream branches are created in the same pachyderm transaction.
		// ex. when the spec and meta repos are created, the meta commit is already created at branch create time.
		if bi.GetHead().GetId() == txnCtx.CommitSetID {
			continue
		}
		newCommit := &pfs.Commit{
			Repo:   bi.Branch.Repo,
			Branch: bi.Branch,
			Id:     txnCtx.CommitSetID,
		}
		newCommitInfo := &pfs.CommitInfo{
			Commit:  newCommit,
			Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
			Started: txnCtx.Timestamp,
		}
		// enumerate the new commit's provenance
		for _, b := range bi.DirectProvenance {
			var provCommit *pfs.Commit
			if pbi, ok := seen[pfsdb.BranchKey(b)]; ok {
				provCommit = client.NewCommit(pbi.Branch.Repo.Project.Name, pbi.Branch.Repo.Name, pbi.Branch.Name, txnCtx.CommitSetID)
			} else {
				provBranchInfo, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, b)
				if err != nil {
					return errors.Wrapf(err, "get provenant branch %q", pfsdb.BranchKey(b))
				}
				provCommit = provBranchInfo.Head
			}
			newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, provCommit)
		}
		// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
		newCommitInfo.ParentCommit = proto.Clone(bi.Head).(*pfs.Commit)
		bi.Head = newCommit
		// create open 'commit'.
		commitID, err := pfsdb.CreateCommit(ctx, txnCtx.SqlTx, newCommitInfo)
		if err != nil {
			return errors.Wrapf(err, "create new commit %q", pfsdb.CommitKey(newCommit))
		}
		if newCommitInfo.ParentCommit != nil {
			if err := pfsdb.CreateCommitParent(ctx, txnCtx.SqlTx, newCommitInfo.ParentCommit, commitID); err != nil {
				return errors.Wrapf(err, "update parent commit %q with child %q", pfsdb.CommitKey(newCommitInfo.ParentCommit), pfsdb.CommitKey(newCommit))
			}
		}
		// add commit provenance
		for _, c := range newCommitInfo.DirectProvenance {
			if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommit, c); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q", pfsdb.CommitKey(newCommit), pfsdb.CommitKey(c))
			}
		}
		// update branch head.
		if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, bi); err != nil {
			return errors.Wrapf(err, "put branch %q with head %q", pfsdb.BranchKey(bi.Branch), pfsdb.CommitKey(bi.Head))
		}
	}
	txnCtx.PropagateJobs()
	return nil
}

// inspectCommit takes a Commit and returns the corresponding CommitInfo.
//
// As a side effect, this function also replaces the ID in the given commit
// with a real commit ID.
func (d *driver) inspectCommit(ctx context.Context, commit *pfs.Commit, wait pfs.CommitState) (*pfs.CommitInfo, error) {
	commitInfo, err := d.resolveCommitWithAuth(ctx, commit)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished == nil {
		switch wait {
		case pfs.CommitState_STARTED:
		case pfs.CommitState_READY:
			for _, c := range commitInfo.DirectProvenance {
				if _, err := d.inspectCommit(ctx, c, pfs.CommitState_FINISHED); err != nil {
					return nil, err
				}
			}
		case pfs.CommitState_FINISHING, pfs.CommitState_FINISHED:
			return d.inspectProcessingCommits(ctx, commitInfo, wait)
		}
	}
	return commitInfo, nil
}

// inspectProcessingCommits waits for the commit to be FINISHING or FINISHED.
func (d *driver) inspectProcessingCommits(ctx context.Context, commitInfo *pfs.CommitInfo, wait pfs.CommitState) (*pfs.CommitInfo, error) {
	var commitID pfsdb.CommitID
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitID, err = pfsdb.GetCommitID(ctx, txnCtx.SqlTx, commitInfo.Commit)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// We only cancel the watcher if we detect the commit is the right state.
	expectedErr := errors.New("commit is in the right state")
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	if err := pfsdb.WatchCommit(ctx, d.env.DB, d.env.Listener, commitID,
		func(id pfsdb.CommitID, ci *pfs.CommitInfo) error {
			switch wait {
			case pfs.CommitState_FINISHING:
				if ci.Finishing != nil {
					commitInfo = ci
					cancel(expectedErr)
					return nil
				}
			case pfs.CommitState_FINISHED:
				if ci.Finished != nil {
					commitInfo = ci
					cancel(expectedErr)
					return nil
				}
			}
			return nil
		},
		func(id pfsdb.CommitID) error {
			return pfsserver.ErrCommitDeleted{Commit: commitInfo.Commit}
		},
	); err != nil && !errors.Is(context.Cause(ctx), expectedErr) {
		return nil, errors.Wrap(err, "inspect finishing or finished commit")
	}
	return commitInfo, nil
}

// resolveCommitWithAuth is like resolveCommit, but it does some pre-resolution checks like repo authorization.
func (d *driver) resolveCommitWithAuth(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	if commit.Repo.Name == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "FileSet - Virtual Commit",
			Finished:    &timestamppb.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return cinfo, nil
	}
	if commit == nil {
		return nil, errors.Errorf("cannot inspect nil commit")
	}
	if err := d.env.Auth.CheckRepoIsAuthorized(ctx, commit.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// Resolve the commit in case it specifies a branch head or commit ancestry
	var commitInfo *pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfo, err = d.resolveCommit(ctx, txnCtx.SqlTx, commit)
		return err
	}); err != nil {
		return nil, err
	}
	return commitInfo, nil
}

// resolveCommitWithID contains the essential implementation of inspectCommit: it converts 'commit' (which may
// be a commit ID or branch reference, plus '~' and/or '^') to a repo + commit
// ID. It accepts a postgres transaction so that it can be used in a transaction
// and avoids an inconsistent call to d.inspectCommit()
func (d *driver) resolveCommitWithID(ctx context.Context, sqlTx *pachsql.Tx, userCommit *pfs.Commit) (*pfsdb.CommitWithID, error) {
	if userCommit == nil {
		return nil, errors.Errorf("cannot resolve nil commit")
	}
	if userCommit.Repo == nil {
		return nil, errors.Errorf("cannot resolve commit with no repo")
	}
	if userCommit.Id == "" && userCommit.GetBranch().GetName() == "" {
		return nil, errors.Errorf("cannot resolve commit with no ID or branch")
	}
	commit := proto.Clone(userCommit).(*pfs.Commit) // back up user commit, for error reporting
	// Extract any ancestor tokens from 'commit.ID' (i.e. ~, ^ and .)
	var ancestryLength int
	var err error
	commit.Id, ancestryLength, err = ancestry.Parse(commit.Id)
	if err != nil {
		return nil, err
	}
	// Now that ancestry has been parsed out, check if the ID is a branch name
	if commit.Id != "" && !uuid.IsUUIDWithoutDashes(commit.Id) {
		if commit.Branch.GetName() != "" {
			return nil, errors.Errorf("invalid commit ID given with a branch (%s): %s\n", commit.Branch, commit.Id)
		}
		commit.Branch = commit.Repo.NewBranch(commit.Id)
		commit.Id = ""
	}
	// If commit.ID is unspecified, get it from the branch head
	if commit.Id == "" {
		branchInfo, err := pfsdb.GetBranchInfoWithID(ctx, sqlTx, commit.Branch)
		if err != nil {
			if pfsdb.IsNotFoundError(err) {
				return nil, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: commit.Branch})
			}
			return nil, errors.Wrap(err, "resolve commit")
		}
		commit.Id = branchInfo.Head.Id
	}
	commitWithID, err := pfsdb.GetCommitWithIDByKey(ctx, sqlTx, commit)
	if err != nil {
		if pfsdb.IsNotFoundError(err) {
			// try to resolve to alias if not found
			resolvedCommit, err := pfsdb.ResolveCommitProvenance(sqlTx, userCommit.Repo, commit.Id)
			if err != nil {
				return nil, err
			}
			commit.Id = resolvedCommit.Id
			commitWithID, err = pfsdb.GetCommitWithIDByKey(ctx, sqlTx, commit)
			if err != nil {
				return nil, errors.EnsureStack(err)
			}
		} else {
			return nil, errors.Wrap(err, "resolve commit")
		}
	}
	// Traverse commits' parents until you've reached the right ancestor
	if ancestryLength >= 0 {
		for i := 1; i <= ancestryLength; i++ {
			if commitWithID.CommitInfo.ParentCommit == nil {
				return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
			}
			parent := commitWithID.CommitInfo.ParentCommit
			commitWithID, err = pfsdb.GetCommitWithIDByKey(ctx, sqlTx, parent)
			if err != nil {
				if pfsdb.IsNotFoundError(err) {
					if i == 0 {
						return nil, errors.Join(err, pfsserver.ErrCommitNotFound{Commit: userCommit})
					}
					return nil, errors.Join(err, pfsserver.ErrParentCommitNotFound{Commit: commit})
				}
				return nil, errors.EnsureStack(err)
			}
		}
	} else {
		cis := make([]*pfsdb.CommitWithID, ancestryLength*-1)
		for i := 0; ; i++ {
			if commit == nil {
				if i >= len(cis) {
					commitWithID = cis[i%len(cis)]
					break
				}
				return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
			}
			cis[i%len(cis)], err = pfsdb.GetCommitWithIDByKey(ctx, sqlTx, commit)
			if err != nil {
				if pfsdb.IsNotFoundError(err) {
					if i == 0 {
						return nil, errors.Join(err, pfsserver.ErrCommitNotFound{Commit: userCommit})
					}
					return nil, errors.Join(err, pfsserver.ErrParentCommitNotFound{Commit: commit})
				}
				return nil, err
			}
			commit = cis[i%len(cis)].CommitInfo.ParentCommit
		}
	}
	return commitWithID, nil
}

// resolveCommit contains the essential implementation of inspectCommit: it converts 'commit' (which may
// be a commit ID or branch reference, plus '~' and/or '^') to a repo + commit
// ID. It accepts a postgres transaction so that it can be used in a transaction
// and avoids an inconsistent call to d.inspectCommit()
func (d *driver) resolveCommit(ctx context.Context, sqlTx *pachsql.Tx, userCommit *pfs.Commit) (*pfs.CommitInfo, error) {
	commitWithId, err := d.resolveCommitWithID(ctx, sqlTx, userCommit)
	if err != nil {
		return nil, err
	}
	return commitWithId.CommitInfo, nil
}

// getCommit is like inspectCommit, without the blocking.
// It does not add the size to the CommitInfo
//
// TODO(acohen4): consider more an architecture where a commit is resolved at the API boundary
func (d *driver) getCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	if commit.AccessRepo().Name == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "FileSet - Virtual Commit",
			Finished:    &timestamppb.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return cinfo, nil
	}
	if commit == nil {
		return nil, errors.Errorf("cannot inspect nil commit")
	}

	// Check if the commitID is a branch name
	var commitInfo *pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfo, err = d.resolveCommit(ctx, txnCtx.SqlTx, commit)
		return err
	}); err != nil {
		return nil, err
	}
	return commitInfo, nil
}

// passesCommitOriginFilter is a helper function for listCommit and
// subscribeCommit to apply filtering to the returned commits.  By default
// we allow users to request all the commits with
// 'all', or a specific type of commit with 'originKind'.
func passesCommitOriginFilter(commitInfo *pfs.CommitInfo, all bool, originKind pfs.OriginKind) bool {
	if all {
		return true
	}
	if originKind != pfs.OriginKind_ORIGIN_KIND_UNKNOWN {
		return commitInfo.Origin.Kind == originKind
	}
	return true
}

func (d *driver) listCommit(
	ctx context.Context,
	repo *pfs.Repo,
	to *pfs.Commit,
	from *pfs.Commit,
	startTime *timestamppb.Timestamp,
	number int64,
	reverse bool,
	all bool,
	originKind pfs.OriginKind,
	cb func(*pfs.CommitInfo) error,
) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	if err := d.env.Auth.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_LIST_COMMIT); err != nil {
		return errors.EnsureStack(err)
	}
	if from != nil && !proto.Equal(from.Repo, repo) || to != nil && !proto.Equal(to.Repo, repo) {
		return errors.Errorf("`from` and `to` commits need to be from repo %s", repo)
	}
	// Make sure that the repo exists
	if repo.Name != "" {
		if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
			if _, err := pfsdb.GetRepoByName(ctx, tx, repo.Project.Name, repo.Name, repo.Type); err != nil {
				if pfsdb.IsErrRepoNotFound(err) {
					return pfsserver.ErrRepoNotFound{Repo: repo}
				}
				return errors.EnsureStack(err)
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "ensure repo exists")
		}
	}

	// Make sure that both from and to are valid commits
	if from != nil {
		if _, err := d.inspectCommit(ctx, from, pfs.CommitState_STARTED); err != nil {
			return err
		}
	}
	if to != nil {
		ci, err := d.inspectCommit(ctx, to, pfs.CommitState_STARTED)
		if err != nil {
			return err
		}
		to = ci.Commit
	}

	// if number is 0, we return all commits that match the criteria
	if number == 0 {
		number = math.MaxInt64
	}

	if from != nil && to == nil {
		return errors.Errorf("cannot use `from` commit without `to` commit")
	} else if from == nil && to == nil {
		// we hold onto a revisions worth of cis so that we can sort them by provenance
		var cis []*pfs.CommitInfo
		// sendCis sorts cis and passes them to f
		sendCis := func() error {
			// We don't sort these because there is no provenance between commits
			// within a repo, so there is no topological sort necessary.
			for i, ci := range cis {
				if number == 0 {
					return errutil.ErrBreak
				}
				number--

				if reverse {
					ci = cis[len(cis)-1-i]
				}
				var err error
				ci.SizeBytesUpperBound, err = d.commitSizeUpperBound(ctx, ci.Commit)
				if err != nil && !pfsserver.IsBaseCommitNotFinishedErr(err) {
					return err
				}
				if err := cb(ci); err != nil {
					return err
				}
			}
			cis = nil
			return nil
		}
		lastRev := int64(-1)
		listCallback := func(ci *pfs.CommitInfo, createRev int64) error {
			if createRev != lastRev {
				if err := sendCis(); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
				lastRev = createRev
			}
			if passesCommitOriginFilter(ci, all, originKind) {
				if startTime != nil {
					createdAt := time.Unix(ci.Started.GetSeconds(), int64(ci.Started.GetNanos())).UTC()
					fromTime := time.Unix(startTime.GetSeconds(), int64(startTime.GetNanos())).UTC()
					if !reverse && createdAt.Before(fromTime) || reverse && createdAt.After(fromTime) {
						cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
					}
					return nil
				}
				cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
			}
			return nil
		}

		// if neither from and to is given, we list all commits in
		// the repo, sorted by revision timestamp.
		var filter *pfs.Commit
		if repo.Name != "" {
			filter = &pfs.Commit{Repo: repo}
		}
		// driver.listCommit should return more recent commits by default, which is the
		// opposite behavior of pfsdb.ForEachCommit.
		order := pfsdb.SortOrderDesc
		if reverse {
			order = pfsdb.SortOrderAsc
		}
		if err := pfsdb.ForEachCommit(ctx, d.env.DB, filter, func(commitWithID pfsdb.CommitWithID) error {
			return listCallback(commitWithID.CommitInfo, commitWithID.Revision)
		}, pfsdb.OrderByCommitColumn{Column: pfsdb.CommitColumnID, Order: order}); err != nil {
			return errors.Wrap(err, "list commit")
		}
		// Call sendCis one last time to send whatever's pending in 'cis'
		if err := sendCis(); err != nil && !errors.Is(err, errutil.ErrBreak) {
			return err
		}
	} else {
		if reverse {
			return errors.Errorf("cannot use 'Reverse' while also using 'From' or 'To'")
		}
		cursor := to
		for number != 0 && cursor != nil && (from == nil || cursor.Id != from.Id) {
			var commitInfo *pfs.CommitInfo
			var err error
			if err := dbutil.WithTx(ctx, d.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
				commitInfo, err = pfsdb.GetCommitByCommitKey(ctx, tx, cursor)
				return err
			}); err != nil {
				return errors.Wrap(err, "list commit")
			}
			if passesCommitOriginFilter(commitInfo, all, originKind) {
				var err error
				commitInfo.SizeBytesUpperBound, err = d.commitSizeUpperBound(ctx, commitInfo.Commit)
				if err != nil && !pfsserver.IsBaseCommitNotFinishedErr(err) {
					return err
				}
				if err := cb(commitInfo); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
				number--
			}
			cursor = commitInfo.ParentCommit
		}
	}
	return nil
}

func (d *driver) subscribeCommit(
	ctx context.Context,
	repo *pfs.Repo,
	branch string,
	from *pfs.Commit,
	state pfs.CommitState,
	all bool,
	originKind pfs.OriginKind,
	cb func(*pfs.CommitInfo) error,
) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}
	if from != nil && !proto.Equal(from.Repo, repo) {
		return errors.Errorf("the `from` commit needs to be from repo %s", repo)
	}
	// keep track of the commits that have been sent
	seen := make(map[string]bool)
	var repoID pfsdb.RepoID
	var err error
	if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		repoID, err = pfsdb.GetRepoID(ctx, tx, repo.Project.Name, repo.Name, repo.Type)
		return errors.Wrap(err, "get repo ID")
	}); err != nil {
		return err
	}
	return pfsdb.WatchCommitsInRepo(ctx, d.env.DB, d.env.Listener, repoID,
		func(id pfsdb.CommitID, commitInfo *pfs.CommitInfo) error { // onUpsert
			// if branch is provided, make sure the commit was created on that branch
			if branch != "" && commitInfo.Commit.Branch.Name != branch {
				return nil
			}
			// If the origin of the commit doesn't match what we're interested in, skip it
			if !passesCommitOriginFilter(commitInfo, all, originKind) {
				return nil
			}
			// We don't want to include the `from` commit itself
			if !(seen[commitInfo.Commit.Id] || (from != nil && from.Id == commitInfo.Commit.Id)) {
				// Wait for the commit to enter the right state
				commitInfo, err := d.inspectCommit(ctx, proto.Clone(commitInfo.Commit).(*pfs.Commit), state)
				if err != nil {
					return err
				}
				if err := cb(commitInfo); err != nil {
					return err
				}
				seen[commitInfo.Commit.Id] = true
			}
			return nil
		},
		func(id pfsdb.CommitID) error { // onDelete
			return nil
		},
	)
}

func (d *driver) clearCommit(ctx context.Context, commit *pfs.Commit) error {
	commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finishing != nil {
		return errors.Errorf("cannot clear finished commit")
	}
	return errors.EnsureStack(d.commitStore.DropFileSets(ctx, commit))
}

func (d *driver) squashCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commit *pfs.Commit, recursive bool) error {
	commitInfo, err := d.resolveCommit(ctx, txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	commit = commitInfo.Commit
	commits, err := pfsdb.GetCommitSubvenance(ctx, txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	commitInfos := []*pfs.CommitInfo{commitInfo}
	for _, c := range commits {
		ci, err := pfsdb.GetCommitByCommitKey(ctx, txnCtx.SqlTx, c)
		if err != nil {
			return err
		}
		commitInfos = append(commitInfos, ci)
	}
	if len(commitInfos) > 1 && !recursive {
		return errors.Errorf("cannot squash commit (%v) with subvenance without recursive", commit)
	}
	for _, ci := range commitInfos {
		if ci.Commit.Branch.Repo.Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		if len(ci.ChildCommits) == 0 {
			return &pfsserver.ErrSquashWithoutChildren{Commit: ci.Commit}
		}
	}
	for _, ci := range commitInfos {
		if err := d.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
		txnCtx.StopJob(ci.Commit)
	}
	return nil
}

func (d *driver) dropCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commit *pfs.Commit, recursive bool) error {
	commitInfo, err := d.resolveCommit(ctx, txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	commit = commitInfo.Commit
	commits, err := pfsdb.GetCommitSubvenance(ctx, txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	commitInfos := []*pfs.CommitInfo{commitInfo}
	for _, c := range commits {
		ci, err := pfsdb.GetCommitByCommitKey(ctx, txnCtx.SqlTx, c)
		if err != nil {
			return err
		}
		commitInfos = append(commitInfos, ci)
	}
	if len(commitInfos) > 1 && !recursive {
		return errors.Errorf("cannot drop commit (%v) with subvenance without recursive", commit)
	}
	for _, ci := range commitInfos {
		if len(ci.ChildCommits) > 0 {
			return &pfsserver.ErrDropWithChildren{Commit: ci.Commit}
		}
	}
	for _, ci := range commitInfos {
		if err := d.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
		txnCtx.StopJob(ci.Commit)
	}
	return nil
}

// fillNewBranches helps create the upstream branches on which a branch is provenant, if they don't exist.
// TODO(provenance): consider removing this functionality
func (d *driver) fillNewBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, provenance []*pfs.Branch) error {
	branch.Repo.EnsureProject()
	newRepoCommits := make(map[string]*pfsdb.CommitWithID)
	for _, p := range provenance {
		biWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, p)
		if err != nil && !pfsdb.IsNotFoundError(err) {
			return errors.Wrap(err, "get branch info")
		}
		if biWithID != nil {
			// Skip because the branch already exists.
			continue
		}
		var updateHead *pfsdb.CommitWithID
		head, ok := newRepoCommits[p.Repo.Key()]
		if !ok {
			head, err = d.makeEmptyCommit(ctx, txnCtx, p, nil, nil)
			if err != nil {
				return errors.Wrap(err, "create empty commit")
			}
			updateHead = head
			newRepoCommits[p.Repo.Key()] = head
		}
		branchInfo := &pfs.BranchInfo{Branch: p, Head: head.CommitInfo.Commit}
		if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, branchInfo); err != nil {
			return errors.Wrap(err, "upsert branch")
		}
		// Now that the branch exists, add the branch_id to the head commit.
		if updateHead != nil {
			if err := pfsdb.UpdateCommit(ctx, txnCtx.SqlTx, updateHead.ID, updateHead.CommitInfo); err != nil {
				return errors.Wrap(err, "update commit")
			}
		}
	}
	return nil
}

// for a DAG to be valid, it may not have a multiple branches from the same repo
// reachable by traveling edges bidirectionally. The reason is that this would complicate resolving
func (d *driver) validateDAGStructure(ctx context.Context, txnCtx *txncontext.TransactionContext, bs []*pfs.Branch) error {
	cache := make(map[string]*pfs.BranchInfo)
	getBranchInfo := func(b *pfs.Branch) (*pfs.BranchInfo, error) {
		if bi, ok := cache[pfsdb.BranchKey(b)]; ok {
			return bi, nil
		}
		biWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, b)
		if err != nil {
			return nil, errors.Wrapf(err, "get branch info")
		}
		cache[pfsdb.BranchKey(b)] = biWithID.BranchInfo
		return biWithID.BranchInfo, nil
	}
	for _, b := range bs {
		dagBranches := make(map[string]*pfs.BranchInfo)
		bi, err := getBranchInfo(b)
		if err != nil {
			return errors.Wrapf(err, "loading branch info for %q during validation", b.String())
		}
		dagBranches[pfsdb.BranchKey(b)] = bi
		expanded := make(map[string]struct{}) // expanded branches
		for {
			hasExpanded := false
			for _, bi := range dagBranches {
				if _, ok := expanded[pfsdb.BranchKey(bi.Branch)]; ok {
					continue
				}
				hasExpanded = true
				expanded[pfsdb.BranchKey(bi.Branch)] = struct{}{}
				for _, b := range bi.Provenance {
					bi, err := getBranchInfo(b)
					if err != nil {
						return err
					}
					dagBranches[pfsdb.BranchKey(b)] = bi
				}
				for _, b := range bi.Subvenance {
					bi, err := getBranchInfo(b)
					if err != nil {
						return err
					}
					dagBranches[pfsdb.BranchKey(b)] = bi
				}
			}
			if !hasExpanded {
				break
			}
		}
		foundRepos := make(map[string]struct{})
		for _, bi := range dagBranches {
			if _, ok := foundRepos[bi.Branch.Repo.String()]; ok {
				return &pfsserver.ErrInvalidProvenanceStructure{Branch: bi.Branch}
			}
			foundRepos[bi.Branch.Repo.String()] = struct{}{}
		}
	}
	return nil
}

func newUserCommitInfo(txnCtx *txncontext.TransactionContext, branch *pfs.Branch) *pfs.CommitInfo {
	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Branch: branch,
			Repo:   branch.Repo,
			Id:     txnCtx.CommitSetID,
		},
		Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_USER},
		Started: txnCtx.Timestamp,
		Details: &pfs.CommitInfo_Details{},
	}
}

// createBranch creates a new branch or updates an existing branch (must be one
// or the other). Most importantly, it sets 'branch.DirectProvenance' to
// 'provenance' and then for all (downstream) branches, restores the invariant:
//
//	∀ b . b.Provenance = ∪ b'.Provenance (where b' ∈ b.DirectProvenance)
//
// This invariant is assumed to hold for all branches upstream of 'branch', but not
// for 'branch' itself once 'b.Provenance' has been set.
//
// i.e. up to one branch in a repo can be present within a DAG
func (d *driver) createBranch(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, commit *pfs.Commit, provenance []*pfs.Branch, trigger *pfs.Trigger) error {
	// Validate arguments
	if branch == nil {
		return errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return errors.New("branch repo cannot be nil")
	}
	if err := d.validateTrigger(ctx, txnCtx, branch, trigger); err != nil {
		return err
	}
	if len(provenance) > 0 && trigger != nil {
		return errors.New("a branch cannot have both provenance and a trigger")
	}
	var err error
	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_CREATE_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	// Validate request
	var allBranches []*pfs.Branch
	allBranches = append(allBranches, provenance...)
	allBranches = append(allBranches, branch)
	for _, b := range allBranches {
		if err := ancestry.ValidateName(b.Name); err != nil {
			return err
		}
	}
	// Create any of the provenance branches that don't exist yet, and give them an empty commit
	if err := d.fillNewBranches(ctx, txnCtx, branch, provenance); err != nil {
		return err
	}
	// if the user passed a commit to point this branch at, resolve it
	if commit != nil {
		ci, err := d.resolveCommit(ctx, txnCtx.SqlTx, commit)
		if err != nil {
			return errors.Wrapf(err, "unable to inspect %s", commit)
		}
		commit = ci.Commit
	}
	// retrieve the current version of this branch and set its head if specified
	var oldProvenance []*pfs.Branch
	propagate := false
	biWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, branch)
	if err != nil && !pfsdb.IsNotFoundError(err) {
		return errors.Wrap(err, "create branch")
	}
	branchInfo := &pfs.BranchInfo{}
	if biWithID != nil {
		branchInfo = biWithID.BranchInfo
	}
	branchInfo.Branch = branch
	oldProvenance = branchInfo.DirectProvenance
	branchInfo.DirectProvenance = nil
	for _, provBranch := range provenance {
		if proto.Equal(provBranch.Repo, branch.Repo) {
			return errors.Errorf("repo %s cannot be in the provenance of its own branch", branch.Repo)
		}
		add(&branchInfo.DirectProvenance, provBranch)
	}
	if commit != nil {
		branchInfo.Head = commit
		propagate = true
	}
	var updateHead *pfsdb.CommitWithID
	// if we don't have a branch head, or the provenance has changed, add a new commit to the branch to capture the changed structure
	// the one edge case here, is that it's undesirable to add a commit in the case where provenance is completely removed...
	//
	// TODO(provenance): This sort of hurts Branch Provenance invariant. See if we can re-assess....
	if branchInfo.Head == nil || (!same(oldProvenance, provenance) && len(provenance) != 0) {
		c, err := d.makeEmptyCommit(ctx, txnCtx, branch, provenance, branchInfo.Head)
		if err != nil {
			return err
		}
		branchInfo.Head = c.CommitInfo.Commit
		updateHead = c
		propagate = true
	}
	if trigger != nil && trigger.Branch != "" {
		branchInfo.Trigger = trigger
	}
	if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, branchInfo); err != nil {
		return errors.Wrap(err, "create branch")
	}
	// need to update head commit's branch_id after the branch is created.
	if updateHead != nil {
		if err := pfsdb.UpdateCommit(ctx, txnCtx.SqlTx, updateHead.ID, updateHead.CommitInfo, pfsdb.AncestryOpt{SkipParent: true, SkipChildren: true}); err != nil {
			return errors.Wrap(err, "create branch")
		}
	}
	// propagate the head commit to 'branch'. This may also modify 'branch', by
	// creating a new HEAD commit if 'branch's provenance was changed and its
	// current HEAD commit has old provenance
	if propagate {
		return txnCtx.PropagateBranch(branch)
	}
	return nil
}

func (d *driver) inspectBranch(ctx context.Context, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	branchInfo := &pfs.BranchInfo{}
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		branchInfo, err = d.inspectBranchInTransaction(ctx, txnCtx, branch)
		return err
	}); err != nil {
		if branch != nil && branch.Repo != nil && branch.Repo.Project != nil && pfsdb.IsNotFoundError(err) {
			return nil, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: branch})
		}
		return nil, err
	}
	return branchInfo, nil
}

func (d *driver) inspectBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	// Validate arguments
	if branch == nil {
		return nil, errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return nil, errors.New("branch repo cannot be nil")
	}
	if branch.Repo.Project == nil {
		branch.Repo.Project = &pfs.Project{Name: "default"}
	}

	// Check that the user is logged in, but don't require any access level
	if _, err := txnCtx.WhoAmI(); err != nil {
		if !auth.IsErrNotActivated(err) {
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to inspect a branch)")
		}
	}

	result, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, branch)
	if err != nil {
		return nil, errors.Wrap(err, "inspect branch in transaction")
	}
	return result.BranchInfo, nil
}

func (d *driver) listBranch(ctx context.Context, reverse bool, cb func(*pfs.BranchInfo) error) error {
	if _, err := d.env.Auth.WhoAmI(ctx, &auth.WhoAmIRequest{}); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	} else if err == nil {
		return errors.New("Cannot list branches from all repos with auth activated")
	}

	var bis []*pfs.BranchInfo
	sendBis := func() error {
		if !reverse {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) < len(bis[j].Provenance) })
		} else {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) > len(bis[j].Provenance) })
		}
		for i := range bis {
			if err := cb(bis[i]); err != nil {
				return err
			}
		}
		bis = nil
		return nil
	}

	lastRev := int64(-1)
	listCallback := func(branchInfoWithID pfsdb.BranchInfoWithID) error {
		if branchInfoWithID.Revision != lastRev {
			if err := sendBis(); err != nil {
				return errors.EnsureStack(err)
			}
			lastRev = branchInfoWithID.Revision
		}
		bis = append(bis, proto.Clone(branchInfoWithID.BranchInfo).(*pfs.BranchInfo))
		return nil
	}

	order := pfsdb.SortOrderDesc
	if reverse {
		order = pfsdb.SortOrderAsc
	}
	orderBys := []pfsdb.OrderByBranchColumn{
		{Column: pfsdb.BranchColumnCreatedAt, Order: order},
		{Column: pfsdb.BranchColumnID, Order: order},
	}
	if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		return pfsdb.ForEachBranch(ctx, tx, nil, listCallback, orderBys...)
	}); err != nil {
		return errors.Wrap(err, "list branches")
	}
	return sendBis()
}

func (d *driver) listBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, reverse bool, cb func(*pfs.BranchInfo) error) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_LIST_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	// Make sure that the repo exists
	if repo.Name != "" {
		if _, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type); err != nil {
			if pfsdb.IsErrRepoNotFound(err) {
				return pfsserver.ErrRepoNotFound{Repo: repo}
			}
			return errors.EnsureStack(err)
		}
	}
	order := pfsdb.SortOrderDesc
	if reverse {
		order = pfsdb.SortOrderAsc
	}
	orderBys := []pfsdb.OrderByBranchColumn{
		{Column: pfsdb.BranchColumnCreatedAt, Order: order},
		{Column: pfsdb.BranchColumnID, Order: order},
	}
	return errors.Wrap(pfsdb.ForEachBranch(ctx, txnCtx.SqlTx, &pfs.Branch{Repo: repo}, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
		return cb(branchInfoWithID.BranchInfo)
	}, orderBys...), "list branch in transaction")
}

func (d *driver) deleteBranch(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, force bool) error {
	if err := d.env.Auth.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_DELETE_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	branchInfoWithID, err := pfsdb.GetBranchInfoWithID(ctx, txnCtx.SqlTx, branch)
	if err != nil {
		return errors.Wrapf(err, "get branch %q", branch.Key())
	}
	return pfsdb.DeleteBranch(ctx, txnCtx.SqlTx, branchInfoWithID, force)
}

func (d *driver) deleteAll(ctx context.Context) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		if _, err := d.deleteReposInTransaction(ctx, txnCtx, nil /* projects */, true /* force */); err != nil {
			return errors.Wrap(err, "could not delete all repos")
		}
		if err := d.listProjectInTransaction(ctx, txnCtx, func(pi *pfs.ProjectInfo) error {
			return errors.Wrapf(d.deleteProject(ctx, txnCtx, pi.Project, true /* force */), "delete project %q", pi.Project.String())
		}); err != nil {
			return err
		} // now that the cluster is empty, recreate the default project
		return d.createProjectInTransaction(ctx, txnCtx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "default"}})
	})
}

// only transform source repos and spouts get a closed commit
func (d *driver) makeEmptyCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, directProvenance []*pfs.Branch, parent *pfs.Commit) (*pfsdb.CommitWithID, error) {
	// Input repos want a closed head commit, so decide if we leave
	// it open by the presence of branch provenance.
	closed := true
	if len(directProvenance) > 0 {
		closed = false
	}
	commit := branch.NewCommit(txnCtx.CommitSetID)
	commit.Repo = branch.Repo
	commitInfo := &pfs.CommitInfo{
		Commit:  commit,
		Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
		Started: txnCtx.Timestamp,
	}
	if closed {
		commitInfo.Finishing = txnCtx.Timestamp
		commitInfo.Finished = txnCtx.Timestamp
		commitInfo.Details = &pfs.CommitInfo_Details{}
	}
	commitId, err := d.addCommit(ctx, txnCtx, commitInfo, parent, directProvenance, false /* needsFinishedParent */)
	if err != nil {
		return nil, err
	}
	if closed {
		total, err := d.storage.Filesets.ComposeTx(txnCtx.SqlTx, nil, defaultTTL)
		if err != nil {
			return nil, err
		}
		if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commit, *total); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	return &pfsdb.CommitWithID{
		ID:         commitId,
		CommitInfo: commitInfo,
	}, nil
}

func (d *driver) putCache(ctx context.Context, key string, value *anypb.Any, fileSetIds []fileset.ID, tag string) error {
	return d.cache.Put(ctx, key, value, fileSetIds, tag)
}

func (d *driver) getCache(ctx context.Context, key string) (*anypb.Any, error) {
	return d.cache.Get(ctx, key)
}

func (d *driver) clearCache(ctx context.Context, tagPrefix string) error {
	return d.cache.Clear(ctx, tagPrefix)
}

// TODO: Is this really necessary?
type branchSet []*pfs.Branch

func (b *branchSet) search(branch *pfs.Branch) (int, bool) {
	key := pfsdb.BranchKey(branch)
	i := sort.Search(len(*b), func(i int) bool {
		return pfsdb.BranchKey((*b)[i]) >= key
	})
	if i == len(*b) {
		return i, false
	}
	return i, pfsdb.BranchKey((*b)[i]) == pfsdb.BranchKey(branch)
}

func (b *branchSet) add(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if !ok {
		*b = append(*b, nil)
		copy((*b)[i+1:], (*b)[i:])
		(*b)[i] = branch
	}
}

func add(bs *[]*pfs.Branch, branch *pfs.Branch) {
	(*branchSet)(bs).add(branch)
}

func (b *branchSet) has(branch *pfs.Branch) bool {
	_, ok := b.search(branch)
	return ok
}

func has(bs *[]*pfs.Branch, branch *pfs.Branch) bool {
	return (*branchSet)(bs).has(branch)
}

func same(bs []*pfs.Branch, branches []*pfs.Branch) bool {
	if len(bs) != len(branches) {
		return false
	}
	for _, br := range branches {
		if !has(&bs, br) {
			return false
		}
	}
	return true
}
