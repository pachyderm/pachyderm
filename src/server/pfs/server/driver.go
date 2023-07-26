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
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
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
	repos    col.PostgresCollection
	commits  col.PostgresCollection
	branches col.PostgresCollection

	storage     *storage.Server
	commitStore commitStore

	cache *fileset.Cache
}

func newDriver(env Env) (*driver, error) {
	objClient := env.ObjectClient
	// test object storage.
	if err := func() error {
		ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
		defer cf()
		return obj.TestStorage(ctx, objClient)
	}(); err != nil {
		return nil, err
	}
	repos := pfsdb.Repos(env.DB, env.Listener)
	commits := pfsdb.Commits(env.DB, env.Listener)
	branches := pfsdb.Branches(env.DB, env.Listener)

	// Setup driver struct.
	d := &driver{
		env:        env,
		etcdClient: env.EtcdClient,
		txnEnv:     env.TxnEnv,
		prefix:     env.EtcdPrefix,
		repos:      repos,
		commits:    commits,
		branches:   branches,
	}
	storageSrv, err := storage.New(storage.Env{DB: env.DB, ObjectStore: env.ObjectClient}, env.StorageConfig)
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
	if _, err := coredb.GetProjectByName(ctx, txnCtx.SqlTx, repo.Project.Name); err != nil {
		return errors.Wrapf(err, "cannot find project %q", repo.Project)
	}

	// Check whether the user has the permission to create a repo in this project.
	if err := d.env.AuthServer.CheckProjectIsAuthorizedInTransaction(txnCtx, repo.Project, auth.Permission_PROJECT_CREATE_REPO); err != nil {
		return errors.Wrap(err, "failed to create repo")
	}

	repos := d.repos.ReadWrite(txnCtx.SqlTx)
	// Check if 'repo' already exists. If so, return that error.  Otherwise,
	// proceed with ACL creation (avoids awkward "access denied" error when
	// calling "createRepo" on a repo that already exists).
	var existingRepoInfo pfs.RepoInfo
	if err := repos.Get(repo, &existingRepoInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "error checking whether repo %q exists", repo)
		}
		// if this is a system repo, make sure the corresponding user repo already exists
		if repo.Type != pfs.UserRepoType {
			baseRepo := &pfs.Repo{Project: &pfs.Project{Name: repo.Project.GetName()}, Name: repo.GetName(), Type: pfs.UserRepoType}
			if err := repos.Get(baseRepo, &existingRepoInfo); err != nil {
				if !col.IsErrNotFound(err) {
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
			if err := d.env.AuthServer.CreateRoleBindingInTransaction(txnCtx, whoAmI.Username, []string{auth.RepoOwnerRole}, repo.AuthResource()); err != nil && (!col.IsErrExists(err) || repo.Type == pfs.UserRepoType) {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not create role binding for new repo %q", repo)
			}
		}
		return errors.EnsureStack(repos.Create(repo, &pfs.RepoInfo{
			Repo:        repo,
			Created:     txnCtx.Timestamp,
			Description: description,
		}))
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
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.Wrapf(err, "could not update description of %q", repo)
	}
	existingRepoInfo.Description = description
	return errors.EnsureStack(repos.Put(repo, &existingRepoInfo))
}

func (d *driver) inspectRepo(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
	// Validate arguments
	if repo == nil {
		return nil, errors.New("repo cannot be nil")
	}
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(repo, repoInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, pfsserver.ErrRepoNotFound{Repo: repo}
		}
		return nil, errors.EnsureStack(err)
	}
	if includeAuth {
		resp, err := d.env.AuthServer.GetPermissionsInTransaction(txnCtx, &auth.GetPermissionsRequest{
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

func (d *driver) getPermissions(ctx context.Context, repo *pfs.Repo) ([]auth.Permission, []string, error) {
	resp, err := d.env.AuthServer.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: repo.AuthResource()})
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}

	return resp.Permissions, resp.Roles, nil
}

func (d *driver) getPermissionsInTransaction(txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]auth.Permission, []string, error) {
	resp, err := d.env.AuthServer.GetPermissionsInTransaction(txnCtx, &auth.GetPermissionsRequest{Resource: repo.AuthResource()})
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}

	return resp.Permissions, resp.Roles, nil
}

// NOTE if includeAuth = true, it is expected that auth is active
func (d *driver) processListRepoInfo(
	repoInfo *pfs.RepoInfo,
	includeAuth bool,
	projects []*pfs.Project,
	checkAccess func(*pfs.Repo) error,
	collectRepoPerms func(*pfs.Repo) ([]auth.Permission, []string, error),
	repoSize func(*pfs.Repo) (int64, error),
	cb func(*pfs.RepoInfo) error,
) func(string) error {
	// Helper func to filter out repos based on projects.
	projectsFilter := make(map[string]bool)
	for _, project := range projects {
		projectsFilter[project.GetName()] = true
	}
	keep := func(repo *pfs.Repo) bool {
		// Assume the user meant all projects by not providing any projects to filter on.
		if len(projectsFilter) == 0 {
			return true
		}
		return projectsFilter[repo.Project.GetName()]
	}
	processFunc := func(string) error {
		if !keep(repoInfo.Repo) {
			return nil
		}
		if includeAuth {
			if err := checkAccess(repoInfo.Repo); err != nil {
				if !errors.As(err, &auth.ErrNotAuthorized{}) {
					return errors.Wrapf(err, "could not check user is authorized to list repo, problem with repo %s", repoInfo.Repo)
				}
				return nil
			}
			permissions, roles, err := collectRepoPerms(repoInfo.Repo)
			if err != nil {
				return errors.Wrapf(err, "error getting access level for %q", repoInfo.Repo)
			}
			repoInfo.AuthInfo = &pfs.AuthInfo{Permissions: permissions, Roles: roles}
		}
		size, err := repoSize(repoInfo.Repo)
		if err != nil {
			return err
		}
		repoInfo.SizeBytesUpperBound = size
		return cb(proto.Clone(repoInfo).(*pfs.RepoInfo))
	}
	return processFunc
}

func (d *driver) listRepoInTransaction(txnCtx *txncontext.TransactionContext, includeAuth bool, repoType string, projects []*pfs.Project, cb func(*pfs.RepoInfo) error) error {
	// Helper func to check whether a user is allowed to see the given repo in the result.
	// Cache the project level access because it applies to every repo within the same project.
	checkProjectAccess := miscutil.CacheFunc(func(project string) error {
		return d.env.AuthServer.CheckProjectIsAuthorizedInTransaction(txnCtx, &pfs.Project{Name: project}, auth.Permission_PROJECT_LIST_REPO)
	}, 100 /* size */)
	checkAccess := func(repo *pfs.Repo) error {
		if err := checkProjectAccess(repo.Project.GetName()); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return err
			}
			// Allow access if user has the right permissions at the individual Repo-level.
			return d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_READ)
		}
		return nil
	}
	collectPerms := func(r *pfs.Repo) ([]auth.Permission, []string, error) {
		return d.getPermissionsInTransaction(txnCtx, r)
	}
	var authActive bool
	if includeAuth {
		if _, err := txnCtx.WhoAmI(); err == nil {
			authActive = true
		} else if !errors.Is(err, auth.ErrNotActivated) {
			return errors.Wrap(err, "check if auth is active")
		}
	}
	repoSize := func(r *pfs.Repo) (int64, error) {
		return d.repoSize(txnCtx, r)
	}
	ri := &pfs.RepoInfo{}
	processFunc := d.processListRepoInfo(ri, includeAuth && authActive, projects, checkAccess, collectPerms, repoSize, cb)
	if repoType == "" {
		// blank type means return all
		return errors.Wrap(
			d.repos.ReadWrite(txnCtx.SqlTx).List(ri, col.DefaultOptions(), processFunc),
			"could not list repos of all types",
		)
	}
	return errors.Wrapf(
		d.repos.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.ReposTypeIndex, repoType, ri, col.DefaultOptions(), processFunc),
		"could not get repos of type %q: ERROR FROM GetByIndex", repoType,
	)

}

func (d *driver) listRepo(ctx context.Context, includeAuth bool, repoType string, projects []*pfs.Project, cb func(*pfs.RepoInfo) error) error {
	// Helper func to check whether a user is allowed to see the given repo in the result.
	// Cache the project level access because it applies to every repo within the same project.
	checkProjectAccess := miscutil.CacheFunc(func(project string) error {
		return d.env.AuthServer.CheckProjectIsAuthorized(ctx, &pfs.Project{Name: project}, auth.Permission_PROJECT_LIST_REPO)
	}, 100 /* size */)
	checkAccess := func(repo *pfs.Repo) error {
		if err := checkProjectAccess(repo.Project.GetName()); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return err
			}
			// Allow access if user has the right permissions at the individual Repo-level.
			return d.env.AuthServer.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_READ)
		}
		return nil
	}
	collectPerms := func(r *pfs.Repo) ([]auth.Permission, []string, error) {
		return d.getPermissions(ctx, r)
	}
	var authActive bool
	if includeAuth {
		var err error
		_, err = d.env.AuthServer.WhoAmI(ctx, &auth.WhoAmIRequest{})
		if err == nil {
			authActive = true
		} else if !errors.Is(err, auth.ErrNotActivated) {
			return errors.Wrap(err, "check if auth is active")
		}
	}
	repoSize := func(r *pfs.Repo) (int64, error) {
		var size int64
		if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			var err error
			size, err = d.repoSize(txnCtx, r)
			return err
		}); err != nil {
			return 0, err
		}
		return size, nil
	}
	ri := &pfs.RepoInfo{}
	processFunc := d.processListRepoInfo(ri, includeAuth && authActive, projects, checkAccess, collectPerms, repoSize, cb)
	if repoType == "" {
		// blank type means return all
		return errors.Wrap(
			d.repos.ReadOnly(ctx).List(ri, col.DefaultOptions(), processFunc),
			"could not list repos of all types",
		)
	}
	return errors.Wrapf(
		d.repos.ReadOnly(ctx).GetByIndex(pfsdb.ReposTypeIndex, repoType, ri, col.DefaultOptions(), processFunc),
		"could not get repos of type %q: ERROR FROM GetByIndex", repoType,
	)
}

func (d *driver) deleteRepos(ctx context.Context, projects []*pfs.Project) ([]*pfs.Repo, error) {
	var deleted []*pfs.Repo
	if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		deleted, err = d.deleteReposInTransaction(txnCtx, projects)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "delete repos")
	}
	return deleted, nil
}

func (d *driver) deleteReposInTransaction(txnCtx *txncontext.TransactionContext, projects []*pfs.Project) ([]*pfs.Repo, error) {
	var ris []*pfs.RepoInfo
	if err := d.listRepoInTransaction(txnCtx, false, "", projects, func(ri *pfs.RepoInfo) error {
		ris = append(ris, ri)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	return d.deleteReposHelper(txnCtx, ris, false)
}

func (d *driver) deleteReposHelper(txnCtx *txncontext.TransactionContext, ris []*pfs.RepoInfo, force bool) ([]*pfs.Repo, error) {
	if len(ris) == 0 {
		return nil, nil
	}
	// filter out repos that cannot be deleted
	var filter []*pfs.RepoInfo
	for _, ri := range ris {
		ok, err := d.canDeleteRepo(txnCtx, ri.Repo)
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
		bs, err := d.listRepoBranches(txnCtx, ri)
		if err != nil {
			return nil, err
		}
		bis = append(bis, bs...)
	}
	if err := d.deleteBranches(txnCtx, bis, force); err != nil {
		return nil, err
	}
	for _, ri := range ris {
		if err := d.deleteRepoInfo(txnCtx, ri); err != nil {
			return nil, err
		}
		deleted = append(deleted, ri.Repo)
	}
	return deleted, nil
}

func (d *driver) deleteRepo(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, force bool) error {
	repos := d.repos.ReadWrite(txnCtx.SqlTx)
	repoInfo := &pfs.RepoInfo{}
	if err := repos.Get(repo, repoInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "error checking whether %q exists", repo)
		}
		return nil
	}
	if ok, err := d.canDeleteRepo(txnCtx, repo); err != nil {
		return errors.Wrapf(err, "error checking whether repo %q can be deleted", repo.String())
	} else if !ok {
		return nil
	}
	related, err := d.relatedRepos(txnCtx, repo)
	if err != nil {
		return err
	}
	var bis []*pfs.BranchInfo
	for _, ri := range related {
		bs, err := d.listRepoBranches(txnCtx, ri)
		if err != nil {
			return err
		}
		bis = append(bis, bs...)
	}
	if err := d.deleteBranches(txnCtx, bis, force); err != nil {
		return err
	}
	for _, ri := range related {
		if err := d.deleteRepoInfo(txnCtx, ri); err != nil {
			return err
		}
	}
	return nil
}

// delete branches from most provenance to least, that way if one
// branch is provenant on another (which is likely the case when
// multiple repos are provided) we delete them in the right order.
func (d *driver) deleteBranches(txnCtx *txncontext.TransactionContext, branchInfos []*pfs.BranchInfo, force bool) error {
	sort.Slice(branchInfos, func(i, j int) bool { return len(branchInfos[i].Provenance) > len(branchInfos[j].Provenance) })
	for _, bi := range branchInfos {
		if err := d.deleteBranch(txnCtx, bi.Branch, force); err != nil {
			return errors.Wrapf(err, "delete branch %s", bi.Branch)
		}
	}
	return nil
}

func (d *driver) listRepoBranches(txnCtx *txncontext.TransactionContext, repo *pfs.RepoInfo) ([]*pfs.BranchInfo, error) {
	var bis []*pfs.BranchInfo
	for _, branch := range repo.Branches {
		bi, err := d.inspectBranchInTransaction(txnCtx, branch)
		if err != nil {
			return nil, errors.Wrapf(err, "error inspecting branch %s", branch)
		}
		bis = append(bis, bi)
	}
	return bis, nil
}

// before this method is called, a caller should make sure this repo can be deleted with d.canDeleteRepo() and that
// all of the repo's branches are deleted using d.deleteBranches()
func (d *driver) deleteRepoInfo(txnCtx *txncontext.TransactionContext, ri *pfs.RepoInfo) error {
	// make a list of all the commits
	commitInfos := make(map[string]*pfs.CommitInfo)
	commitInfo := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(ri.Repo), commitInfo, col.DefaultOptions(), func(string) error {
		commitInfos[commitInfo.Commit.Id] = proto.Clone(commitInfo).(*pfs.CommitInfo)
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	// and then delete them
	for _, ci := range commitInfos {
		if err := d.commitStore.DropFileSetsTx(txnCtx.SqlTx, ci.Commit); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Despite the fact that we already deleted each branch with
	// deleteBranch, we also do branches.DeleteAll(), this insulates us
	// against certain corruption situations where the RepoInfo doesn't
	// exist in postgres but branches do.
	if err := d.branches.ReadWrite(txnCtx.SqlTx).DeleteByIndex(pfsdb.BranchesRepoIndex, pfsdb.RepoKey(ri.Repo)); err != nil {
		return errors.EnsureStack(err)
	}
	// Similarly with commits
	if err := d.commits.ReadWrite(txnCtx.SqlTx).DeleteByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(ri.Repo)); err != nil {
		return errors.EnsureStack(err)
	}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Delete(ri.Repo); err != nil && !col.IsErrNotFound(err) {
		return errors.Wrapf(err, "repos.Delete")
	}
	// since system repos share a role binding, only delete it if this is the user repo, in which case the other repos will be deleted anyway
	if ri.Repo.Type == pfs.UserRepoType {
		if err := d.env.AuthServer.DeleteRoleBindingInTransaction(txnCtx, ri.Repo.AuthResource()); err != nil && !auth.IsErrNotActivated(err) {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

func (d *driver) relatedRepos(txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]*pfs.RepoInfo, error) {
	repos := d.repos.ReadWrite(txnCtx.SqlTx)
	if repo.Type != pfs.UserRepoType {
		ri := &pfs.RepoInfo{}
		if err := repos.Get(repo, ri); err != nil {
			return nil, err
		}
		return []*pfs.RepoInfo{ri}, nil
	}
	var related []*pfs.RepoInfo
	otherRepo := &pfs.RepoInfo{}
	if err := repos.GetByIndex(pfsdb.ReposNameIndex, pfsdb.ReposNameKey(repo), otherRepo, col.DefaultOptions(), func(key string) error {
		related = append(related, proto.Clone(otherRepo).(*pfs.RepoInfo))
		return nil
	}); err != nil && !col.IsErrNotFound(err) { // TODO(acohen4): !NotFound may be unnecessary
		return nil, errors.Wrapf(err, "error finding dependent repos for %q", repo.Name)
	}
	return related, nil
}

func (d *driver) canDeleteRepo(txnCtx *txncontext.TransactionContext, repo *pfs.Repo) (bool, error) {
	userRepo := proto.Clone(repo).(*pfs.Repo)
	userRepo.Type = pfs.UserRepoType
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, userRepo, auth.Permission_REPO_DELETE); err != nil {
		if auth.IsErrNotAuthorized(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "check repo %q is authorized for deletion", userRepo.String())
	}
	if _, err := d.env.GetPPSServer().InspectPipelineInTransaction(txnCtx, pps.RepoPipeline(repo)); err == nil {
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
		return errors.Wrapf(coredb.UpsertProject(ctx, txnCtx.SqlTx,
			&pfs.ProjectInfo{Project: req.Project, Description: req.Description}),
			"failed to update project %s", req.Project.Name)
	}
	// If auth is active, make caller the owner of this new project.
	if whoAmI, err := txnCtx.WhoAmI(); err == nil {
		if err := d.env.AuthServer.CreateRoleBindingInTransaction(
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
	if err := coredb.CreateProject(ctx, txnCtx.SqlTx, &pfs.ProjectInfo{
		Project:     req.Project,
		Description: req.Description,
		CreatedAt:   timestamppb.Now(),
	}); err != nil {
		if errors.As(err, &col.ErrExists{}) {
			return pfsserver.ErrProjectExists{
				Project: req.Project,
			}
		}
		return errors.Wrap(err, "could not create project")
	}
	return nil
}

func (d *driver) inspectProject(ctx context.Context, project *pfs.Project) (*pfs.ProjectInfo, error) {
	var pi *pfs.ProjectInfo
	if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		pi, err = coredb.GetProjectByName(ctx, tx, pfsdb.ProjectKey(project))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "error getting project %q", project)
	}
	resp, err := d.env.AuthServer.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: project.AuthResource()})
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
		}, zap.String("commit", commit.Id), zap.String("branch", commit.GetBranch().String()), zap.String("repo", commit.Repo.String()), zap.String("target", request.FilePath)); err != nil {
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
					resp, err := d.env.AuthServer.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: proj.GetProject().AuthResource()})
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
	projIter, err := coredb.ListProject(ctx, txnCtx.SqlTx)
	if err != nil {
		return errors.Wrap(err, "could not list project")
	}
	return errors.Wrap(stream.ForEach[*pfs.ProjectInfo](ctx, projIter, cb), "list projects")
}

// TODO: delete all repos and pipelines within project
func (d *driver) deleteProject(ctx context.Context, txnCtx *txncontext.TransactionContext, project *pfs.Project, force bool) error {
	if err := project.ValidateName(); err != nil {
		return errors.Wrap(err, "invalid project name")
	}
	if err := d.env.AuthServer.CheckProjectIsAuthorizedInTransaction(txnCtx, project, auth.Permission_PROJECT_DELETE, auth.Permission_PROJECT_MODIFY_BINDINGS); err != nil {
		return errors.Wrapf(err, "user is not authorized to delete project %q", project)
	}
	var errs error
	if err := d.listRepoInTransaction(txnCtx, false /* includeAuth */, "" /* repoType */, []*pfs.Project{project} /* projectsFilter */, func(repoInfo *pfs.RepoInfo) error {
		errs = errors.Join(errs, fmt.Errorf("repo %v still exists", repoInfo.GetRepo()))
		return nil
	}); err != nil {
		return err
	}
	if errs != nil {
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("cannot delete project %s: %v", project.Name, errs))
	}
	if err := coredb.DeleteProject(ctx, txnCtx.SqlTx, pfsdb.ProjectKey(project)); err != nil {
		return errors.Wrapf(err, "delete project %q", project)
	}
	if err := d.env.AuthServer.DeleteRoleBindingInTransaction(txnCtx, project.AuthResource()); err != nil {
		if !errors.Is(err, auth.ErrNotActivated) {
			return errors.Wrapf(err, "delete role binding for project %q", project)
		}
	}
	return nil
}

// Set child.ParentCommit (if 'parent' has been determined) and write child to parent's ChildCommits
func (d *driver) linkParent(txnCtx *txncontext.TransactionContext, child *pfs.CommitInfo, parent *pfs.Commit, needsFinishedParent bool) error {
	if parent == nil {
		return nil
	}
	// Resolve 'parent' if it's a branch that isn't 'branch' (which can
	// happen if 'branch' is new and diverges from the existing branch in
	// 'parent').
	parentCommitInfo, err := d.resolveCommit(txnCtx.SqlTx, parent)
	if err != nil {
		return errors.Wrapf(err, "parent commit not found")
	}
	// fail if the parent commit has not been finished
	if needsFinishedParent && parentCommitInfo.Finishing == nil {
		return errors.Errorf("parent commit %s has not been finished", parentCommitInfo.Commit)
	}
	child.ParentCommit = parentCommitInfo.Commit
	parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, child.Commit)
	return errors.Wrapf(d.commits.ReadWrite(txnCtx.SqlTx).Put(parentCommitInfo.Commit, parentCommitInfo),
		"could not resolve parent commit %s", parent)
}

// creates a new commit, and adds both commit ancestry, and commit provenance pointers
//
// NOTE: Requiring source commits to have finishing / finished parents ensures that the commits are not compacted
// in a pathological order (finishing later commits before earlier commits will result with us compacting
// the earlier commits multiple times).
func (d *driver) addCommit(txnCtx *txncontext.TransactionContext, newCommitInfo *pfs.CommitInfo, parent *pfs.Commit, directProvenance []*pfs.Branch, needsFinishedParent bool) error {
	if err := d.linkParent(txnCtx, newCommitInfo, parent, needsFinishedParent); err != nil {
		return err
	}
	for _, prov := range directProvenance {
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(pfsdb.BranchKey(prov), branchInfo); err != nil {
			if col.IsErrNotFound(err) {
				return pfsserver.ErrBranchNotFound{Branch: prov}
			}
			return errors.EnsureStack(err)
		}
		newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, branchInfo.Head)
	}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(newCommitInfo.Commit, newCommitInfo); err != nil {
		if col.IsErrExists(err) {
			return errors.EnsureStack(pfsserver.ErrInconsistentCommit{Commit: newCommitInfo.Commit})
		}
		return errors.EnsureStack(err)
	}
	for _, p := range newCommitInfo.DirectProvenance {
		if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommitInfo.Commit, p); err != nil {
			return err
		}
	}
	return nil
}

// startCommit makes a new commit in 'branch', with the parent 'parent':
//   - 'parent' may be omitted, in which case the parent commit is inferred
//     from 'branch'.
//   - If 'parent' is set, it determines the parent commit, but 'branch' is
//     still moved to point at the new commit
func (d *driver) startCommit(
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
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_WRITE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// New commit and commitInfo
	newCommitInfo := newUserCommitInfo(txnCtx, branch)
	newCommitInfo.Description = description
	if err := ancestry.ValidateName(branch.Name); err != nil {
		return nil, err
	}
	// Check if repo exists and load it in case we need to add a new branch
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(branch.Repo, repoInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, pfsserver.ErrRepoNotFound{Repo: branch.Repo}
		}
		return nil, errors.EnsureStack(err)
	}
	// update 'branch' (which must always be set) and set parent.ID (if 'parent'
	// was not set)
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Upsert(branch, branchInfo, func() error {
		if branchInfo.Branch == nil {
			// New branch, update the RepoInfo
			add(&repoInfo.Branches, branch)
			if err := d.repos.ReadWrite(txnCtx.SqlTx).Put(repoInfo.Repo, repoInfo); err != nil {
				return errors.EnsureStack(err)
			}
			branchInfo.Branch = branch
		}
		// If the parent is unspecified, use the current head of the branch
		if parent == nil {
			parent = branchInfo.Head
		}
		// Point 'branch' at the new commit
		branchInfo.Head = newCommitInfo.Commit
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// check if this is happening in a spout pipeline, and alias the spec commit
	_, ok1 := os.LookupEnv(client.PPSPipelineNameEnv)
	_, ok2 := os.LookupEnv("PPS_SPEC_COMMIT")
	if !(ok1 && ok2) && len(branchInfo.Provenance) > 0 {
		// Otherwise, we don't allow user code to start commits on output branches
		return nil, pfsserver.ErrCommitOnOutputBranch{Branch: branch}
	}
	if err := d.addCommit(txnCtx, newCommitInfo, parent, branchInfo.DirectProvenance, true); err != nil {
		return nil, err
	}
	// Defer propagation of the commit until the end of the transaction so we can
	// batch downstream commits together if there are multiple changes.
	if err := txnCtx.PropagateBranch(branch); err != nil {
		return nil, err
	}
	return newCommitInfo.Commit, nil
}

func (d *driver) finishCommit(txnCtx *txncontext.TransactionContext, commit *pfs.Commit, description, commitError string, force bool) error {
	commitInfo, err := d.resolveCommit(txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	if commitInfo.Finishing != nil {
		return pfsserver.ErrCommitFinished{
			Commit: commitInfo.Commit,
		}
	}
	if !force && len(commitInfo.DirectProvenance) > 0 {
		if info, err := d.env.GetPPSServer().InspectPipelineInTransaction(txnCtx, pps.RepoPipeline(commitInfo.Commit.Repo)); err != nil && !errutil.IsNotFoundError(err) {
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
	return errors.EnsureStack(d.commits.ReadWrite(txnCtx.SqlTx).Put(commitInfo.Commit, commitInfo))
}

func (d *driver) repoSize(txnCtx *txncontext.TransactionContext, repo *pfs.Repo) (int64, error) {
	repoInfo := new(pfs.RepoInfo)
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(repo, repoInfo); err != nil {
		return 0, errors.EnsureStack(err)
	}
	for _, branch := range repoInfo.Branches {
		if branch.Name == "master" {
			branchInfo := &pfs.BranchInfo{}
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(branch, branchInfo); err != nil {
				return 0, errors.EnsureStack(err)
			}
			commit := branchInfo.Head
			for commit != nil {
				commitInfo, err := d.resolveCommit(txnCtx.SqlTx, commit)
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
func (d *driver) propagateBranches(txnCtx *txncontext.TransactionContext, branches []*pfs.Branch) error {
	if len(branches) == 0 {
		return nil
	}
	var propagatedBranches []*pfs.BranchInfo
	seen := make(map[string]*pfs.BranchInfo)
	bi := &pfs.BranchInfo{}
	for _, b := range branches {
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(b, bi); err != nil {
			return errors.EnsureStack(err)
		}
		for _, b := range bi.Subvenance {
			if _, ok := seen[pfsdb.BranchKey(b)]; !ok {
				bi := &pfs.BranchInfo{}
				if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(b, bi); err != nil {
					return errors.Wrapf(err, "get subvenant branch %q", pfsdb.BranchKey(b))
				}
				seen[pfsdb.BranchKey(b)] = bi
				propagatedBranches = append(propagatedBranches, bi)
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
				provBranchInfo := &pfs.BranchInfo{}
				if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(pfsdb.BranchKey(b), provBranchInfo); err != nil {
					return errors.Wrapf(err, "get provenant branch %q", pfsdb.BranchKey(b))
				}
				provCommit = provBranchInfo.Head
			}
			newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, provCommit)
		}
		// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
		newCommitInfo.ParentCommit = proto.Clone(bi.Head).(*pfs.Commit)
		bi.Head = newCommit
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(bi.Branch, bi); err != nil {
			return errors.Wrapf(err, "put branch %q with head %q", pfsdb.BranchKey(bi.Branch), pfsdb.CommitKey(bi.Head))
		}
		// create open 'commit'.
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(newCommit, newCommitInfo); err != nil {
			return errors.Wrapf(err, "create new commit %q", pfsdb.CommitKey(newCommit))
		}
		if newCommitInfo.ParentCommit != nil {
			parentCommitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(newCommitInfo.ParentCommit, parentCommitInfo, func() error {
				parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
				return nil
			}); err != nil {
				return errors.Wrapf(err, "update parent commit %q with child %q", pfsdb.CommitKey(newCommitInfo.ParentCommit), pfsdb.CommitKey(newCommit))
			}
		}
		// add commit provenance
		for _, c := range newCommitInfo.DirectProvenance {
			if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommit, c); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q", pfsdb.CommitKey(newCommit), pfsdb.CommitKey(c))
			}
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
			if err := d.commits.ReadOnly(ctx).WatchOneF(commitInfo.Commit, func(ev *watch.Event) error {
				if ev.Type == watch.EventDelete {
					return pfsserver.ErrCommitDeleted{Commit: commitInfo.Commit}
				}
				var key string
				newCommitInfo := &pfs.CommitInfo{}
				if err := ev.Unmarshal(&key, newCommitInfo); err != nil {
					return errors.Wrapf(err, "unmarshal")
				}
				switch wait {
				case pfs.CommitState_FINISHING:
					if newCommitInfo.Finishing != nil {
						commitInfo = newCommitInfo
						return errutil.ErrBreak
					}
				case pfs.CommitState_FINISHED:
					if newCommitInfo.Finished != nil {
						commitInfo = newCommitInfo
						return errutil.ErrBreak
					}
				}
				return nil
			}); err != nil {
				return nil, errors.EnsureStack(err)
			}
		}
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
	if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, commit.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// Resolve the commit in case it specifies a branch head or commit ancestry
	var commitInfo *pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfo, err = d.resolveCommit(txnCtx.SqlTx, commit)
		return err
	}); err != nil {
		return nil, err
	}
	return commitInfo, nil
}

// resolveCommit contains the essential implementation of inspectCommit: it converts 'commit' (which may
// be a commit ID or branch reference, plus '~' and/or '^') to a repo + commit
// ID. It accepts a postgres transaction so that it can be used in a transaction
// and avoids an inconsistent call to d.inspectCommit()
func (d *driver) resolveCommit(sqlTx *pachsql.Tx, userCommit *pfs.Commit) (*pfs.CommitInfo, error) {
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
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(sqlTx).Get(commit.Branch, branchInfo); err != nil {
			return nil, errors.EnsureStack(err)
		}
		commit.Id = branchInfo.Head.Id
	}
	commitInfo := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(sqlTx).Get(commit, commitInfo); err != nil {
		if col.IsErrNotFound(err) {
			// try to resolve to alias if not found
			resolvedCommit, err := pfsdb.ResolveCommitProvenance(sqlTx, userCommit.Repo, commit.Id)
			if err != nil {
				return nil, err
			}
			commit.Id = resolvedCommit.Id
			if err := d.commits.ReadWrite(sqlTx).Get(commit, commitInfo); err != nil {
				return nil, errors.EnsureStack(err)
			}
		} else {
			return nil, errors.EnsureStack(err)
		}
	}
	// Traverse commits' parents until you've reached the right ancestor
	if ancestryLength >= 0 {
		for i := 1; i <= ancestryLength; i++ {
			if commitInfo.ParentCommit == nil {
				return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
			}
			if err := d.commits.ReadWrite(sqlTx).Get(commitInfo.ParentCommit, commitInfo); err != nil {
				if col.IsErrNotFound(err) {
					if i == 0 {
						return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
					}
					return nil, pfsserver.ErrParentCommitNotFound{Commit: commit}
				}
				return nil, errors.EnsureStack(err)
			}
		}
	} else {
		cis := make([]pfs.CommitInfo, ancestryLength*-1)
		for i := 0; ; i++ {
			if commit == nil {
				if i >= len(cis) {
					commitInfo = &cis[i%len(cis)]
					break
				}
				return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
			}
			if err := d.commits.ReadWrite(sqlTx).Get(commit, &cis[i%len(cis)]); err != nil {
				if col.IsErrNotFound(err) {
					if i == 0 {
						return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
					}
					return nil, pfsserver.ErrParentCommitNotFound{Commit: commit}
				}
				return nil, err
			}
			commit = cis[i%len(cis)].ParentCommit
		}
	}
	return commitInfo, nil
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
		commitInfo, err = d.resolveCommit(txnCtx.SqlTx, commit)
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

	if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_LIST_COMMIT); err != nil {
		return errors.EnsureStack(err)
	}
	if from != nil && !proto.Equal(from.Repo, repo) || to != nil && !proto.Equal(to.Repo, repo) {
		return errors.Errorf("`from` and `to` commits need to be from repo %s", repo)
	}
	// Make sure that the repo exists
	if repo.Name != "" {
		if err := d.repos.ReadOnly(ctx).Get(repo, &pfs.RepoInfo{}); err != nil {
			if col.IsErrNotFound(err) {
				return pfsserver.ErrRepoNotFound{Repo: repo}
			}
			return errors.EnsureStack(err)
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
		ci := &pfs.CommitInfo{}
		lastRev := int64(-1)
		listCallback := func(key string, createRev int64) error {
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
					createdAt := time.Unix(int64(ci.Started.GetSeconds()), int64(ci.Started.GetNanos())).UTC()
					fromTime := time.Unix(int64(startTime.GetSeconds()), int64(startTime.GetNanos())).UTC()
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
		// the repo, sorted by revision timestamp (or reversed if so requested.)
		opts := &col.Options{Target: col.SortByCreateRevision, Order: col.SortDescend}
		if reverse {
			opts.Order = col.SortAscend
		}

		if repo.Name == "" {
			if err := d.commits.ReadOnly(ctx).ListRev(ci, opts, listCallback); err != nil {
				return errors.EnsureStack(err)
			}
		} else {
			if err := d.commits.ReadOnly(ctx).GetRevByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repo), ci, opts, listCallback); err != nil {
				return errors.EnsureStack(err)
			}
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
			commitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadOnly(ctx).Get(cursor, commitInfo); err != nil {
				return errors.EnsureStack(err)
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

	// Note that this watch may leave events unread for a long amount of time
	// while waiting for the commit state - if the watch channel fills up, it will
	// error out.
	err := d.commits.ReadOnly(ctx).WatchByIndexF(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repo), func(ev *watch.Event) error {
		var key string
		commitInfo := &pfs.CommitInfo{}
		if err := ev.Unmarshal(&key, commitInfo); err != nil {
			return errors.Wrapf(err, "unmarshal")
		}

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
	}, watch.WithSort(col.SortByCreateRevision, col.SortAscend), watch.IgnoreDelete)
	return errors.EnsureStack(err)
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

// TODO(provenance): consider removing this functionality
func (d *driver) fillNewBranches(txnCtx *txncontext.TransactionContext, branch *pfs.Branch, provenance []*pfs.Branch) error {
	repoBranches := map[*pfs.Repo][]*pfs.Branch{branch.Repo: {branch}}
	newRepoCommits := make(map[string]*pfs.Commit)
	for _, p := range provenance {
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Upsert(p, branchInfo, func() error {
			if branchInfo.Branch == nil {
				branchInfo.Branch = p
				var head *pfs.Commit
				var ok bool
				if head, ok = newRepoCommits[pfsdb.RepoKey(branchInfo.Branch.Repo)]; !ok {
					var err error
					head, err = d.makeEmptyCommit(txnCtx, branchInfo.Branch, nil, nil)
					if err != nil {
						return err
					}
					newRepoCommits[pfsdb.RepoKey(branchInfo.Branch.Repo)] = head
				}
				branchInfo.Head = head
				if branches, ok := repoBranches[p.Repo]; ok {
					add(&branches, p)
				} else {
					repoBranches[p.Repo] = []*pfs.Branch{p}
				}
			}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Add the new branches to their repo infos
	for repo, branches := range repoBranches {
		repoInfo := &pfs.RepoInfo{}
		if err := d.repos.ReadWrite(txnCtx.SqlTx).Update(repo, repoInfo, func() error {
			for _, b := range branches {
				add(&repoInfo.Branches, b)
			}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// Given branchInfo.DirectProvenance and its oldProvenance, compute Provenance & Subvenance of branchInfo.
// Also update the Subvenance of branchInfo's old and new provenant branches.
//
// This algorithm updates every branch in branchInfo's Subvenance, new Provenance, and old Provenance.
// Complexity is O(m*n*log(n)) where m is the complete Provenance of branchInfo, and n is the Subvenance of branchInfo
func (d *driver) computeBranchProvenance(txnCtx *txncontext.TransactionContext, branchInfo *pfs.BranchInfo, oldDirectProvenance []*pfs.Branch) error {
	for _, p := range branchInfo.DirectProvenance {
		if has(&branchInfo.Subvenance, p) {
			return errors.Errorf("branch %q cannot be both in %q's provenance and subvenance", p.String(), branchInfo.Branch.String())
		}
	}
	branchInfoCache := map[string]*pfs.BranchInfo{pfsdb.BranchKey(branchInfo.Branch): branchInfo}
	getBranchInfo := func(b *pfs.Branch) (*pfs.BranchInfo, error) {
		if bi, ok := branchInfoCache[pfsdb.BranchKey(b)]; ok {
			return bi, nil
		}
		bi := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(b, bi); err != nil {
			return nil, errors.Wrapf(err, "get branch info")
		}
		branchInfoCache[pfsdb.BranchKey(b)] = bi
		return bi, nil
	}
	toUpdate := []*pfs.BranchInfo{branchInfo}
	for _, sb := range branchInfo.Subvenance {
		sbi, err := getBranchInfo(sb)
		if err != nil {
			return err
		}
		toUpdate = append(toUpdate, sbi)
	}
	// Sorting is important here because it sorts topologically. This means
	// that when evaluating element i of `toUpdate` all elements < i will
	// have already been evaluated and thus we can safely use their
	// Provenance field.
	sort.Slice(toUpdate, func(i, j int) bool { return len(toUpdate[i].Provenance) < len(toUpdate[j].Provenance) })
	// re-compute the complete provenance of branchInfo and all its subvenant branches
	for _, bi := range toUpdate {
		bi.Provenance = make([]*pfs.Branch, 0)
		for _, directProv := range bi.DirectProvenance {
			add(&bi.Provenance, directProv)
			directProvBI, err := getBranchInfo(directProv)
			if err != nil {
				return err
			}
			for _, p := range directProvBI.Provenance {
				add(&bi.Provenance, p)
			}
		}
	}
	// add branchInfo and its subvenance to the subvenance of all of its provenance branches
	for _, p := range branchInfo.Provenance {
		pbi, err := getBranchInfo(p)
		if err != nil {
			return err
		}
		for _, ubi := range toUpdate {
			add(&pbi.Subvenance, ubi.Branch)
		}
	}
	// remove branchInfo and its subvenance from all branches that are no longer in branchInfo's Provenance
	oldProvenance := make([]*pfs.Branch, 0)
	for _, odp := range oldDirectProvenance {
		if !has(&branchInfo.Provenance, odp) {
			add(&oldProvenance, odp)
			oldProvenance = append(oldProvenance, odp)
			opbi, err := getBranchInfo(odp)
			if err != nil {
				return err
			}
			for _, b := range opbi.Provenance {
				add(&oldProvenance, b)
			}
		}
	}
	for _, op := range oldProvenance {
		opbi, err := getBranchInfo(op)
		if err != nil {
			return err
		}
		for _, bi := range toUpdate {
			if !has(&bi.Provenance, op) {
				del(&opbi.Subvenance, bi.Branch)
			}
		}
	}
	// now that all Provenance + Subvenance fields are up to date, save all the branches
	for _, updateBi := range branchInfoCache {
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(updateBi.Branch, updateBi); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// for a DAG to be valid, it may not have a multiple branches from the same repo
// reachable by traveling edges bidirectionally. The reason is that this would complicate resolving
func (d *driver) validateDAGStructure(txnCtx *txncontext.TransactionContext, bs []*pfs.Branch) error {
	cache := make(map[string]*pfs.BranchInfo)
	getBranchInfo := func(b *pfs.Branch) (*pfs.BranchInfo, error) {
		if bi, ok := cache[pfsdb.BranchKey(b)]; ok {
			return bi, nil
		}
		bi := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(b, bi); err != nil {
			return nil, errors.Wrapf(err, "get branch info")
		}
		cache[pfsdb.BranchKey(b)] = bi
		return bi, nil
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
func (d *driver) createBranch(txnCtx *txncontext.TransactionContext, branch *pfs.Branch, commit *pfs.Commit, provenance []*pfs.Branch, trigger *pfs.Trigger) error {
	// Validate arguments
	if branch == nil {
		return errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return errors.New("branch repo cannot be nil")
	}
	if err := d.validateTrigger(txnCtx, branch, trigger); err != nil {
		return err
	}
	if len(provenance) > 0 && trigger != nil {
		return errors.New("a branch cannot have both provenance and a trigger")
	}
	var err error
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_CREATE_BRANCH); err != nil {
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
	if err := d.fillNewBranches(txnCtx, branch, provenance); err != nil {
		return err
	}
	// if the user passed a commit to point this branch at, resolve it
	var ci *pfs.CommitInfo
	if commit != nil {
		ci, err = d.resolveCommit(txnCtx.SqlTx, commit)
		if err != nil {
			return errors.Wrapf(err, "unable to inspect %s", commit)
		}
		commit = ci.Commit
	}
	// retrieve the current version of this branch and set its head if specified
	var oldProvenance []*pfs.Branch
	branchInfo := &pfs.BranchInfo{}
	propagate := false
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Upsert(branch, branchInfo, func() error {
		// check whether direct provenance has changed
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
		// if we don't have a branch head, or the provenance has changed, add a new commit to the branch to capture the changed structure
		// the one edge case here, is that it's undesirable to add a commit in the case where provenance is completely removed...
		//
		// TODO(provenance): This sort of hurts Branch Provenance invariant. See if we can re-assess....
		if branchInfo.Head == nil || (!same(oldProvenance, provenance) && len(provenance) != 0) {
			c, err := d.makeEmptyCommit(txnCtx, branch, provenance, branchInfo.Head)
			if err != nil {
				return err
			}
			branchInfo.Head = c
			propagate = true
		}
		if trigger != nil && trigger.Branch != "" {
			branchInfo.Trigger = trigger
		}
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	// update the total provenance of this branch and all of its subvenant branches.
	// load all branches in the complete closure once and saves all of them.
	if err := d.computeBranchProvenance(txnCtx, branchInfo, oldProvenance); err != nil {
		return err
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
		branchInfo, err = d.inspectBranchInTransaction(txnCtx, branch)
		return err
	}); err != nil {
		return nil, err
	}
	return branchInfo, nil
}

func (d *driver) inspectBranchInTransaction(txnCtx *txncontext.TransactionContext, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	// Validate arguments
	if branch == nil {
		return nil, errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return nil, errors.New("branch repo cannot be nil")
	}

	// Check that the user is logged in, but don't require any access level
	if _, err := txnCtx.WhoAmI(); err != nil {
		if !auth.IsErrNotActivated(err) {
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to inspect a branch)")
		}
	}

	result := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(branch, result); err != nil {
		if col.IsErrNotFound(err) {
			return nil, pfsserver.ErrBranchNotFound{Branch: branch}
		}
		return nil, errors.EnsureStack(err)
	}
	return result, nil
}

func (d *driver) listBranch(ctx context.Context, reverse bool, cb func(*pfs.BranchInfo) error) error {
	if _, err := d.env.AuthServer.WhoAmI(ctx, &auth.WhoAmIRequest{}); err != nil && !auth.IsErrNotActivated(err) {
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
	branchInfo := &pfs.BranchInfo{}
	listCallback := func(_ string, createRev int64) error {
		if createRev != lastRev {
			if err := sendBis(); err != nil {
				return errors.EnsureStack(err)
			}
			lastRev = createRev
		}
		bis = append(bis, proto.Clone(branchInfo).(*pfs.BranchInfo))
		return nil
	}

	opts := &col.Options{Target: col.SortByCreateRevision, Order: col.SortDescend}
	if reverse {
		opts.Order = col.SortAscend
	}
	if err := d.branches.ReadOnly(ctx).ListRev(branchInfo, opts, listCallback); err != nil {
		return errors.EnsureStack(err)
	}

	return sendBis()
}

func (d *driver) listBranchInTransaction(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, reverse bool, cb func(*pfs.BranchInfo) error) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_LIST_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}

	// Make sure that the repo exists
	if repo.Name != "" {
		if err := d.repos.ReadWrite(txnCtx.SqlTx).Get(pfsdb.RepoKey(repo), &pfs.RepoInfo{}); err != nil {
			if col.IsErrNotFound(err) {
				return pfsserver.ErrRepoNotFound{Repo: repo}
			}
			return errors.EnsureStack(err)
		}
	}

	opts := &col.Options{Target: col.SortByCreateRevision, Order: col.SortDescend}
	if reverse {
		opts.Order = col.SortAscend
	}
	branchInfo := &pfs.BranchInfo{}
	err := d.branches.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.BranchesRepoIndex, pfsdb.RepoKey(repo), branchInfo, opts, func(_ string) error {
		return cb(branchInfo)
	})
	return errors.EnsureStack(err)
}

func (d *driver) deleteBranch(txnCtx *txncontext.TransactionContext, branch *pfs.Branch, force bool) error {
	// Validate arguments
	if branch == nil {
		return errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return errors.New("branch repo cannot be nil")
	}
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo, auth.Permission_REPO_DELETE_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(branch, branchInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "get branch %q", pfsdb.BranchKey(branch))
		}
	}
	if branchInfo.Branch != nil {
		if !force {
			if len(branchInfo.Subvenance) > 0 {
				return errors.Errorf("branch %s has %v as subvenance, deleting it would break those branches", branch.Name, branchInfo.Subvenance)
			}
		}
		// For provenant branches, remove this branch from subvenance
		for _, provBranch := range branchInfo.Provenance {
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(provBranch, provBranchInfo, func() error {
				del(&provBranchInfo.Subvenance, branch)
				return nil
			}); err != nil && !errutil.IsNotFoundError(err) {
				return errors.Wrapf(err, "error deleting subvenance")
			}
		}
		// For subvenant branches, recalculate provenance
		for _, subvBranch := range branchInfo.Subvenance {
			subvBranchInfo := &pfs.BranchInfo{}
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(subvBranch, subvBranchInfo); err != nil {
				return errors.EnsureStack(err)
			}
			del(&subvBranchInfo.DirectProvenance, branch)
			if err := d.createBranch(txnCtx, subvBranch, nil, subvBranchInfo.DirectProvenance, nil); err != nil {
				return err
			}
		}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Delete(branch); err != nil {
			return errors.Wrapf(err, "branches.Delete")
		}
	}
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Update(branch.Repo, repoInfo, func() error {
		del(&repoInfo.Branches, branch)
		return nil
	}); err != nil {
		if !col.IsErrNotFound(err) || !force {
			return errors.EnsureStack(err)
		}
	}
	txnCtx.DeleteBranch(branch)
	return nil
}

func (d *driver) deleteAll(ctx context.Context) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		if _, err := d.deleteReposInTransaction(txnCtx, nil); err != nil {
			return errors.Wrap(err, "could not delete all repos")
		}
		if err := d.listProjectInTransaction(ctx, txnCtx, func(pi *pfs.ProjectInfo) error {
			return errors.Wrapf(d.deleteProject(ctx, txnCtx, pi.Project, true), "delete project %q", pi.Project.String())
		}); err != nil {
			return err
		} // now that the cluster is empty, recreate the default project
		return d.createProjectInTransaction(ctx, txnCtx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "default"}})
	})
}

// only transform source repos and spouts get a closed commit
func (d *driver) makeEmptyCommit(txnCtx *txncontext.TransactionContext, branch *pfs.Branch, directProvenance []*pfs.Branch, parent *pfs.Commit) (*pfs.Commit, error) {
	// Input repos and spouts want a closed head commit, so decide if we leave
	// it open by the presence of branch provenance.  If it's only provenant on
	// a spec repo, we assume it's a spout and close the commit.
	closed := true
	for _, prov := range directProvenance {
		if prov.Repo.Type != pfs.SpecRepoType {
			closed = false
			break
		}
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
		total, err := d.storage.Filesets.ComposeTx(txnCtx.SqlTx, nil, defaultTTL)
		if err != nil {
			return nil, err
		}
		if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commit, *total); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	if err := d.addCommit(txnCtx, commitInfo, parent, directProvenance, false /* needsFinishedParent */); err != nil {
		return nil, err
	}
	return commit, nil
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

func (b *branchSet) del(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if ok {
		copy((*b)[i:], (*b)[i+1:])
		(*b)[len((*b))-1] = nil
		*b = (*b)[:len((*b))-1]
	}
}

func del(bs *[]*pfs.Branch, branch *pfs.Branch) {
	(*branchSet)(bs).del(branch)
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
