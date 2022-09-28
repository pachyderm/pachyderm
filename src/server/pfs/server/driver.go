package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
)

const (
	// Makes calls to ListRepo and InspectRepo more legible
	includeAuth = true
)

const (
	storageTaskNamespace = "storage"
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
	log *logrus.Logger
	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	txnEnv     *txnenv.TransactionEnv
	prefix     string

	// collections
	repos    col.PostgresCollection
	commits  col.PostgresCollection
	branches col.PostgresCollection

	storage     *fileset.Storage
	commitStore commitStore

	cache *fileset.Cache
}

func newDriver(env Env) (*driver, error) {
	storageConfig := env.StorageConfig
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
		log:        env.Logger,
	}
	// Setup tracker and chunk / fileset storage.
	tracker := track.NewPostgresTracker(env.DB)
	chunkStorageOpts, err := chunk.StorageOptions(&storageConfig)
	if err != nil {
		return nil, err
	}
	memCache := storageConfig.ChunkMemoryCache()
	keyStore := chunk.NewPostgresKeyStore(env.DB)
	secret, err := getOrCreateKey(context.TODO(), keyStore, "default")
	if err != nil {
		return nil, err
	}
	chunkStorageOpts = append(chunkStorageOpts, chunk.WithSecret(secret))
	chunkStorage := chunk.NewStorage(objClient, memCache, env.DB, tracker, chunkStorageOpts...)
	d.storage = fileset.NewStorage(fileset.NewPostgresStore(env.DB), tracker, chunkStorage, fileset.StorageOptions(&storageConfig)...)
	// Set up compaction worker.
	taskSource := env.TaskService.NewSource(storageTaskNamespace)
	go compactionWorker(env.BackgroundContext, taskSource, d.storage) //nolint:errcheck
	d.commitStore = newPostgresCommitStore(env.DB, tracker, d.storage)
	// TODO: Make the cache max size configurable.
	d.cache = fileset.NewCache(env.DB, tracker, 10000)
	return d, nil
}

func (d *driver) createRepo(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, description string, update bool) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	// Check that the user is logged in (user doesn't need any access level to
	// create a repo, but they must be authenticated if auth is active)
	whoAmI, err := txnCtx.WhoAmI()
	authIsActivated := !auth.IsErrNotActivated(err)
	if authIsActivated && err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to create a repo)")
	}
	if err := ancestry.ValidateName(repo.Name); err != nil {
		return err
	}

	if repo.Type == "" {
		// default to user type
		repo.Type = pfs.UserRepoType
	}

	repos := d.repos.ReadWrite(txnCtx.SqlTx)

	// check if 'repo' already exists. If so, return that error. Otherwise,
	// proceed with ACL creation (avoids awkward "access denied" error when
	// calling "createRepo" on a repo that already exists)
	var existingRepoInfo pfs.RepoInfo
	err = repos.Get(repo, &existingRepoInfo)
	if err != nil && !col.IsErrNotFound(err) {
		return errors.Wrapf(err, "error checking whether %q exists", repo)
	} else if err == nil {
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
	} else {
		// if this is a system repo, make sure the corresponding user repo already exists
		if repo.Type != pfs.UserRepoType {
			baseRepo := client.NewRepo(repo.Name)
			err = repos.Get(baseRepo, &existingRepoInfo)
			if err != nil && col.IsErrNotFound(err) {
				return errors.Errorf("cannot create a system repo without a corresponding 'user' repo")
			} else if err != nil {
				return errors.Wrapf(err, "error checking whether user repo for %q exists", repo.Name)
			}
		}

		// New repo case
		if authIsActivated {
			// Create ACL for new repo. Make caller the sole owner. If this is a user repo,
			// and the ACL already exists with a different owner, this will fail.
			// For now, we expect system repos to share auth info with their corresponding
			// user repo, so the role binding should exist
			if err := d.env.AuthServer.CreateRoleBindingInTransaction(
				txnCtx,
				whoAmI.Username,
				[]string{auth.RepoOwnerRole},
				&auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name},
			); err != nil && (!col.IsErrExists(err) || repo.Type == pfs.UserRepoType) {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not create role binding for new repo %q", repo)
			}
		}
		return errors.EnsureStack(repos.Create(repo, &pfs.RepoInfo{
			Repo:        repo,
			Created:     txnCtx.Timestamp,
			Description: description,
		}))
	}
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
			Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name},
		})
		if err != nil {
			if auth.IsErrNotActivated(err) {
				return repoInfo, nil
			}
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error getting access level for %q", repo)
		}
		repoInfo.AuthInfo = &pfs.RepoAuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
	}
	return repoInfo, nil
}

func (d *driver) getPermissions(ctx context.Context, repo *pfs.Repo) ([]auth.Permission, []string, error) {
	resp, err := d.env.AuthServer.GetPermissions(ctx, &auth.GetPermissionsRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name},
	})
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}

	return resp.Permissions, resp.Roles, nil
}

func (d *driver) listRepo(ctx context.Context, includeAuth bool, repoType string, cb func(*pfs.RepoInfo) error) error {
	authSeemsActive := true
	repoInfo := &pfs.RepoInfo{}

	processFunc := func(string) error {
		size, err := d.repoSize(ctx, repoInfo.Repo)
		if err != nil {
			return err
		}
		repoInfo.SizeBytesUpperBound = size
		if includeAuth && authSeemsActive {
			permissions, roles, err := d.getPermissions(ctx, repoInfo.Repo)
			if err == nil {
				repoInfo.AuthInfo = &pfs.RepoAuthInfo{Permissions: permissions, Roles: roles}
			} else if auth.IsErrNotActivated(err) {
				authSeemsActive = false
			} else {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error getting access level for %q", repoInfo.Repo)
			}
		}
		return cb(proto.Clone(repoInfo).(*pfs.RepoInfo))
	}

	if repoType == "" {
		// blank type means return all
		return errors.EnsureStack(d.repos.ReadOnly(ctx).List(repoInfo, col.DefaultOptions(), processFunc))
	} else {
		err := d.repos.ReadOnly(ctx).GetByIndex(pfsdb.ReposTypeIndex, repoType, repoInfo, col.DefaultOptions(), processFunc)
		return errors.EnsureStack(err)
	}
}

func (d *driver) deleteAllBranchesFromRepos(txnCtx *txncontext.TransactionContext, repos []pfs.RepoInfo, force bool) error {
	var branchInfos []*pfs.BranchInfo
	for _, repo := range repos {
		for _, branch := range repo.Branches {
			bi, err := d.inspectBranch(txnCtx, branch)
			if err != nil {
				return errors.Wrapf(err, "error inspecting branch %s", branch)
			}
			branchInfos = append(branchInfos, bi)
		}
	}
	// sort ascending provenance
	sort.Slice(branchInfos, func(i, j int) bool { return len(branchInfos[i].Provenance) < len(branchInfos[j].Provenance) })
	for i := range branchInfos {
		// delete branches from most provenance to least, that way if one
		// branch is provenant on another (which is likely the case when
		// multiple repos are provided) we delete them in the right order.
		branch := branchInfos[len(branchInfos)-1-i].Branch
		if err := d.deleteBranch(txnCtx, branch, force); err != nil {
			return errors.Wrapf(err, "delete branch %s", branch)
		}
	}
	return nil
}

func (d *driver) deleteRepo(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, force bool) error {
	repos := d.repos.ReadWrite(txnCtx.SqlTx)

	// check if 'repo' is already gone. If so, return that error. Otherwise,
	// proceed with auth check (avoids awkward "access denied" error when calling
	// "deleteRepo" on a repo that's already gone)
	var repoInfo pfs.RepoInfo
	err := repos.Get(repo, &repoInfo)
	if err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "error checking whether %q exists", repo)
		}
	}

	// Check if the caller is authorized to delete this repo
	if err := d.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_DELETE); err != nil {
		return errors.EnsureStack(err)
	}

	if !force {
		if _, err := d.env.GetPPSServer().InspectPipelineInTransaction(txnCtx, repo.Name); err == nil {
			return errors.Errorf("cannot delete a repo associated with a pipeline - delete the pipeline instead")
		} else if err != nil && !errutil.IsNotFoundError(err) {
			return errors.EnsureStack(err)
		}
	}
	// if this is a user repo, delete any dependent repos
	if repo.Type == pfs.UserRepoType {
		var dependentRepos []pfs.RepoInfo
		var otherRepo pfs.RepoInfo
		if err := repos.GetByIndex(pfsdb.ReposNameIndex, repo.Name, &otherRepo, col.DefaultOptions(), func(key string) error {
			if otherRepo.Repo.Type != repo.Type {
				dependentRepos = append(dependentRepos, otherRepo)
			}
			return nil
		}); err != nil && !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "error finding dependent repos for %q", repo.Name)
		}

		// we expect potentially complicated provenance relationships between dependent repos
		// deleting all branches at once allows for topological sorting, avoiding deletion order issues
		if err := d.deleteAllBranchesFromRepos(txnCtx, append(dependentRepos, repoInfo), force); err != nil {
			return errors.Wrap(err, "error deleting branches")
		}

		// delete the repos we found
		for _, dep := range dependentRepos {
			if err := d.deleteRepo(txnCtx, dep.Repo, force); err != nil {
				return errors.Wrapf(err, "error deleting dependent repo %q", dep.Repo)
			}
		}
	} else {
		if err := d.deleteAllBranchesFromRepos(txnCtx, []pfs.RepoInfo{repoInfo}, force); err != nil {
			return err
		}
	}

	// make a list of all the commits
	commitInfos := make(map[string]*pfs.CommitInfo)
	commitInfo := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repo), commitInfo, col.DefaultOptions(), func(string) error {
		commitInfos[commitInfo.Commit.ID] = proto.Clone(commitInfo).(*pfs.CommitInfo)
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
	if err := d.branches.ReadWrite(txnCtx.SqlTx).DeleteByIndex(pfsdb.BranchesRepoIndex, pfsdb.RepoKey(repo)); err != nil {
		return errors.EnsureStack(err)
	}
	// Similarly with commits
	if err := d.commits.ReadWrite(txnCtx.SqlTx).DeleteByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repo)); err != nil {
		return errors.EnsureStack(err)
	}
	if err := repos.Delete(repo); err != nil && !col.IsErrNotFound(err) {
		return errors.Wrapf(err, "repos.Delete")
	}

	// since system repos share a role binding, only delete it if this is the user repo, in which case the other repos will be deleted anyway
	if repo.Type == pfs.UserRepoType {
		if err := d.env.AuthServer.DeleteRoleBindingInTransaction(txnCtx, &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name}); err != nil && !auth.IsErrNotActivated(err) {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

// startCommit makes a new commit in 'branch', with the parent 'parent':
// - 'parent' may be omitted, in which case the parent commit is inferred
//   from 'branch'.
// - If 'parent' is set, it determines the parent commit, but 'branch' is
//   still moved to point at the new commit
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
	newCommit := &pfs.Commit{
		Branch: branch,
		ID:     txnCtx.CommitSetID,
	}
	newCommitInfo := &pfs.CommitInfo{
		Commit:      newCommit,
		Origin:      &pfs.CommitOrigin{Kind: pfs.OriginKind_USER},
		Description: description,
		Started:     txnCtx.Timestamp,
	}
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
		branchInfo.Head = newCommit
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}

	// Snapshot the branch's direct provenance into the new commit
	newCommitInfo.DirectProvenance = branchInfo.DirectProvenance

	// check if this is happening in a spout pipeline, and alias the spec commit
	spoutName, ok1 := os.LookupEnv(client.PPSPipelineNameEnv)
	spoutCommit, ok2 := os.LookupEnv("PPS_SPEC_COMMIT")
	if ok1 && ok2 {
		specBranch := client.NewSystemRepo(spoutName, pfs.SpecRepoType).NewBranch("master")
		specCommit := specBranch.NewCommit(spoutCommit)
		log.Infof("Adding spout spec commit to current commitset: %s", specCommit)
		if _, err := d.aliasCommit(txnCtx, specCommit, specBranch); err != nil {
			return nil, err
		}
	} else if len(branchInfo.Provenance) > 0 {
		// Otherwise, we don't allow user code to start commits on output branches
		return nil, pfsserver.ErrCommitOnOutputBranch{Branch: branch}
	}

	// Set newCommit.ParentCommit (if 'parent' has been determined) and add
	// newCommit to parent's ChildCommits
	if parent != nil {
		// Resolve 'parent' if it's a branch that isn't 'branch' (which can
		// happen if 'branch' is new and diverges from the existing branch in
		// 'parent').
		// Clone the parent proto because resolveCommit will modify it.
		parent = proto.Clone(parent).(*pfs.Commit)
		parentCommitInfo, err := d.resolveCommit(txnCtx.SqlTx, parent)
		if err != nil {
			return nil, errors.Wrapf(err, "parent commit not found")
		}
		// fail if the parent commit has not been finished
		if parentCommitInfo.Finishing == nil {
			return nil, errors.Errorf("parent commit %s has not been finished", parent)
		}

		newCommitInfo.ParentCommit = parentCommitInfo.Commit
		parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)

		if err := d.commits.ReadWrite(txnCtx.SqlTx).Put(parentCommitInfo.Commit, parentCommitInfo); err != nil {
			// Note: error is emitted if parent.ID is a missing/invalid branch OR a
			// missing/invalid commit ID
			return nil, errors.Wrapf(err, "could not resolve parent commit %s", parent)
		}
	}

	// Finally, create the commit
	if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(newCommit, newCommitInfo); err != nil {
		if col.IsErrExists(err) {
			return nil, errors.EnsureStack(pfsserver.ErrInconsistentCommit{Commit: newCommit, Branch: newCommit.Branch})
		}
		return nil, errors.EnsureStack(err)
	}
	// Defer propagation of the commit until the end of the transaction so we can
	// batch downstream commits together if there are multiple changes.
	if err := txnCtx.PropagateBranch(branch); err != nil {
		return nil, err
	}
	return newCommit, nil
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
	if commitInfo.Origin.Kind == pfs.OriginKind_ALIAS {
		return errors.Errorf("cannot finish an alias commit: %s", commitInfo.Commit)
	}
	if !force && len(commitInfo.DirectProvenance) > 0 {
		if info, err := d.env.GetPPSServer().InspectPipelineInTransaction(txnCtx,
			commit.Branch.Repo.Name,
		); err != nil && !errutil.IsNotFoundError(err) {
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

// resolveAlias finds the first ancestor of the source commit which is not an alias (possibly source itself)
func (d *driver) resolveAlias(txnCtx *txncontext.TransactionContext, source *pfs.Commit) (*pfs.CommitInfo, error) {
	baseInfo, err := d.resolveCommit(txnCtx.SqlTx, proto.Clone(source).(*pfs.Commit))
	if err != nil {
		return nil, err
	}

	for baseInfo.Origin.Kind == pfs.OriginKind_ALIAS {
		if baseInfo, err = d.resolveCommit(txnCtx.SqlTx, baseInfo.ParentCommit); err != nil {
			return nil, err
		}
	}
	return baseInfo, nil
}

func (d *driver) aliasCommit(txnCtx *txncontext.TransactionContext, parent *pfs.Commit, branch *pfs.Branch) (*pfs.CommitInfo, error) {
	// It is considered an error if the CommitSet attempts to use two different
	// commits from the same branch.  Therefore, if there is already a row for the
	// given branch and it doesn't reference the same parent commit, we fail.  In
	// the future it might be useful to be able to start and finish multiple
	// commits on the same branch within a transaction, but this should have the
	// same end result as starting and finshing a single commit on that branch, so
	// there isn't a clear use case, so it is treated like an error for now to
	// simplify PFS logic.
	commit := &pfs.Commit{
		Branch: proto.Clone(branch).(*pfs.Branch),
		ID:     txnCtx.CommitSetID,
	}

	// Update the branch head to point to the alias
	// TODO(global ids): we likely want this behavior to be optional, like when
	// doing a 'run pipeline' with explicit provenance (to make off-head commits
	// in the branch).
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(branch, branchInfo); err != nil {
		return nil, errors.EnsureStack(err)
	}

	// Check if the alias already exists
	commitInfo := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(commit, commitInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return nil, errors.EnsureStack(err)
		}
		// No commit already exists, create a new one
		// First load the parent commit and update it to point to the child
		parentCommitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(parent, parentCommitInfo, func() error {
			parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, commit)
			return nil
		}); err != nil {
			if col.IsErrNotFound(err) {
				return nil, pfsserver.ErrCommitNotFound{Commit: parent}
			}
			return nil, errors.EnsureStack(err)
		}

		commitInfo = &pfs.CommitInfo{
			Commit:           commit,
			Origin:           &pfs.CommitOrigin{Kind: pfs.OriginKind_ALIAS},
			ParentCommit:     parent,
			ChildCommits:     []*pfs.Commit{},
			Started:          txnCtx.Timestamp,
			DirectProvenance: branchInfo.DirectProvenance,
		}
		if parentCommitInfo.Finishing != nil {
			commitInfo.Finishing = txnCtx.Timestamp
			if parentCommitInfo.Finished != nil {
				commitInfo.Finished = txnCtx.Timestamp
				commitInfo.Details = parentCommitInfo.Details
				if parentCommitInfo.Error == "" {
					total, err := d.commitStore.GetTotalFileSetTx(txnCtx.SqlTx, parentCommitInfo.Commit)
					if err != nil {
						return nil, errors.EnsureStack(err)
					}
					if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commitInfo.Commit, *total); err != nil {
						return nil, errors.EnsureStack(err)
					}
				}
			}
			commitInfo.Error = parentCommitInfo.Error
		}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(commitInfo.Commit, commitInfo); err != nil {
			return nil, errors.EnsureStack(err)
		}
	} else {
		// A commit at the current transaction's ID already exists, make sure it is already compatible
		parentRoot, err := d.resolveAlias(txnCtx, parent)
		if err != nil {
			return nil, err
		}
		prevRoot, err := d.resolveAlias(txnCtx, commitInfo.Commit)
		if err != nil {
			return nil, err
		}
		if !proto.Equal(parentRoot.Commit, prevRoot.Commit) {
			return nil, errors.EnsureStack(pfsserver.ErrInconsistentCommit{Commit: parent, Branch: branch})
		}
	}

	// Update the branch head
	branchInfo.Head = commit
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(branch, branchInfo); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return commitInfo, nil
}

func (d *driver) repoSize(ctx context.Context, repo *pfs.Repo) (int64, error) {
	repoInfo := new(pfs.RepoInfo)
	if err := d.repos.ReadOnly(ctx).Get(repo, repoInfo); err != nil {
		return 0, errors.EnsureStack(err)
	}
	for _, branch := range repoInfo.Branches {
		if branch.Name == "master" {
			branchInfo := &pfs.BranchInfo{}
			if err := d.branches.ReadOnly(ctx).Get(branch, branchInfo); err != nil {
				return 0, errors.EnsureStack(err)
			}
			commit := branchInfo.Head
			for commit != nil {
				commitInfo, err := d.getCommit(ctx, commit)
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

// propagateBranches selectively starts commits in or downstream of 'branches'
// in order to restore the invariant that branch provenance matches HEAD commit
// provenance:
//   B.Head is provenant on A.Head <=>
//   branch B is provenant on branch A
// The implementation assumes that the invariant already holds for all branches
// upstream of 'branches', but not necessarily for each 'branch' itself. Despite
// the name, 'branches' do not need a HEAD commit to propagate, though one may
// be created.
//
// In other words, propagateBranches scans all branches b_downstream that are
// equal to or downstream of 'branches', and if the HEAD of b_downstream isn't
// provenant on the HEADs of b_downstream's provenance, propagateBranches starts
// a new HEAD commit in b_downstream that is. For example, propagateBranches
// starts downstream output commits (which trigger PPS jobs) when new input
// commits arrive on 'branch', when 'branches's HEAD is deleted, or when
// 'branches' are newly created (i.e. in CreatePipeline).
func (d *driver) propagateBranches(txnCtx *txncontext.TransactionContext, branches []*pfs.Branch) error {
	branchInfoCache := map[string]*pfs.BranchInfo{}
	getBranchInfo := func(branch *pfs.Branch) (*pfs.BranchInfo, error) {
		if branchInfo, ok := branchInfoCache[pfsdb.BranchKey(branch)]; ok {
			return branchInfo, nil
		}
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(branch, branchInfo); err != nil {
			return nil, errors.EnsureStack(err)
		}
		branchInfoCache[pfsdb.BranchKey(branch)] = branchInfo
		return branchInfo, nil
	}

	// subvBIMap = ( ⋃{b.subvenance | b ∈ branches} ) ∪ branches
	// (after branches has been pruned of no-op members, which don't actually need propagation)
	subvBIMap := map[string]*pfs.BranchInfo{}
	for _, branch := range branches {
		branchInfo, err := getBranchInfo(branch)
		if err != nil {
			return err
		}

		// We need to create new commits or aliases if any of this branch and its
		// provenances disagree on their commit set.
		ids := []string{branchInfo.Head.ID}
		for _, provBranch := range branchInfo.Provenance {
			provInfo, err := getBranchInfo(provBranch)
			if err != nil {
				return err
			}
			ids = append(ids, provInfo.Head.ID)
		}
		if allSameString(ids) && ids[0] != txnCtx.CommitSetID {
			// this branch hasn't changed during the transaction, and is still consistent with its provenance,
			// so we don't need to include it or its subvenance
			continue
		}

		subvBIMap[pfsdb.BranchKey(branch)] = branchInfo
		for _, subvBranch := range branchInfo.Subvenance {
			subvInfo, err := getBranchInfo(subvBranch)
			if err != nil {
				return err
			}
			subvBIMap[pfsdb.BranchKey(subvBranch)] = subvInfo
		}
	}

	// 'subvBIs' is the collection of downstream branches that will get a new
	// commit. Populate subvBIs and sort it so that upstream branches are
	// processed before their descendants (this guarantees that if branch B is
	// provenant on branch A, we create a new commit in A before creating a new
	// commit in B provenant on the new HEAD of A).
	var subvBIs []*pfs.BranchInfo
	for _, branchData := range subvBIMap {
		subvBIs = append(subvBIs, branchData)
	}
	sort.Slice(subvBIs, func(i, j int) bool {
		return len(subvBIs[i].Provenance) < len(subvBIs[j].Provenance)
	})

	// Iterate through downstream branches and add new commits.
	hasNewCommits := false
	for _, subvBI := range subvBIs {
		// Do not propagate an open commit onto spout output branches (which should
		// only have a single provenance on a spec commit)
		if len(subvBI.Provenance) == 1 && subvBI.Provenance[0].Repo.Type == pfs.SpecRepoType {
			continue
		}
		hasNewCommits = true

		// Create aliases for any provenant branches which are not already part of this CommitSet
		for _, provOfSubvB := range subvBI.Provenance {
			provOfSubvBI, err := getBranchInfo(provOfSubvB)
			if err != nil {
				return err
			}
			if provOfSubvBI.Head.ID != txnCtx.CommitSetID {
				if _, err := d.aliasCommit(txnCtx, provOfSubvBI.Head, provOfSubvBI.Head.Branch); err != nil {
					return err
				}
				// Update the cached branch head
				provOfSubvBI.Head.ID = txnCtx.CommitSetID
			}
			// if this is a pipeline output branch, we need to also create an alias commit on the meta branch
			// to maintain pipeline system invariants.
			if provOfSubvBI.Branch.Repo.Type == pfs.UserRepoType {
				metaBranch := client.NewSystemRepo(provOfSubvBI.Branch.Repo.Name, pfs.MetaRepoType).
					NewBranch(provOfSubvBI.Branch.Name)
				metaBI, err := getBranchInfo(metaBranch)
				if err != nil {
					if col.IsErrNotFound(err) {
						// no corresponding meta branch, so not a pipeline. Ignore
						continue
					}
					return err
				}
				// create the alias if necessary, just like above
				if metaBI.Head.ID != txnCtx.CommitSetID {
					if _, err := d.aliasCommit(txnCtx, metaBI.Head, metaBI.Head.Branch); err != nil {
						return err
					}
					metaBI.Head.ID = txnCtx.CommitSetID
				}
			}
		}

		if subvBI.Head.ID == txnCtx.CommitSetID {
			continue // this branch is already updated
		}

		// determine whether we can use the contents of an old commit
		var oldCommit pfs.CommitInfo
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(
			subvBI.Branch.NewCommit(txnCtx.CommitSetID), &oldCommit,
		); err != nil && !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		} else if err == nil {
			if len(subvBI.DirectProvenance) != len(oldCommit.DirectProvenance) {
				return errors.EnsureStack(pfsserver.ErrInconsistentCommit{Branch: subvBI.Branch, Commit: oldCommit.Commit})
			}
			for _, br := range oldCommit.DirectProvenance {
				if !has(&subvBI.DirectProvenance, br) {
					return errors.EnsureStack(pfsserver.ErrInconsistentCommit{Branch: subvBI.Branch, Commit: oldCommit.Commit})
				}
			}
			// the old commit is compatible with the current provenance, so use it.
			// This will reuse the old data and not create a job, meaning if the reprocess spec is "every job",
			// moving a branch head back is different from doing the inverse changes in PFS
			subvBI.Head = oldCommit.Commit
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(subvBI.Branch, subvBI); err != nil {
				return errors.EnsureStack(err)
			}
		} else {
			// This branch has no commit for this CommitSet, start a new output commit in 'subvBI.Branch'
			newCommit := &pfs.Commit{
				Branch: subvBI.Branch,
				ID:     txnCtx.CommitSetID,
			}
			newCommitInfo := &pfs.CommitInfo{
				Commit:           newCommit,
				Origin:           &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
				Started:          txnCtx.Timestamp,
				DirectProvenance: subvBI.DirectProvenance,
			}

			// we might be able to find an older parent commit that better reflects the provenance state, saving work

			// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
			newCommitInfo.ParentCommit = subvBI.Head
			subvBI.Head = newCommit
			if newCommitInfo.ParentCommit != nil {
				parentCommitInfo := &pfs.CommitInfo{}
				if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(newCommitInfo.ParentCommit, parentCommitInfo, func() error {
					parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
					return nil
				}); err != nil {
					return errors.EnsureStack(err)
				}
			}

			if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(subvBI.Branch, subvBI); err != nil {
				return errors.EnsureStack(err)
			}

			// finally create open 'commit'
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(newCommit, newCommitInfo); err != nil && col.IsErrExists(err) {
				return errors.EnsureStack(pfsserver.ErrInconsistentCommit{
					Commit: newCommit,
					Branch: newCommit.Branch,
				})
			} else if err != nil {
				return errors.EnsureStack(err)
			}
		}
	}

	// If we have any PFS changes in this transaction, write out the CommitSet
	if hasNewCommits {
		txnCtx.PropagateJobs()
	}

	return nil
}

// inspectCommit takes a Commit and returns the corresponding CommitInfo.
//
// As a side effect, this function also replaces the ID in the given commit
// with a real commit ID.
func (d *driver) inspectCommit(ctx context.Context, commit *pfs.Commit, wait pfs.CommitState) (*pfs.CommitInfo, error) {
	if commit.Branch.Repo.Name == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "FileSet - Virtual Commit",
			Finished:    &types.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return cinfo, nil
	}
	if commit == nil {
		return nil, errors.Errorf("cannot inspect nil commit")
	}
	if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, commit.Branch.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
		return nil, errors.EnsureStack(err)
	}

	// TODO(global ids): it's possible the commit doesn't exist yet (but will,
	// following a trigger).  If the commit isn't found, check if the associated
	// commitset _could_ reach the requested branch or ID via a trigger and wait
	// to find out.
	// Resolve the commit in case it specifies a branch head or commit ancestry
	var commitInfo *pfs.CommitInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commitInfo, err = d.resolveCommit(txnCtx.SqlTx, commit)
		return err
	}); err != nil {
		return nil, err
	}

	if commitInfo.Finished == nil {
		switch wait {
		case pfs.CommitState_READY:
			for _, branch := range commitInfo.DirectProvenance {
				if _, err := d.inspectCommit(ctx, branch.NewCommit(commit.ID), pfs.CommitState_FINISHED); err != nil {
					return nil, err
				}
			}
		case pfs.CommitState_FINISHED:
			// Watch the CommitInfo until the commit has been finished
			if err := d.commits.ReadOnly(ctx).WatchOneF(commit, func(ev *watch.Event) error {
				if ev.Type == watch.EventDelete {
					return pfsserver.ErrCommitDeleted{Commit: commit}
				}

				var key string
				newCommitInfo := &pfs.CommitInfo{}
				if err := ev.Unmarshal(&key, newCommitInfo); err != nil {
					return errors.Wrapf(err, "unmarshal")
				}
				if newCommitInfo.Finished != nil {
					commitInfo = newCommitInfo
					return errutil.ErrBreak
				}
				return nil
			}); err != nil {
				return nil, errors.EnsureStack(err)
			}
		case pfs.CommitState_STARTED:
			// Do nothing
		}
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
	if userCommit.Branch == nil {
		return nil, errors.Errorf("cannot resolve commit with no branch")
	}
	if userCommit.Branch.Repo == nil {
		return nil, errors.Errorf("cannot resolve commit with no repo")
	}
	if userCommit.ID == "" && userCommit.Branch.Name == "" {
		return nil, errors.Errorf("cannot resolve commit with no ID or branch")
	}
	commit := proto.Clone(userCommit).(*pfs.Commit) // back up user commit, for error reporting
	// Extract any ancestor tokens from 'commit.ID' (i.e. ~, ^ and .)
	var ancestryLength int
	var err error
	commit.ID, ancestryLength, err = ancestry.Parse(commit.ID)
	if err != nil {
		return nil, err
	}

	// Now that ancestry has been parsed out, check if the ID is a branch name
	if commit.ID != "" && !uuid.IsUUIDWithoutDashes(commit.ID) {
		if commit.Branch.Name != "" {
			return nil, errors.Errorf("invalid commit ID given with a branch (%s): %s\n", commit.Branch, commit.ID)
		}
		commit.Branch.Name = commit.ID
		commit.ID = ""
	}

	if commit.ID == "" {
		// If commit.ID is unspecified, get it from the branch head
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(sqlTx).Get(commit.Branch, branchInfo); err != nil {
			return nil, errors.EnsureStack(err)
		}
		commit.ID = branchInfo.Head.ID
	} else if commit.Branch.Name == "" {
		// If the branch is unspecified, make sure the ID is unique (a repo may have
		// one commit on each branch with the same ID) and load the branch name.
		commitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(sqlTx).GetByIndex(pfsdb.CommitsBranchlessIndex, pfsdb.CommitBranchlessKey(commit), commitInfo, col.DefaultOptions(), func(string) error {
			if commit.Branch.Name != "" {
				return pfsserver.ErrAmbiguousCommit{Commit: userCommit}
			}
			commit.Branch.Name = commitInfo.Commit.Branch.Name
			return nil
		}); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}

	// Traverse commits' parents until you've reached the right ancestor
	commitInfo := &pfs.CommitInfo{}
	if ancestryLength >= 0 {
		for i := 0; i <= ancestryLength; i++ {
			if commit == nil {
				return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
			}
			if err := d.commits.ReadWrite(sqlTx).Get(commit, commitInfo); err != nil {
				if col.IsErrNotFound(err) {
					if i == 0 {
						return nil, pfsserver.ErrCommitNotFound{Commit: userCommit}
					}
					return nil, pfsserver.ErrParentCommitNotFound{Commit: commit}
				}
				return nil, errors.EnsureStack(err)
			}
			commit = commitInfo.ParentCommit
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
			}
			commit = cis[i%len(cis)].ParentCommit
		}
	}
	userCommit.Branch = proto.Clone(commitInfo.Commit.Branch).(*pfs.Branch)
	userCommit.ID = commitInfo.Commit.ID
	return commitInfo, nil
}

// getCommit is like inspectCommit, without the blocking.
// It does not add the size to the CommitInfo
func (d *driver) getCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	if commit.Branch.Repo.Name == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "FileSet - Virtual Commit",
			Finished:    &types.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
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
// subscribeCommit to apply filtering to the returned commits.  By default we
// skip over alias commits, but we allow users to request all the commits with
// 'all', or a specific type of commit with 'originKind'.
func passesCommitOriginFilter(commitInfo *pfs.CommitInfo, all bool, originKind pfs.OriginKind) bool {
	if all {
		return true
	} else if originKind != pfs.OriginKind_ORIGIN_KIND_UNKNOWN {
		return commitInfo.Origin.Kind == originKind
	}
	return commitInfo.Origin.Kind != pfs.OriginKind_ALIAS
}

func (d *driver) listCommit(
	ctx context.Context,
	repo *pfs.Repo,
	to *pfs.Commit,
	from *pfs.Commit,
	startTime *types.Timestamp,
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
	if from != nil && !proto.Equal(from.Branch.Repo, repo) || to != nil && !proto.Equal(to.Branch.Repo, repo) {
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
		if _, err := d.inspectCommit(ctx, to, pfs.CommitState_STARTED); err != nil {
			return err
		}
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
		for number != 0 && cursor != nil && (from == nil || cursor.ID != from.ID) {
			commitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadOnly(ctx).Get(cursor, commitInfo); err != nil {
				return errors.EnsureStack(err)
			}
			if passesCommitOriginFilter(commitInfo, all, originKind) {
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

func (d *driver) inspectCommitSetImmediate(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) ([]*pfs.CommitInfo, error) {
	commitMap := map[string]*pfs.CommitInfo{}
	commitInfo := &pfs.CommitInfo{}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).GetByIndex(pfsdb.CommitsCommitSetIndex, commitset.ID, commitInfo, col.DefaultOptions(), func(string) error {
		commitMap[pfsdb.BranchKey(commitInfo.Commit.Branch)] = proto.Clone(commitInfo).(*pfs.CommitInfo)
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}

	if len(commitMap) == 0 {
		return nil, pfsserver.ErrCommitSetNotFound{CommitSet: commitset}
	}

	// Do a topological sort of the commitInfos (note that this isn't a stable
	// sort, but we could do it if that becomes a problem)
	result := []*pfs.CommitInfo{}
	added := map[string]struct{}{}
	provMap := map[string][]string{} // map of provenance -> subvenance
	for _, commitInfo := range commitMap {
		commitKey := pfsdb.BranchKey(commitInfo.Commit.Branch)
		for _, prov := range commitInfo.DirectProvenance {
			provKey := pfsdb.BranchKey(prov)
			if provs, ok := provMap[provKey]; ok {
				provMap[provKey] = append(provs, commitKey)
			} else {
				provMap[provKey] = []string{commitKey}
			}
		}
		if len(commitInfo.DirectProvenance) == 0 {
			result = append(result, commitInfo)
			added[commitKey] = struct{}{}
		}
	}

	for i := 0; i < len(result); i++ {
		commitInfo := result[i]
		for _, provKey := range provMap[pfsdb.BranchKey(commitInfo.Commit.Branch)] {
			if ci, ok := commitMap[provKey]; ok {
				if _, ok := added[provKey]; !ok {
					result = append(result, ci)
					added[provKey] = struct{}{}
				}
			}
		}
	}

	if len(result) != len(commitMap) {
		return nil, errors.Errorf("internal error: incomplete commitset provenance for %s", commitset.ID)
	}
	return result, nil
}

func (d *driver) inspectCommitSet(ctx context.Context, commitset *pfs.CommitSet, wait bool, cb func(*pfs.CommitInfo) error) error {
	sent := map[string]struct{}{}

	// The commits in this CommitSet may change if any triggers or CreateBranches
	// add more, so reload it after each wait.
reloadCommitSet:
	for {
		var commitInfos []*pfs.CommitInfo
		if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			var err error
			commitInfos, err = d.inspectCommitSetImmediate(txnCtx, commitset)
			return err
		}); err != nil {
			return err
		}

		reload := false
		for _, commitInfo := range commitInfos {
			// If we aren't blocking, we can just loop over everything once and return
			if wait {
				if _, ok := sent[pfsdb.CommitKey(commitInfo.Commit)]; ok {
					continue
				}
				// TODO: make a dedicated call just for the blocking part, inspectCommit is a little heavyweight?
				var err error
				commitInfo, err = d.inspectCommit(ctx, commitInfo.Commit, pfs.CommitState_FINISHED)
				if err != nil {
					return err
				}
				reload = true
			}
			if err := cb(commitInfo); err != nil {
				return err
			}
			sent[pfsdb.CommitKey(commitInfo.Commit)] = struct{}{}
			if reload {
				continue reloadCommitSet
			}
		}
		// If we didn't find any commits we haven't already sent, it is safe to return
		return nil
	}
}

func (d *driver) listCommitSet(ctx context.Context, cb func(*pfs.CommitSetInfo) error) error {
	// Track the commitsets we've already processed
	seen := map[string]struct{}{}

	// Return commitsets by the newest commit in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	commitInfo := &pfs.CommitInfo{}
	err := d.commits.ReadOnly(ctx).List(commitInfo, col.DefaultOptions(), func(string) error {
		if _, ok := seen[commitInfo.Commit.ID]; ok {
			return nil
		}
		seen[commitInfo.Commit.ID] = struct{}{}
		var commitInfos []*pfs.CommitInfo
		err := d.inspectCommitSet(ctx, &pfs.CommitSet{ID: commitInfo.Commit.ID}, false, func(ci *pfs.CommitInfo) error {
			commitInfos = append(commitInfos, ci)
			return nil
		})
		if err != nil {
			return err
		}
		return cb(&pfs.CommitSetInfo{
			CommitSet: client.NewCommitSet(commitInfo.Commit.ID),
			Commits:   commitInfos,
		})
	})
	return errors.EnsureStack(err)
}

func (d *driver) squashCommitSetInternal(txnCtx *txncontext.TransactionContext, commitInfos []*pfs.CommitInfo) error {
	deleted := make(map[string]*pfs.CommitInfo) // deleted commits

	// 1) Delete each commit in the CommitSet
	affectedBranches := []*pfs.Branch{}
	for _, commitInfo := range commitInfos {
		deleted[pfsdb.CommitKey(commitInfo.Commit)] = commitInfo
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Delete(commitInfo.Commit); err != nil {
			return errors.EnsureStack(err)
		}

		// make sure all children are finished, so we don't lose data
		for _, child := range commitInfo.ChildCommits {
			if _, ok := deleted[pfsdb.CommitKey(child)]; ok {
				// this child is being deleted, any files from this commit will end up
				// as part of *its* children, which have already been checked
				continue
			}
			var childInfo pfs.CommitInfo
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(child, &childInfo); err != nil {
				return errors.Wrapf(err, "error checking child commit state")
			}
			if childInfo.Finished == nil {
				var suffix string
				if childInfo.Finishing != nil {
					// user might already have called "finish",
					suffix = ", consider using WaitCommit"
				}
				return errors.Errorf("cannot squash until child commit %s is finished%s", child, suffix)
			}
		}

		// Delete the commit's filesets
		if err := d.commitStore.DropFileSetsTx(txnCtx.SqlTx, commitInfo.Commit); err != nil {
			return errors.EnsureStack(err)
		}

		// Update the commit's branch's branchInfo in case this was the head of the branch
		branchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(commitInfo.Commit.Branch, branchInfo, func() error {
			if branchInfo.Head.ID == commitInfo.Commit.ID {
				if commitInfo.ParentCommit == nil || !proto.Equal(commitInfo.ParentCommit.Branch, commitInfo.Commit.Branch) {
					// Create a new empty commit for the branch head
					var err error
					branchInfo.Head, err = d.makeEmptyCommit(txnCtx, branchInfo)
					if err != nil {
						return err
					}
				} else {
					branchInfo.Head = commitInfo.ParentCommit
				}
				affectedBranches = append(affectedBranches, commitInfo.Commit.Branch)
			}
			return nil
		}); err != nil && !col.IsErrNotFound(err) {
			// If err is NotFound, branch is in downstream provenance but
			// doesn't exist yet (or branch may have been deleted) --nothing to update
			return errors.Wrapf(err, "error updating branch %s", commitInfo.Commit.Branch)
		}
	}

	// 2) Rewrite ParentCommit of deleted commits' children, and
	// ChildCommits of deleted commits' parents
	visited := make(map[string]struct{}) // visited child/parent commits
	for _, deletedInfo := range deleted {
		if _, ok := visited[pfsdb.CommitKey(deletedInfo.Commit)]; ok {
			continue
		}

		// Traverse parents until we find the most ancestral non-nil, deleted commit
		oldestCommitInfo := deletedInfo
		for {
			if oldestCommitInfo.ParentCommit == nil {
				break // parent is nil
			}
			parentInfo, ok := deleted[pfsdb.CommitKey(oldestCommitInfo.ParentCommit)]
			if !ok {
				break // parent is not deleted
			}
			oldestCommitInfo = parentInfo // parent exists and is deleted, keep going
		}

		// BFS for all non-deleted children
		var next *pfs.Commit                            // next vertex to search
		queue := []*pfs.Commit{oldestCommitInfo.Commit} // queue of vertices to explore
		liveChildren := make(map[string]*pfs.Commit)    // live children discovered so far
		for len(queue) > 0 {
			next, queue = queue[0], queue[1:]
			if _, ok := visited[pfsdb.CommitKey(next)]; ok {
				continue
			}
			visited[pfsdb.CommitKey(next)] = struct{}{}
			nextInfo, ok := deleted[pfsdb.CommitKey(next)]
			if !ok {
				liveChildren[pfsdb.CommitKey(next)] = next
				continue
			}
			queue = append(queue, nextInfo.ChildCommits...)
		}

		// Point all non-deleted children at the first valid parent (or nil),
		// and point first non-deleted parent at all non-deleted children
		parent := oldestCommitInfo.ParentCommit
		for _, commit := range liveChildren {
			commitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(commit, commitInfo, func() error {
				commitInfo.ParentCommit = parent
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err updating child commit %s", oldestCommitInfo.Commit)
			}
		}
		if parent != nil {
			commitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(parent, commitInfo, func() error {
				// Add existing live commits in commitInfo.ChildCommits to the
				// live children above oldestCommitInfo, then put them all in
				// 'parent'
				for _, child := range commitInfo.ChildCommits {
					if _, ok := deleted[pfsdb.CommitKey(child)]; ok {
						continue
					}
					liveChildren[pfsdb.CommitKey(child)] = child
				}
				commitInfo.ChildCommits = make([]*pfs.Commit, 0, len(liveChildren))
				for _, commit := range liveChildren {
					commitInfo.ChildCommits = append(commitInfo.ChildCommits, commit)
				}
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err rewriting children of ancestor commit %s", oldestCommitInfo.Commit)
			}
		}
	}

	// 4) propagate the changes to 'branch' and its subvenance. This may start
	// new HEAD commits downstream, if the new branch heads haven't been
	// processed yet
	for _, branch := range affectedBranches {
		if err := txnCtx.PropagateBranch(branch); err != nil {
			return err
		}
	}

	return nil
}

// dropCommitSet is only implemented for commits with no children, so if any
// commits in the commitSet have children the operation will fail.
func (d *driver) dropCommitSet(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	// Look up the commits in the CommitSet
	commitInfos, err := d.inspectCommitSetImmediate(txnCtx, commitset)
	if err != nil {
		return err
	}

	for _, ci := range commitInfos {
		if len(ci.ChildCommits) > 0 {
			return &pfsserver.ErrDropWithChildren{Commit: ci.Commit}
		}
	}

	// While this is a 'drop' operation and not a 'squash', proper drop semantics
	// aren't implemented at the moment.  Squashing the head of a branch is
	// effectively a drop, though, because there is no child commit that contains
	// the data from the given commits, which is why it is an error to drop any
	// non-head commits (until generalized drop semantics are implemented).
	if err := d.squashCommitSetInternal(txnCtx, commitInfos); err != nil {
		return err
	}

	// notify PPS that this commitset has been dropped so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobs(commitset)
	return nil
}

func (d *driver) squashCommitSet(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) error {
	// Look up the commits in the CommitSet
	commitInfos, err := d.inspectCommitSetImmediate(txnCtx, commitset)
	if err != nil {
		return err
	}

	for _, ci := range commitInfos {
		if ci.Commit.Branch.Repo.Type == pfs.SpecRepoType && ci.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", ci.Commit)
		}
		if len(ci.ChildCommits) == 0 {
			return &pfsserver.ErrSquashWithoutChildren{Commit: ci.Commit}
		}
	}

	if err := d.squashCommitSetInternal(txnCtx, commitInfos); err != nil {
		return err
	}

	// notify PPS that this commitset has been squashed so it can clean up any
	// jobs associated with it at the end of the transaction
	txnCtx.StopJobs(commitset)
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
	if from != nil && !proto.Equal(from.Branch.Repo, repo) {
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
		if !(seen[commitInfo.Commit.ID] || (from != nil && from.ID == commitInfo.Commit.ID)) {
			// Wait for the commit to enter the right state
			commitInfo, err := d.inspectCommit(ctx, proto.Clone(commitInfo.Commit).(*pfs.Commit), state)
			if err != nil {
				return err
			}
			if err := cb(commitInfo); err != nil {
				return err
			}
			seen[commitInfo.Commit.ID] = true
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

// createBranch creates a new branch or updates an existing branch (must be one
// or the other). Most importantly, it sets 'branch.DirectProvenance' to
// 'provenance' and then for all (downstream) branches, restores the invariant:
//   ∀ b . b.Provenance = ∪ b'.Provenance (where b' ∈ b.DirectProvenance)
//
// This invariant is assumed to hold for all branches upstream of 'branch', but not
// for 'branch' itself once 'b.Provenance' has been set.
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
	if err := ancestry.ValidateName(branch.Name); err != nil {
		return err
	}

	// Retrieve (and create, if necessary) the current version of this branch
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Upsert(branch, branchInfo, func() error {
		branchInfo.Branch = branch
		branchInfo.DirectProvenance = nil
		for _, provBranch := range provenance {
			if proto.Equal(provBranch.Repo, branch.Repo) {
				return errors.Errorf("repo %s cannot be in the provenance of its own branch", branch.Repo)
			}
			add(&branchInfo.DirectProvenance, provBranch)
		}
		if trigger != nil && trigger.Branch != "" {
			branchInfo.Trigger = trigger
		}
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}

	var ci *pfs.CommitInfo
	if commit != nil {
		// resolve the given commit
		ci, err = d.resolveCommit(txnCtx.SqlTx, commit)
		if err != nil {
			return errors.Wrapf(err, "unable to inspect %s", commit)
		}

		// Verify the provenance of the new branch head and lock in its upstream commits
		for _, provBranch := range provenance {
			// Check that the CommitSet for the given commit has values for every branch in provenance and alias them
			if _, err := d.aliasCommit(txnCtx, provBranch.NewCommit(ci.Commit.ID), provBranch); err != nil {
				if pfsserver.IsCommitNotFoundErr(err) {
					return errors.Errorf("cannot create branch %s with commit %s as head because it does not have provenance in the %s branch", branch, ci.Commit, provBranch)
				}
			}
		}

		if commit.ID == txnCtx.CommitSetID && proto.Equal(commit.Branch, branchInfo.Branch) {
			// We can reuse the existing commit only if it is already on this branch
			branchInfo.Head = commit
		} else if branchInfo.Head == nil || branchInfo.Head.ID != commit.ID {
			// Create an alias of the head commit onto this branch - this will move the
			// head of the branch and update the repo size if necessary
			aliasCommitInfo, err := d.aliasCommit(txnCtx, commit, branch)
			if err != nil {
				return err
			}
			// Update the local branchInfo.Head
			branchInfo.Head = aliasCommitInfo.Commit
		}
	}

	// If the branch still has no head, create an empty commit on it so that we
	// can maintain an invariant that branches always have a head commit.
	if branchInfo.Head == nil {
		branchInfo.Head, err = d.makeEmptyCommit(txnCtx, branchInfo)
		if err != nil {
			return err
		}
	}

	// Update (or create)
	// 1) 'branch's Provenance
	// 2) the Provenance of all branches in 'branch's Subvenance (in the case of an update), and
	// 3) the Subvenance of all branches in the *old* provenance of 'branch's Subvenance
	toUpdate := []*pfs.BranchInfo{branchInfo}
	for _, subvBranch := range branchInfo.Subvenance {
		subvBranchInfo := &pfs.BranchInfo{}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(subvBranch, subvBranchInfo); err != nil {
			return errors.EnsureStack(err)
		}
		toUpdate = append(toUpdate, subvBranchInfo)
	}
	// Sorting is important here because it sorts topologically. This means
	// that when evaluating element i of `toUpdate` all elements < i will
	// have already been evaluated and thus we can safely use their
	// Provenance field.
	sort.Slice(toUpdate, func(i, j int) bool { return len(toUpdate[i].Provenance) < len(toUpdate[j].Provenance) })
	for _, branchInfo := range toUpdate {
		oldProvenance := branchInfo.Provenance
		branchInfo.Provenance = nil
		// Re-compute Provenance
		for _, provBranch := range branchInfo.DirectProvenance {
			if err := d.addBranchProvenance(txnCtx, branchInfo, provBranch); err != nil {
				return err
			}
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Get(provBranch, provBranchInfo); err != nil {
				return errors.Wrapf(err, "error getting prov branch")
			}
			for _, provBranch := range provBranchInfo.Provenance {
				// add provBranch to branchInfo.Provenance, and branchInfo.Branch to
				// provBranch subvenance
				if err := d.addBranchProvenance(txnCtx, branchInfo, provBranch); err != nil {
					return err
				}
			}
		}
		if err := d.branches.ReadWrite(txnCtx.SqlTx).Put(branchInfo.Branch, branchInfo); err != nil {
			return errors.EnsureStack(err)
		}
		// Update Subvenance of 'branchInfo's Provenance (incl. all Subvenance)
		for _, oldProvBranch := range oldProvenance {
			if !has(&branchInfo.Provenance, oldProvBranch) {
				// Provenance was deleted, so we delete ourselves from their subvenance
				oldProvBranchInfo := &pfs.BranchInfo{}
				if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(oldProvBranch, oldProvBranchInfo, func() error {
					del(&oldProvBranchInfo.Subvenance, branchInfo.Branch)
					return nil
				}); err != nil {
					return errors.EnsureStack(err)
				}
			}
		}
	}

	// Add the new branch to the repo info
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.SqlTx).Update(branch.Repo, repoInfo, func() error {
		add(&repoInfo.Branches, branch)
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}

	if commit != nil && ci.Finished != nil {
		if err = d.triggerCommit(txnCtx, branchInfo.Head); err != nil {
			return err
		}
	}

	// propagate the head commit to 'branch'. This may also modify 'branch', by
	// creating a new HEAD commit if 'branch's provenance was changed and its
	// current HEAD commit has old provenance
	if err := txnCtx.PropagateBranch(branch); err != nil {
		return err
	}
	return nil
}

func (d *driver) inspectBranch(txnCtx *txncontext.TransactionContext, branch *pfs.Branch) (*pfs.BranchInfo, error) {
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
			return errors.Wrapf(err, "branches.Get")
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

func (d *driver) addBranchProvenance(txnCtx *txncontext.TransactionContext, branchInfo *pfs.BranchInfo, provBranch *pfs.Branch) error {
	if pfsdb.BranchKey(provBranch) == pfsdb.BranchKey(branchInfo.Branch) {
		return errors.Errorf("provenance loop, branch %s cannot be provenant on itself", provBranch)
	}
	add(&branchInfo.Provenance, provBranch)
	provBranchInfo := &pfs.BranchInfo{}
	if err := d.branches.ReadWrite(txnCtx.SqlTx).Upsert(provBranch, provBranchInfo, func() error {
		if provBranchInfo.Branch == nil {
			// We are creating this branch for the first time, set the Branch and Head
			provBranchInfo.Branch = provBranch

			head, err := d.makeEmptyCommit(txnCtx, provBranchInfo)
			if err != nil {
				return err
			}
			provBranchInfo.Head = head
		}
		add(&provBranchInfo.Subvenance, branchInfo.Branch)
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	repoInfo := &pfs.RepoInfo{}
	err := d.repos.ReadWrite(txnCtx.SqlTx).Update(provBranch.Repo, repoInfo, func() error {
		add(&repoInfo.Branches, provBranch)
		return nil
	})
	return errors.EnsureStack(err)
}

func (d *driver) deleteAll(ctx context.Context) error {
	var repoInfos []*pfs.RepoInfo
	if err := d.listRepo(ctx, !includeAuth, "", func(repoInfo *pfs.RepoInfo) error {
		repoInfos = append(repoInfos, repoInfo)
		return nil
	}); err != nil {
		return err
	}
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		// the list does not use the transaction
		for _, repoInfo := range repoInfos {
			if err := d.deleteRepo(txnCtx, repoInfo.Repo, true); err != nil && !auth.IsErrNotAuthorized(err) {
				return err
			}
		}
		return nil
	})
}

func (d *driver) makeEmptyCommit(txnCtx *txncontext.TransactionContext, branchInfo *pfs.BranchInfo) (*pfs.Commit, error) {
	// Input repos and spouts want a closed head commit, so decide if we leave
	// it open by the presence of branch provenance.  If it's only provenant on
	// a spec repo, we assume it's a spout and close the commit.
	closed := true
	for _, prov := range branchInfo.DirectProvenance {
		if prov.Repo.Type != pfs.SpecRepoType {
			closed = false
			break
		}
	}

	commit := branchInfo.Branch.NewCommit(txnCtx.CommitSetID)
	commitInfo := &pfs.CommitInfo{
		Commit:           commit,
		Origin:           &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
		Started:          txnCtx.Timestamp,
		DirectProvenance: branchInfo.DirectProvenance,
	}
	if closed {
		commitInfo.Finishing = txnCtx.Timestamp
		commitInfo.Finished = txnCtx.Timestamp
		commitInfo.Details = &pfs.CommitInfo_Details{}
		total, err := d.storage.ComposeTx(txnCtx.SqlTx, nil, defaultTTL)
		if err != nil {
			return nil, err
		}
		if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commit, *total); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	if err := d.commits.ReadWrite(txnCtx.SqlTx).Create(commit, commitInfo); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return commit, nil
}

func (d *driver) putCache(ctx context.Context, key string, value *types.Any, fileSetIds []fileset.ID, tag string) error {
	return d.cache.Put(ctx, key, value, fileSetIds, tag)
}

func (d *driver) getCache(ctx context.Context, key string) (*types.Any, error) {
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

func getOrCreateKey(ctx context.Context, keyStore chunk.KeyStore, name string) ([]byte, error) {
	secret, err := keyStore.Get(ctx, name)
	if !errors.Is(err, sql.ErrNoRows) {
		return secret, errors.EnsureStack(err)
	}
	secret = make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, errors.EnsureStack(err)
	}
	log.Infof("generated new secret: %q", name)
	if err := keyStore.Create(ctx, name, secret); err != nil {
		return nil, errors.EnsureStack(err)
	}
	res, err := keyStore.Get(ctx, name)
	return res, errors.EnsureStack(err)
}

func allSameString(slice []string) bool {
	for _, str := range slice {
		if str != slice[0] {
			return false
		}
	}
	return true
}
