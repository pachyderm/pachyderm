package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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

type collectionFactory func(string) col.Collection

type driver struct {
	env *serviceenv.ServiceEnv
	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	txnEnv     *txnenv.TransactionEnv
	prefix     string

	// collections
	repos       col.Collection
	commits     collectionFactory
	branches    collectionFactory
	openCommits col.Collection

	storage         *fileset.Storage
	commitStore     commitStore
	compactionQueue *work.TaskQueue
}

func newDriver(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, db *sqlx.DB) (*driver, error) {
	// Setup etcd, object storage, and database clients.
	etcdClient := env.GetEtcdClient()
	objClient, err := NewObjClient(env.Configuration)
	if err != nil {
		return nil, err
	}
	// Setup driver struct.
	d := &driver{
		env:        env,
		txnEnv:     txnEnv,
		etcdClient: etcdClient,
		prefix:     etcdPrefix,
		repos:      pfsdb.Repos(etcdClient, etcdPrefix),
		commits: func(repo string) col.Collection {
			return pfsdb.Commits(etcdClient, etcdPrefix, repo)
		},
		branches: func(repo string) col.Collection {
			return pfsdb.Branches(etcdClient, etcdPrefix, repo)
		},
		openCommits: pfsdb.OpenCommits(etcdClient, etcdPrefix),
		// TODO: set maxFanIn based on downward API.
	}
	// Setup tracker and chunk / fileset storage.
	tracker := track.NewPostgresTracker(db)
	chunkStorageOpts, err := env.ChunkStorageOptions()
	if err != nil {
		return nil, err
	}
	memCache := env.ChunkMemoryCache()
	keyStore := chunk.NewPostgresKeyStore(db)
	secret, err := getOrCreateKey(context.TODO(), keyStore, "default")
	if err != nil {
		return nil, err
	}
	chunkStorageOpts = append(chunkStorageOpts, chunk.WithSecret(secret))
	chunkStorage := chunk.NewStorage(objClient, memCache, db, tracker, chunkStorageOpts...)
	d.storage = fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunkStorage, env.FileSetStorageOptions()...)
	// Setup compaction queue and worker.
	d.compactionQueue, err = work.NewTaskQueue(context.Background(), etcdClient, etcdPrefix, storageTaskNamespace)
	if err != nil {
		return nil, err
	}
	d.commitStore = newPostgresCommitStore(db, tracker, d.storage)
	// Create spec repo (default repo)
	repo := client.NewRepo(ppsconsts.SpecRepo)
	repoInfo := &pfs.RepoInfo{
		Repo:    repo,
		Created: types.TimestampNow(),
	}
	if _, err := col.NewSTM(context.Background(), etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		return repos.Create(repo.Name, repoInfo)
	}); err != nil && !col.IsErrExists(err) {
		return nil, err
	}
	// Setup PFS master
	go d.master(env, db)
	go d.compactionWorker()
	return d, nil
}

func (d *driver) activateAuth(txnCtx *txnenv.TransactionContext) error {
	repos := d.repos.ReadOnly(txnCtx.ClientContext)
	repoInfo := &pfs.RepoInfo{}
	return repos.List(repoInfo, col.DefaultOptions, func(repoName string) error {
		err := txnCtx.Auth().CreateRoleBindingInTransaction(txnCtx, "", nil, &auth.Resource{
			Type: auth.ResourceType_REPO,
			Name: repoInfo.Repo.Name,
		})
		if err != nil && !col.IsErrExists(err) {
			return err
		}
		return nil
	})
}

func (d *driver) createRepo(txnCtx *txnenv.TransactionContext, repo *pfs.Repo, description string, update bool) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	// Check that the user is logged in (user doesn't need any access level to
	// create a repo, but they must be authenticated if auth is active)
	whoAmI, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	authIsActivated := !auth.IsErrNotActivated(err)
	if authIsActivated && err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to create a repo)")
	}
	if err := ancestry.ValidateName(repo.Name); err != nil {
		return err
	}

	repos := d.repos.ReadWrite(txnCtx.Stm)

	// check if 'repo' already exists. If so, return that error. Otherwise,
	// proceed with ACL creation (avoids awkward "access denied" error when
	// calling "createRepo" on a repo that already exists)
	var existingRepoInfo pfs.RepoInfo
	err = repos.Get(repo.Name, &existingRepoInfo)
	if err != nil && !col.IsErrNotFound(err) {
		return errors.Wrapf(err, "error checking whether \"%s\" exists", repo.Name)
	} else if err == nil {
		// Existing repo case--just update the repo description.
		if !update {
			return pfsserver.ErrRepoExists{repo}
		}

		if existingRepoInfo.Description == description {
			// Don't overwrite the stored proto with an identical value. This
			// optimization is impactful because pps will frequently update the __spec__
			// repo to make sure it exists.
			return nil
		}

		// Check if the caller is authorized to modify this repo
		// Note, we don't do this before checking if the description changed because
		// there is client code that calls CreateRepo(R, update=true) as an
		// idempotent way to ensure that R exists. By permitting these calls when
		// they don't actually change anything, even if the caller doesn't have
		// WRITER access, we make the pattern more generally useful.
		if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, repo.Name, auth.Permission_REPO_WRITE); err != nil {
			return errors.Wrapf(err, "could not update description of %q", repo)
		}
		existingRepoInfo.Description = description
		return repos.Put(repo.Name, &existingRepoInfo)
	} else {
		// New repo case
		if authIsActivated {
			// Create ACL for new repo. Make caller the sole owner. If the ACL already
			// exists with a different owner, this will fail.
			if err := txnCtx.Auth().CreateRoleBindingInTransaction(txnCtx, whoAmI.Username, []string{auth.RepoOwnerRole}, &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name}); err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not create role binding for new repo \"%s\"", repo.Name)
			}
		}
		return repos.Create(repo.Name, &pfs.RepoInfo{
			Repo:        repo,
			Created:     types.TimestampNow(),
			Description: description,
		})
	}
}

func (d *driver) inspectRepo(txnCtx *txnenv.TransactionContext, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
	// Validate arguments
	if repo == nil {
		return nil, errors.New("repo cannot be nil")
	}

	result := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.Stm).Get(repo.Name, result); err != nil {
		return nil, err
	}
	if includeAuth {
		permissions, err := d.getPermissions(txnCtx.Client, repo)
		if err != nil {
			if auth.IsErrNotActivated(err) {
				return result, nil
			}
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error getting access level for \"%s\"", repo.Name)
		}
		result.AuthInfo = &pfs.RepoAuthInfo{Permissions: permissions}
	}
	return result, nil
}

func (d *driver) getPermissions(pachClient *client.APIClient, repo *pfs.Repo) ([]auth.Permission, error) {
	ctx := pachClient.Ctx()

	resp, err := pachClient.AuthAPIClient.Authorize(ctx, &auth.AuthorizeRequest{
		Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name},
		Permissions: []auth.Permission{
			auth.Permission_REPO_READ,
			auth.Permission_REPO_WRITE,
			auth.Permission_REPO_MODIFY_BINDINGS,
			auth.Permission_REPO_DELETE,
			auth.Permission_REPO_INSPECT_COMMIT,
			auth.Permission_REPO_LIST_COMMIT,
			auth.Permission_REPO_DELETE_COMMIT,
			auth.Permission_REPO_CREATE_BRANCH,
			auth.Permission_REPO_LIST_BRANCH,
			auth.Permission_REPO_DELETE_BRANCH,
			auth.Permission_REPO_LIST_FILE,
			auth.Permission_REPO_INSPECT_FILE,
			auth.Permission_REPO_ADD_PIPELINE_READER,
			auth.Permission_REPO_REMOVE_PIPELINE_READER,
			auth.Permission_REPO_ADD_PIPELINE_WRITER,
		},
	})
	if err != nil {
		return nil, err
	}

	return resp.Satisfied, nil
}

func (d *driver) listRepo(pachClient *client.APIClient, includeAuth bool) (*pfs.ListRepoResponse, error) {
	ctx := pachClient.Ctx()
	repos := d.repos.ReadOnly(ctx)
	result := &pfs.ListRepoResponse{}
	authSeemsActive := true
	repoInfo := &pfs.RepoInfo{}
	if err := repos.List(repoInfo, col.DefaultOptions, func(repoName string) error {
		if repoName == ppsconsts.SpecRepo {
			return nil
		}
		if includeAuth && authSeemsActive {
			permissions, err := d.getPermissions(pachClient, repoInfo.Repo)
			if err == nil {
				repoInfo.AuthInfo = &pfs.RepoAuthInfo{Permissions: permissions}
			} else if auth.IsErrNotActivated(err) {
				authSeemsActive = false
			} else {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "error getting access level for \"%s\"", repoName)
			}
		}
		result.RepoInfo = append(result.RepoInfo, proto.Clone(repoInfo).(*pfs.RepoInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) deleteRepo(txnCtx *txnenv.TransactionContext, repo *pfs.Repo, force bool) error {
	// TODO(msteffen): Fix d.deleteAll() so that it doesn't need to delete and
	// recreate the PPS spec repo, then uncomment this block to prevent users from
	// deleting it and breaking their cluster
	// if repo.Name == ppsconsts.SpecRepo {
	// 	return errors.Errorf("cannot delete the special PPS repo %s", ppsconsts.SpecRepo)
	// }
	repos := d.repos.ReadWrite(txnCtx.Stm)

	// check if 'repo' is already gone. If so, return that error. Otherwise,
	// proceed with auth check (avoids awkward "access denied" error when calling
	// "deleteRepo" on a repo that's already gone)
	var existingRepoInfo pfs.RepoInfo
	err := repos.Get(repo.Name, &existingRepoInfo)
	if err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "error checking whether \"%s\" exists", repo.Name)
		}
	}

	// Check if the caller is authorized to delete this repo
	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, repo.Name, auth.Permission_REPO_DELETE); err != nil {
		return err
	}

	repoInfo := new(pfs.RepoInfo)
	if err := repos.Get(repo.Name, repoInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return errors.Wrapf(err, "repos.Get")
		}
	}

	// make a list of all the commits
	commits := d.commits(repo.Name).ReadOnly(txnCtx.ClientContext)
	commitInfos := make(map[string]*pfs.CommitInfo)
	commitInfo := &pfs.CommitInfo{}
	if err := commits.List(commitInfo, col.DefaultOptions, func(commitID string) error {
		commitInfos[commitID] = proto.Clone(commitInfo).(*pfs.CommitInfo)
		return nil
	}); err != nil {
		return err
	}

	visited := make(map[string]bool) // visitied upstream (provenant) commits
	// and then delete them while making sure that the subvenance of upstream commits gets updated
	for _, ci := range commitInfos {
		// Remove the deleted commit from the upstream commits' subvenance.
		for _, prov := range ci.Provenance {
			// Check if we've fixed prov already (or if it's in this repo and
			// doesn't need to be fixed
			if visited[prov.Commit.ID] || prov.Commit.Repo.Name == repo.Name {
				continue
			}
			// or if the repo has already been deleted
			ri := new(pfs.RepoInfo)
			if err := repos.Get(prov.Commit.Repo.Name, ri); err != nil {
				if !col.IsErrNotFound(err) {
					return errors.Wrapf(err, "repo %v was not found", prov.Commit.Repo.Name)
				}
				continue
			}
			visited[prov.Commit.ID] = true

			// fix prov's subvenance
			provCI := &pfs.CommitInfo{}
			if err := d.commits(prov.Commit.Repo.Name).ReadWrite(txnCtx.Stm).Update(prov.Commit.ID, provCI, func() error {
				subvTo := 0 // copy subvFrom to subvTo, excepting subv ranges to delete (so that they're overwritten)
				for subvFrom, subv := range provCI.Subvenance {
					if subv.Upper.Repo.Name == repo.Name {
						continue
					}
					provCI.Subvenance[subvTo] = provCI.Subvenance[subvFrom]
					subvTo++
				}
				provCI.Subvenance = provCI.Subvenance[:subvTo]
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err fixing subvenance of upstream commit %s/%s", prov.Commit.Repo.Name, prov.Commit.ID)
			}
		}
		// TODO: use DropFilesetsTx
		if err := d.commitStore.DropFilesets(txnCtx.ClientContext, ci.Commit); err != nil {
			return err
		}
	}

	var branchInfos []*pfs.BranchInfo
	for _, branch := range repoInfo.Branches {
		bi, err := d.inspectBranch(txnCtx, branch)
		if err != nil {
			return errors.Wrapf(err, "error inspecting branch %s", branch)
		}
		branchInfos = append(branchInfos, bi)
	}
	// sort ascending provenance
	sort.Slice(branchInfos, func(i, j int) bool { return len(branchInfos[i].Provenance) < len(branchInfos[j].Provenance) })
	for i := range branchInfos {
		// delete branches from most provenance to least, that way if one
		// branch is provenant on another (such as with stats branches) we
		// delete them in the right order.
		branch := branchInfos[len(branchInfos)-1-i].Branch
		if err := d.deleteBranch(txnCtx, branch, force); err != nil {
			return errors.Wrapf(err, "delete branch %s", branch)
		}
	}
	// Despite the fact that we already deleted each branch with
	// deleteBranch, we also do branches.DeleteAll(), this insulates us
	// against certain corruption situations where the RepoInfo doesn't
	// exist in etcd but branches do.
	branches := d.branches(repo.Name).ReadWrite(txnCtx.Stm)
	branches.DeleteAll()
	// Similarly with commits
	commitsX := d.commits(repo.Name).ReadWrite(txnCtx.Stm)
	commitsX.DeleteAll()
	if err := repos.Delete(repo.Name); err != nil && !col.IsErrNotFound(err) {
		return errors.Wrapf(err, "repos.Delete")
	}

	if err := txnCtx.Auth().DeleteRoleBindingInTransaction(txnCtx, &auth.Resource{Type: auth.ResourceType_REPO, Name: repo.Name}); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	return nil
}

// ID can be passed in for transactions, which need to ensure the ID doesn't
// change after the commit ID has been reported to a client.
func (d *driver) startCommit(txnCtx *txnenv.TransactionContext, ID string, parent *pfs.Commit, branch string, provenance []*pfs.CommitProvenance, description string) (*pfs.Commit, error) {
	return d.makeCommit(txnCtx, ID, parent, branch, nil, provenance, description, time.Time{}, time.Time{}, 0)
}

// make commit makes a new commit in 'branch', with the parent 'parent' and the
// direct provenance 'provenance'. Note that
// - 'parent' must not be nil, but the only required field is 'parent.Repo'.
// - 'parent.ID' may be set to "", in which case the parent commit is inferred
//   from 'parent.Repo' and 'branch'.
// - If both 'parent.ID' and 'branch' are set, 'parent.ID' determines the parent
//   commit, but 'branch' is still moved to point at the new commit
// - If neither 'parent.ID' nor 'branch' are set, the new commit will have no
//   parent
// - If only 'parent.ID' is set, and it contains a branch, then the new commit's
//   parent will be the HEAD of that branch, but the branch will not be moved
// TODO: Remove the v1 storage data structures from this function, they are not
// used for now.
func (d *driver) makeCommit(
	txnCtx *txnenv.TransactionContext,
	ID string,
	parent *pfs.Commit,
	branch string,
	origin *pfs.CommitOrigin,
	provenance []*pfs.CommitProvenance,
	description string,
	started time.Time,
	finished time.Time,
	sizeBytes uint64,
) (*pfs.Commit, error) {
	// Validate arguments:
	if parent == nil {
		return nil, errors.Errorf("parent cannot be nil")
	}
	// Check that caller is authorized
	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, parent.Repo.Name, auth.Permission_REPO_WRITE); err != nil {
		return nil, err
	}

	// New commit and commitInfo
	newCommit := &pfs.Commit{
		Repo: parent.Repo,
		ID:   ID,
	}
	if newCommit.ID == "" {
		newCommit.ID = uuid.NewWithoutDashes()
	}
	if origin == nil {
		origin = &pfs.CommitOrigin{Kind: pfs.OriginKind_USER}
	}
	newCommitInfo := &pfs.CommitInfo{
		Commit:      newCommit,
		Origin:      origin,
		Description: description,
	}
	if branch != "" {
		if err := ancestry.ValidateName(branch); err != nil {
			return nil, err
		}
	}

	// check if this is happening in a spout pipeline, and append the correct provenance
	spoutName, ok1 := os.LookupEnv("SPOUT_PIPELINE_NAME")
	spoutCommit, ok2 := os.LookupEnv("PPS_SPEC_COMMIT")
	if ok1 && ok2 {
		log.Infof("Appending provenance for spout: %v %v", spoutName, spoutCommit)
		provenance = append(provenance, client.NewCommitProvenance(ppsconsts.SpecRepo, spoutName, spoutCommit))
	}

	// Set newCommitInfo.Started and possibly newCommitInfo.Finished. Enforce:
	// 1) 'started' and 'newCommitInfo.Started' must be set
	// 2) 'started' <= 'finished'
	switch {
	case started.IsZero() && finished.IsZero():
		// set 'started' to Now() && leave 'finished' unset (open commit)
		started = time.Now()
	case started.IsZero() && !finished.IsZero():
		// set 'started' to 'finished'
		started = finished
	case !started.IsZero() && finished.IsZero():
		if now := time.Now(); now.Before(started) {
			log.Warnf("attempted to start commit at future time %v, resetting start time to now (%v)", started, now)
			started = now // prevent finished < started (if user finishes commit soon)
		}
	case !started.IsZero() && !finished.IsZero():
		if finished.Before(started) {
			log.Warnf("attempted to create commit with finish time %[1]v that is before start time %[2]v, resetting start time to %[1]v", finished, started)
			started = finished // prevent finished < started
		}
	}
	var err error
	newCommitInfo.Started, err = types.TimestampProto(started)
	if err != nil {
		return nil, errors.Wrapf(err, "could not convert 'started' time")
	}
	if !finished.IsZero() {
		newCommitInfo.Finished, err = types.TimestampProto(finished)
		if err != nil {
			return nil, errors.Wrapf(err, "could not convert 'finished' time")
		}
	}

	// create the actual commit in etcd and update the branch + parent/child
	// Clone the parent, as this stm modifies it and might wind up getting
	// run more than once (if there's a conflict.)
	parent = proto.Clone(parent).(*pfs.Commit)
	repos := d.repos.ReadWrite(txnCtx.Stm)
	commits := d.commits(parent.Repo.Name).ReadWrite(txnCtx.Stm)
	branches := d.branches(parent.Repo.Name).ReadWrite(txnCtx.Stm)

	// Check if repo exists
	repoInfo := new(pfs.RepoInfo)
	if err := repos.Get(parent.Repo.Name, repoInfo); err != nil {
		return nil, err
	}

	// create/update 'branch' (if it was set) and set parent.ID (if, in
	// addition, 'parent.ID' was not set)
	key := path.Join
	branchProvMap := make(map[string]bool)
	if branch != "" {
		branchInfo := &pfs.BranchInfo{}
		if err := branches.Upsert(branch, branchInfo, func() error {
			// validate branch
			if parent.ID == "" && branchInfo.Head != nil {
				parent.ID = branchInfo.Head.ID
			}
			// include the branch and its provenance in the branch provenance map
			branchProvMap[key(newCommit.Repo.Name, branch)] = true
			for _, b := range branchInfo.Provenance {
				branchProvMap[key(b.Repo.Name, b.Name)] = true
			}
			if branchInfo.Head != nil {
				headCommitInfo := &pfs.CommitInfo{}
				if err := commits.Get(branchInfo.Head.ID, headCommitInfo); err != nil {
					return err
				}
				for _, prov := range headCommitInfo.Provenance {
					branchProvMap[key(prov.Branch.Repo.Name, prov.Branch.Name)] = true
				}
			}
			// Don't count the __spec__ repo towards the provenance count
			// since spouts will have __spec__ as provenance, but need to accept commits
			provenanceCount := len(branchInfo.Provenance)
			for _, p := range branchInfo.Provenance {
				if p.Repo.Name == ppsconsts.SpecRepo {
					provenanceCount--
					break
				}
			}

			// if 'provenance' includes a spec commit, (note the difference from the
			// prev condition) then it was created by pps and is allowed to be in an
			// output branch
			hasSpec := false
			for _, prov := range provenance {
				if prov.Commit.Repo.Name == ppsconsts.SpecRepo {
					hasSpec = true
				}
			}

			if provenanceCount > 0 && !hasSpec {
				return errors.Errorf("cannot start a commit on an output branch")
			}
			// Point 'branch' at the new commit
			branchInfo.Name = branch // set in case 'branch' is new
			branchInfo.Head = newCommit
			branchInfo.Branch = client.NewBranch(newCommit.Repo.Name, branch)
			return nil
		}); err != nil {
			return nil, err
		}
		// Add branch to repo (see "Update repoInfo" below)
		add(&repoInfo.Branches, branchInfo.Branch)
		// and add the branch to the commit info
		newCommitInfo.Branch = branchInfo.Branch
	}

	if err := d.openCommits.ReadWrite(txnCtx.Stm).Put(newCommit.ID, newCommit); err != nil {
		return nil, err
	}

	// Update repoInfo (potentially with new branch and new size)
	if err := repos.Put(parent.Repo.Name, repoInfo); err != nil {
		return nil, err
	}

	// Set newCommit.ParentCommit (if 'parent' and/or 'branch' was set) and add
	// newCommit to parent's ChildCommits
	if parent.ID != "" {
		// Resolve parent.ID if it's a branch that isn't 'branch' (which can
		// happen if 'branch' is new and diverges from the existing branch in
		// 'parent.ID')
		parentCommitInfo, err := d.resolveCommit(txnCtx.Stm, parent)
		if err != nil {
			return nil, errors.Wrapf(err, "parent commit not found")
		}
		// fail if the parent commit has not been finished
		if parentCommitInfo.Finished == nil {
			return nil, errors.Errorf("parent commit %s@%s has not been finished", parent.Repo.Name, parent.ID)
		}
		if err := commits.Update(parent.ID, parentCommitInfo, func() error {
			newCommitInfo.ParentCommit = parent
			// If we don't know the branch the commit belongs to at this point, assume it is the same as the parent branch
			if newCommitInfo.Branch == nil {
				newCommitInfo.Branch = parentCommitInfo.Branch
			}
			parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
			return nil
		}); err != nil {
			// Note: error is emitted if parent.ID is a missing/invalid branch OR a
			// missing/invalid commit ID
			return nil, errors.Wrapf(err, "could not resolve parent commit \"%s\"", parent.ID)
		}
	}

	// Build newCommit's full provenance. B/c commitInfo.Provenance is a
	// transitive closure, there's no need to search the full provenance graph,
	// just take the union of the immediate parents' (in the 'provenance' arg)
	// commitInfo.Provenance

	// keep track of which branches are represented in the commit provenance
	provBranches := make(map[string]bool)
	for _, prov := range provenance {
		// resolve the provenance
		var err error
		prov, err = d.resolveCommitProvenance(txnCtx.Stm, prov)
		if err != nil {
			return nil, err
		}
		provBranches[key(prov.Branch.Repo.Name, prov.Branch.Name)] = true
	}

	newCommitProv := make(map[string]*pfs.CommitProvenance)
	for _, prov := range provenance {
		provCommitInfo, err := d.resolveCommit(txnCtx.Stm, prov.Commit)
		if err != nil {
			return nil, err
		}
		newCommitProv[prov.Commit.ID] = prov
		provBranches[key(prov.Branch.Repo.Name, prov.Branch.Name)] = true
		for _, provProv := range provCommitInfo.Provenance {
			if _, ok := provBranches[key(provProv.Branch.Repo.Name, provProv.Branch.Name)]; !ok {
				newCommitProv[provProv.Commit.ID] = provProv
				provBranches[key(provProv.Branch.Repo.Name, provProv.Branch.Name)] = true
			}
		}
	}

	// keep track of which branches are represented in the commit provenance
	provenantBranches := make(map[string]bool)
	// Copy newCommitProv into newCommitInfo.Provenance, and update upstream subv
	for _, prov := range newCommitProv {
		// resolve the provenance
		var err error
		prov, err = d.resolveCommitProvenance(txnCtx.Stm, prov)
		if err != nil {
			return nil, err
		}
		// there should only be one representative of each branch in the commit provenance
		if _, ok := provenantBranches[key(prov.Branch.Repo.Name, prov.Branch.Name)]; ok {
			return nil, errors.Errorf("the commit provenance contains multiple commits from the same branch")
		}
		provenantBranches[key(prov.Branch.Repo.Name, prov.Branch.Name)] = true

		// ensure the commit provenance is consistent with the branch provenance
		if len(branchProvMap) != 0 {
			// the check for empty branch names is for the run pipeline case in which a commit with no branch are expected in the stats commit provenance
			if prov.Branch.Repo.Name != ppsconsts.SpecRepo && prov.Branch.Name != "" && !branchProvMap[key(prov.Branch.Repo.Name, prov.Branch.Name)] {
				return nil, errors.Errorf("the commit provenance contains a branch which the branch is not provenant on")
			}
		}

		newCommitInfo.Provenance = append(newCommitInfo.Provenance, prov)
		provCommitInfo := &pfs.CommitInfo{}
		if err := d.commits(prov.Commit.Repo.Name).ReadWrite(txnCtx.Stm).Update(prov.Commit.ID, provCommitInfo, func() error {
			d.appendSubvenance(provCommitInfo, newCommitInfo)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// this isn't necessary here, but is done for consistency with commits created by propagateCommits
	// it ensures that the last commit to appear from a specific branch is the latest one on that branch
	var sortErr error
	sort.SliceStable(newCommitInfo.Provenance, func(i, j int) bool {
		// to make sure the parent relationship is respected during sort, we need to make sure that we organize
		// the provenance by repo name and branch name
		if newCommitInfo.Provenance[i].Commit.Repo.Name != newCommitInfo.Provenance[j].Commit.Repo.Name {
			return newCommitInfo.Provenance[i].Commit.Repo.Name < newCommitInfo.Provenance[j].Commit.Repo.Name
		} else if newCommitInfo.Provenance[i].Branch.Name != newCommitInfo.Provenance[j].Branch.Name {
			return newCommitInfo.Provenance[i].Branch.Name < newCommitInfo.Provenance[j].Branch.Name
		}

		// we need to check the commit info of the 'j' provenance commit to get the parent
		provCommitInfo, err := d.resolveCommit(txnCtx.Stm, newCommitInfo.Provenance[j].Commit)
		if err != nil {
			// capture error
			sortErr = err
			return true
		}
		// the parent commit of 'j' should precede it
		if provCommitInfo.ParentCommit != nil &&
			newCommitInfo.Provenance[i].Commit.ID == provCommitInfo.ParentCommit.ID {
			return true
		}
		return false
	})
	// capture any errors during sorting
	if sortErr != nil {
		return nil, sortErr
	}

	// Finally, create the commit
	if err := commits.Create(newCommit.ID, newCommitInfo); err != nil {
		return nil, err
	}
	// Defer propagation of the commit until the end of the transaction so we can
	// batch downstream commits together if there are multiple changes.
	if branch != "" {
		var triggeredBranches []*pfs.Branch
		if newCommitInfo.Finished != nil {
			triggeredBranches, err = d.triggerCommit(txnCtx, newCommit)
			if err != nil {
				return nil, err
			}
		}
		for _, b := range append(triggeredBranches, client.NewBranch(newCommit.Repo.Name, branch)) {
			if err := txnCtx.PropagateCommit(b, true); err != nil {
				return nil, err
			}
		}
	}
	return newCommit, nil
}

// resolveCommitProvenance resolves a user 'commit' (which may be a commit ID or
// branch reference) to a commit + branch pair interpreted as commit provenance.
// If a complete commit provenance is passed in it just uses that.
// It accepts an STM so that it can be used in a transaction and avoids an
// inconsistent call to d.inspectCommit()
func (d *driver) resolveCommitProvenance(stm col.STM, userCommitProvenance *pfs.CommitProvenance) (*pfs.CommitProvenance, error) {
	if userCommitProvenance == nil {
		return nil, errors.Errorf("cannot resolve nil commit provenance")
	}
	// resolve the commit in case the commit is actually a branch name
	userCommitProvInfo, err := d.resolveCommit(stm, userCommitProvenance.Commit)
	if err != nil {
		return nil, err
	}
	if userCommitProvenance.Branch == nil || userCommitProvenance.Branch.Name == "" {
		// if the branch isn't specified, default to using the commit's branch (as long as that isn't nil too)
		if userCommitProvInfo.Branch != nil {
			userCommitProvenance.Branch = userCommitProvInfo.Branch
		}

		// but if the original "commit id" was a branch name, use that as the branch instead
		if userCommitProvInfo.Commit.ID != userCommitProvenance.Commit.ID {
			userCommitProvenance.Branch.Name = userCommitProvenance.Commit.ID
			userCommitProvenance.Commit = userCommitProvInfo.Commit
		}
	}
	return userCommitProvenance, nil
}

func (d *driver) finishCommit(txnCtx *txnenv.TransactionContext, commit *pfs.Commit, description string) error {
	ctx := txnCtx.Client.Ctx()
	commitInfo, err := d.resolveCommit(txnCtx.Stm, commit)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{commitInfo.Commit}
	}
	commit = commitInfo.Commit
	if description != "" {
		commitInfo.Description = description
	}
	var parentFSID *fileset.ID
	if commitInfo.ParentCommit != nil {
		id, err := d.commitStore.GetTotalFileset(ctx, commitInfo.ParentCommit)
		if err != nil {
			return err
		}
		parentFSID = id
	}
	// Run compaction task.
	return d.compactionQueue.RunTaskBlock(ctx, func(m *work.Master) error {
		id, err := d.commitStore.GetDiffFileset(ctx, commit)
		if err != nil {
			return err
		}
		var ids []fileset.ID
		// if the commit has a parent, then include the parents fileset in the compaction
		if parentFSID != nil {
			ids = append(ids, *parentFSID)
		}
		ids = append(ids, *id)
		compactedID, err := d.compact(m, ids)
		if err != nil {
			return err
		}
		if err := d.commitStore.SetTotalFileset(ctx, commit, *compactedID); err != nil {
			return err
		}
		outputSize, err := d.storage.SizeOf(ctx, *compactedID)
		if err != nil {
			return err
		}
		commitInfo.SizeBytes = uint64(outputSize)
		commitInfo.Finished = types.TimestampNow()
		empty := strings.Contains(commitInfo.Description, pfs.EmptyStr)
		if err := d.updateProvenanceProgress(txnCtx, !empty, commitInfo); err != nil {
			return err
		}
		if err := d.writeFinishedCommit(txnCtx.Stm, commit, commitInfo); err != nil {
			return err
		}
		triggeredBranches, err := d.triggerCommit(txnCtx, commitInfo.Commit)
		if err != nil {
			return err
		}
		for _, b := range triggeredBranches {
			if err := txnCtx.PropagateCommit(b, false); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *driver) updateProvenanceProgress(txnCtx *txnenv.TransactionContext, success bool, ci *pfs.CommitInfo) error {
	if d.env.DisableCommitProgressCounter {
		return nil
	}
	for _, provC := range ci.Provenance {
		provCi := &pfs.CommitInfo{}
		if err := d.commits(provC.Commit.Repo.Name).ReadWrite(txnCtx.Stm).Update(provC.Commit.ID, provCi, func() error {
			if success {
				provCi.SubvenantCommitsSuccess++
			} else {
				provCi.SubvenantCommitsFailure++
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// writeFinishedCommit writes these changes to etcd:
// 1) it closes the input commit (i.e., it writes any changes made to it and
//    removes it from the open commits)
// 2) if the commit is the new HEAD of master, it updates the repo size
func (d *driver) writeFinishedCommit(stm col.STM, commit *pfs.Commit, commitInfo *pfs.CommitInfo) error {
	commits := d.commits(commit.Repo.Name).ReadWrite(stm)
	if err := commits.Put(commit.ID, commitInfo); err != nil {
		return err
	}
	if err := d.openCommits.ReadWrite(stm).Delete(commit.ID); err != nil {
		return errors.Wrapf(err, "could not confirm that commit %s is open; this is likely a bug", commit.ID)
	}
	// update the repo size if this is the head of master
	repos := d.repos.ReadWrite(stm)
	repoInfo := new(pfs.RepoInfo)
	if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
		return err
	}
	for _, branch := range repoInfo.Branches {
		if branch.Name == "master" {
			branchInfo := &pfs.BranchInfo{}
			if err := d.branches(commit.Repo.Name).ReadWrite(stm).Get(branch.Name, branchInfo); err != nil {
				return err
			}
			// If the head commit of master has been deleted, we could get here if another branch
			// had shared its head commit with master, and then we created a new commit on that branch
			if branchInfo.Head != nil && branchInfo.Head.ID == commit.ID {
				repoInfo.SizeBytes = commitInfo.SizeBytes
				if err := repos.Put(commit.Repo.Name, repoInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// propagateCommits selectively starts commits in or downstream of 'branches' in
// order to restore the invariant that branch provenance matches HEAD commit
// provenance:
//   B.Head is provenant on A.Head <=>
//   branch B is provenant on branch A and A.Head != nil
// The implementation assumes that the invariant already holds for all branches
// upstream of 'branches', but not necessarily for each 'branch' itself. Despite
// the name, 'branches' do not need a HEAD commit to propagate, though one may be
// created.
//
// In other words, propagateCommits scans all branches b_downstream that are
// equal to or downstream of 'branches', and if the HEAD of b_downstream isn't
// provenant on the HEADs of b_downstream's provenance, propagateCommits starts
// a new HEAD commit in b_downstream that is. For example, propagateCommits
// starts downstream output commits (which trigger PPS jobs) when new input
// commits arrive on 'branch', when 'branches's HEAD is deleted, or when
// 'branches' are newly created (i.e. in CreatePipeline).
//
// The isNewCommit flag indicates whether propagateCommits was called during the creation of a new commit.
func (d *driver) propagateCommits(stm col.STM, branches []*pfs.Branch, isNewCommit bool) error {
	key := path.Join
	// subvBIMap = ( ⋃{b.subvenance | b ∈ branches} ) ∪ branches
	subvBIMap := map[string]*pfs.BranchInfo{}
	for _, branch := range branches {
		branchInfo, ok := subvBIMap[key(branch.Repo.Name, branch.Name)]
		if !ok {
			branchInfo = &pfs.BranchInfo{}
			if err := d.branches(branch.Repo.Name).ReadWrite(stm).Get(branch.Name, branchInfo); err != nil {
				return err
			}
			subvBIMap[key(branch.Repo.Name, branch.Name)] = branchInfo
		}
		for _, subvBranch := range branchInfo.Subvenance {
			_, ok := subvBIMap[key(subvBranch.Repo.Name, subvBranch.Name)]
			if !ok {
				subvInfo := &pfs.BranchInfo{}
				if err := d.branches(subvBranch.Repo.Name).ReadWrite(stm).Get(subvBranch.Name, subvInfo); err != nil {
					return err
				}
				subvBIMap[key(subvBranch.Repo.Name, subvBranch.Name)] = subvInfo
			}
		}
	}

	// 'subvBIs' is the collection of downstream branches that may get a new
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

	// Iterate through downstream branches and determine which need a new commit.
nextSubvBI:
	for _, subvBI := range subvBIs {
		subvB := subvBI.Branch
		stmCommits := d.commits(subvB.Repo.Name).ReadWrite(stm)
		stmBranches := d.branches(subvB.Repo.Name).ReadWrite(stm)

		// Compute the full provenance of hypothetical new output commit to decide
		// if we need it.
		newCommitProvMap := make(map[string]*pfs.CommitProvenance)
		for _, provOfSubvB := range subvBI.Provenance {
			// get the branch info from the provenance branch
			provOfSubvBI := &pfs.BranchInfo{}
			if err := d.branches(provOfSubvB.Repo.Name).ReadWrite(stm).Get(provOfSubvB.Name, provOfSubvBI); err != nil && !col.IsErrNotFound(err) {
				return errors.Wrapf(err, "could not read branch %s/%s", provOfSubvB.Repo.Name, provOfSubvB.Name)
			}
			// if provOfSubvB has no head commit, then it doesn't contribute to newCommit
			if provOfSubvBI.Head == nil {
				continue
			}
			// - Add provOfSubvBI.Head to the new commit's provenance
			// - Since we want the new commit's provenance to be a transitive closure,
			//   we add provOfSubvBI.Head's *provenance* to newCommit's provenance.
			//   - Note: In most cases, every commit in there will be the Head of some
			//     other provOfSubvBI, but not when e.g. deferred downstream
			//     processing, where an upstream branch has no branch provenance but
			//     its head commit has commit provenance.
			// - We need to key on both the commit id and the branch name, so that
			//   branches with a shared commit are both represented in the provenance
			newCommitProvMap[key(provOfSubvBI.Head.ID, provOfSubvB.Name)] = &pfs.CommitProvenance{
				Commit: provOfSubvBI.Head,
				Branch: provOfSubvB,
			}
			provOfSubvBHeadInfo := &pfs.CommitInfo{}
			if err := d.commits(provOfSubvB.Repo.Name).ReadWrite(stm).Get(provOfSubvBI.Head.ID, provOfSubvBHeadInfo); err != nil {
				return err
			}
			for _, provProv := range provOfSubvBHeadInfo.Provenance {
				newProvProv, err := d.resolveCommitProvenance(stm, provProv)
				if err != nil {
					return errors.Wrapf(err, "could not resolve provenant commit %s@%s (%s)",
						provProv.Commit.Repo.Name, provProv.Commit.ID, provProv.Branch.Name)
				}
				provProv = newProvProv
				newCommitProvMap[key(provProv.Commit.ID, provProv.Branch.Name)] = provProv
			}
		}
		if len(newCommitProvMap) == 0 {
			// no input commits to process; don't create a new output commit
			continue nextSubvBI
		}

		// 'subvB' may already have a HEAD commit, so compute whether the new output
		// commit's provenance would be a subset of the existing HEAD commit's
		// provenance. If so, a new output commit would be a duplicate, so don't
		// create it.
		if subvBI.Head != nil {
			// get the info for subvB's HEAD commit
			subvBHeadInfo := &pfs.CommitInfo{}
			if err := stmCommits.Get(subvBI.Head.ID, subvBHeadInfo); err != nil {
				return pfsserver.ErrCommitNotFound{subvBI.Head}
			}
			provIntersection := make(map[string]struct{})
			for _, p := range subvBHeadInfo.Provenance {
				if _, ok := newCommitProvMap[key(p.Commit.ID, p.Branch.Name)]; ok {
					provIntersection[key(p.Commit.ID, p.Branch.Name)] = struct{}{}
				}
			}
			if len(newCommitProvMap) == len(provIntersection) {
				// newCommit's provenance is subset of existing HEAD's provenance
				continue nextSubvBI
			}
		}

		// If the only branches in the hypothetical output commit's provenance are
		// in the 'spec' repo, creating it would mean creating a confusing "dummy"
		// job with no non-spec input data. If this is the case, don't create a new
		// output commit
		allSpec := true
		for _, p := range newCommitProvMap {
			if p.Branch.Repo.Name != ppsconsts.SpecRepo {
				allSpec = false
				break
			}
		}
		if allSpec {
			// Only input data is PipelineInfo; don't create new output commit
			continue nextSubvBI
		}

		// if a commit was just created and this is the same branch as the one being
		// propagated, we don't need to do anything
		if isNewCommit && len(branches) == 1 &&
			branches[0].Repo.Name == subvB.Repo.Name && branches[0].Name == subvB.Name {
			continue nextSubvBI
		}

		// *All checks passed* start a new output commit in 'subvB'
		newCommit := &pfs.Commit{
			Repo: subvB.Repo,
			ID:   uuid.NewWithoutDashes(),
		}
		newCommitInfo := &pfs.CommitInfo{
			Commit:  newCommit,
			Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
			Started: types.TimestampNow(),
		}

		// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
		newCommitInfo.ParentCommit = subvBI.Head
		if subvBI.Head != nil {
			parentCommitInfo := &pfs.CommitInfo{}
			if err := stmCommits.Update(newCommitInfo.ParentCommit.ID, parentCommitInfo, func() error {
				parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
				return nil
			}); err != nil {
				return err
			}
		}
		subvBI.Head = newCommit
		newCommitInfo.Branch = subvB
		if err := stmBranches.Put(subvB.Name, subvBI); err != nil {
			return err
		}

		// Set provenance and upstream subvenance (appendSubvenance needs
		// newCommitInfo.ParentCommit to extend the correct subvenance range)
		for _, prov := range newCommitProvMap {
			// set provenance of 'newCommit'
			newCommitInfo.Provenance = append(newCommitInfo.Provenance, prov)

			// update subvenance of 'prov'
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(prov.Commit.Repo.Name).ReadWrite(stm).Update(prov.Commit.ID, provCommitInfo, func() error {
				d.appendSubvenance(provCommitInfo, newCommitInfo)
				return nil
			}); err != nil {
				return err
			}
		}

		// this ensures that the job's output commit uses the latest commit on the branch, by ensuring it is the
		// last commit to appear in the provenance slice
		var sortErr error
		sort.SliceStable(newCommitInfo.Provenance, func(i, j int) bool {
			// to make sure the parent relationship is respected during sort, we need to make sure that we organize
			// the provenance by repo name and branch name
			if newCommitInfo.Provenance[i].Commit.Repo.Name != newCommitInfo.Provenance[j].Commit.Repo.Name {
				return newCommitInfo.Provenance[i].Commit.Repo.Name < newCommitInfo.Provenance[j].Commit.Repo.Name
			} else if newCommitInfo.Provenance[i].Branch.Name != newCommitInfo.Provenance[j].Branch.Name {
				return newCommitInfo.Provenance[i].Branch.Name < newCommitInfo.Provenance[j].Branch.Name
			}

			// we need to check the commit info of the 'j' provenance commit to get the parent
			provCommitInfo, err := d.resolveCommit(stm, newCommitInfo.Provenance[j].Commit)
			if err != nil {
				// capture error
				sortErr = err
				return true
			}
			// the parent commit of 'j' should precede it
			if provCommitInfo.ParentCommit != nil &&
				newCommitInfo.Provenance[i].Commit.ID == provCommitInfo.ParentCommit.ID {
				return true
			}
			return false
		})
		// capture any errors during sorting
		if sortErr != nil {
			return sortErr
		}

		// finally create open 'commit'
		if err := stmCommits.Create(newCommit.ID, newCommitInfo); err != nil {
			return err
		}
		if err := d.openCommits.ReadWrite(stm).Put(newCommit.ID, newCommit); err != nil {
			return err
		}
	}
	return nil
}

// inspectCommit takes a Commit and returns the corresponding CommitInfo.
//
// As a side effect, this function also replaces the ID in the given commit
// with a real commit ID.
func (d *driver) inspectCommit(pachClient *client.APIClient, commit *pfs.Commit, blockState pfs.CommitState) (*pfs.CommitInfo, error) {
	if commit.GetRepo().GetName() == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "FileSet - Virtual Commit",
			Finished:    &types.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return cinfo, nil
	}
	ctx := pachClient.Ctx()
	if commit == nil {
		return nil, errors.Errorf("cannot inspect nil commit")
	}
	if err := authserver.CheckRepoIsAuthorized(pachClient, commit.Repo.Name, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
		return nil, err
	}

	// Check if the commitID is a branch name
	var commitInfo *pfs.CommitInfo
	if err := col.NewDryrunSTM(ctx, d.etcdClient, func(stm col.STM) error {
		var err error
		commitInfo, err = d.resolveCommit(stm, commit)
		return err
	}); err != nil {
		return nil, err
	}

	commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
	if blockState == pfs.CommitState_READY {
		// Wait for each provenant commit to be finished
		for _, p := range commitInfo.Provenance {
			d.inspectCommit(pachClient, p.Commit, pfs.CommitState_FINISHED)
		}
	}
	if blockState == pfs.CommitState_FINISHED {
		// Watch the CommitInfo until the commit has been finished
		if err := func() error {
			commitInfoWatcher, err := commits.WatchOne(commit.ID)
			if err != nil {
				return err
			}
			defer commitInfoWatcher.Close()
			for {
				var commitID string
				_commitInfo := new(pfs.CommitInfo)
				event := <-commitInfoWatcher.Watch()
				switch event.Type {
				case watch.EventError:
					return event.Err
				case watch.EventPut:
					if err := event.Unmarshal(&commitID, _commitInfo); err != nil {
						return errors.Wrapf(err, "unmarshal")
					}
				case watch.EventDelete:
					return pfsserver.ErrCommitDeleted{commit}
				}
				if _commitInfo.Finished != nil {
					commitInfo = _commitInfo
					break
				}
			}
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	return commitInfo, nil
}

// resolveCommit contains the essential implementation of inspectCommit: it converts 'commit' (which may
// be a commit ID or branch reference, plus '~' and/or '^') to a repo + commit
// ID. It accepts an STM so that it can be used in a transaction and avoids an
// inconsistent call to d.inspectCommit()
func (d *driver) resolveCommit(stm col.STM, userCommit *pfs.Commit) (*pfs.CommitInfo, error) {
	if userCommit == nil {
		return nil, errors.Errorf("cannot resolve nil commit")
	}
	if userCommit.Repo == nil {
		return nil, errors.Errorf("cannot resolve commit with no repo")
	}
	if userCommit.ID == "" {
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

	// Keep track of the commit branch, in case it isn't set in the commitInfo already
	var commitBranch *pfs.Branch
	// Check if commit.ID is already a commit ID (i.e. a UUID).
	if !uuid.IsUUIDWithoutDashes(commit.ID) {
		branches := d.branches(commit.Repo.Name).ReadWrite(stm)
		branchInfo := &pfs.BranchInfo{}
		// See if we are given a branch
		if err := branches.Get(commit.ID, branchInfo); err != nil {
			return nil, err
		}
		if branchInfo.Head == nil {
			return nil, pfsserver.ErrNoHead{branchInfo.Branch}
		}
		commitBranch = branchInfo.Branch
		commit.ID = branchInfo.Head.ID
	}

	// Traverse commits' parents until you've reached the right ancestor
	commits := d.commits(commit.Repo.Name).ReadWrite(stm)
	commitInfo := &pfs.CommitInfo{}
	if ancestryLength >= 0 {
		for i := 0; i <= ancestryLength; i++ {
			if commit == nil {
				return nil, pfsserver.ErrCommitNotFound{userCommit}
			}
			if err := commits.Get(commit.ID, commitInfo); err != nil {
				if col.IsErrNotFound(err) {
					if i == 0 {
						return nil, pfsserver.ErrCommitNotFound{userCommit}
					}
					return nil, pfsserver.ErrParentCommitNotFound{commit}
				}
				return nil, err
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
				return nil, pfsserver.ErrCommitNotFound{userCommit}
			}
			if err := commits.Get(commit.ID, &cis[i%len(cis)]); err != nil {
				if col.IsErrNotFound(err) {
					if i == 0 {
						return nil, pfsserver.ErrCommitNotFound{userCommit}
					}
					return nil, pfsserver.ErrParentCommitNotFound{commit}
				}
			}
			commit = cis[i%len(cis)].ParentCommit
		}
	}
	if commitInfo.Branch == nil {
		commitInfo.Branch = commitBranch
	}
	userCommit.ID = commitInfo.Commit.ID
	return commitInfo, nil
}

func (d *driver) listCommit(pachClient *client.APIClient, repo *pfs.Repo, to *pfs.Commit, from *pfs.Commit, number uint64, reverse bool, cb func(*pfs.CommitInfo) error) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	ctx := pachClient.Ctx()
	if err := authserver.CheckRepoIsAuthorized(pachClient, repo.Name, auth.Permission_REPO_LIST_COMMIT); err != nil {
		return err
	}
	if from != nil && from.Repo.Name != repo.Name || to != nil && to.Repo.Name != repo.Name {
		return errors.Errorf("`from` and `to` commits need to be from repo %s", repo.Name)
	}

	// Make sure that the repo exists
	if repo.Name != "" {
		err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
			_, err := d.inspectRepo(txnCtx, repo, !includeAuth)
			return err
		})
		if err != nil {
			return err
		}
	}

	// Make sure that both from and to are valid commits
	if from != nil {
		_, err := d.inspectCommit(pachClient, from, pfs.CommitState_STARTED)
		if err != nil {
			return err
		}
	}
	if to != nil {
		_, err := d.inspectCommit(pachClient, to, pfs.CommitState_STARTED)
		if err != nil {
			if isNoHeadErr(err) {
				return nil
			}
			return err
		}
	}

	// if number is 0, we return all commits that match the criteria
	if number == 0 {
		number = math.MaxUint64
	}
	commits := d.commits(repo.Name).ReadOnly(ctx)
	ci := &pfs.CommitInfo{}

	if from != nil && to == nil {
		return errors.Errorf("cannot use `from` commit without `to` commit")
	} else if from == nil && to == nil {
		// if neither from and to is given, we list all commits in
		// the repo, sorted by revision timestamp (or reversed if so requested.)
		opts := *col.DefaultOptions // Note we dereference here so as to make a copy
		if reverse {
			opts.Order = etcd.SortAscend
		}
		// we hold onto a revisions worth of cis so that we can sort them by provenance
		var cis []*pfs.CommitInfo
		// sendCis sorts cis and passes them to f
		sendCis := func() error {
			// Sort in reverse provenance order, i.e. commits come before their provenance
			sort.Slice(cis, func(i, j int) bool { return len(cis[i].Provenance) > len(cis[j].Provenance) })
			for i, ci := range cis {
				if number == 0 {
					return errutil.ErrBreak
				}
				number--

				if reverse {
					ci = cis[len(cis)-1-i]
				}
				if err := cb(ci); err != nil {
					return err
				}
			}
			cis = nil
			return nil
		}
		lastRev := int64(-1)
		if err := commits.ListRev(ci, &opts, func(commitID string, createRev int64) error {
			if createRev != lastRev {
				if err := sendCis(); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
				lastRev = createRev
			}
			cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
			return nil
		}); err != nil {
			return err
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
			var commitInfo pfs.CommitInfo
			if err := commits.Get(cursor.ID, &commitInfo); err != nil {
				return err
			}
			if err := cb(&commitInfo); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
			cursor = commitInfo.ParentCommit
			number--
		}
	}
	return nil
}

func (d *driver) squashCommit(txnCtx *txnenv.TransactionContext, userCommit *pfs.Commit) error {
	// Main txn: Delete all downstream commits, and update subvenance of upstream commits
	// TODO update branches inside this txn, by storing a repo's branches in its
	// RepoInfo or its HEAD commit
	deleted := make(map[string]*pfs.CommitInfo) // deleted commits
	affectedRepos := make(map[string]struct{})  // repos containing deleted commits

	// 1) re-read CommitInfo inside txn
	userCommitInfo, err := d.resolveCommit(txnCtx.Stm, userCommit)
	if err != nil {
		return errors.Wrapf(err, "resolveCommit")
	}

	// 2) Define helper for deleting commits. 'lower' corresponds to
	// pfs.CommitRange.Lower, and is an ancestor of 'upper'
	deleteCommit := func(lower, upper *pfs.Commit) error {
		// Validate arguments
		if lower.Repo.Name != upper.Repo.Name {
			return errors.Errorf("cannot delete commit range with mismatched repos \"%s\" and \"%s\"", lower.Repo.Name, upper.Repo.Name)
		}
		affectedRepos[lower.Repo.Name] = struct{}{}
		commits := d.commits(lower.Repo.Name).ReadWrite(txnCtx.Stm)

		// delete commits on path upper -> ... -> lower (traverse ParentCommits)
		commit := upper
		for {
			if commit == nil {
				return errors.Errorf("encountered nil parent commit in %s/%s...%s", lower.Repo.Name, lower.ID, upper.ID)
			}
			// Store commitInfo in 'deleted' and remove commit from etcd
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Get(commit.ID, commitInfo); err != nil {
				return err
			}
			// If a commit has already been deleted, we don't want to overwrite the existing information, since commitInfo will be nil
			if _, ok := deleted[commit.ID]; !ok {
				deleted[commit.ID] = commitInfo
			}
			if err := commits.Delete(commit.ID); err != nil {
				return err
			}
			// Delete the commit's filesets
			if err := d.commitStore.DropFilesets(txnCtx.Client.Ctx(), commit); err != nil {
				return err
			}
			if commit.ID == lower.ID {
				break // check after deletion so we delete 'lower' (inclusive range)
			}
			commit = commitInfo.ParentCommit
		}

		return nil
	}

	// 3) Validate the commit (check that it has no provenance) and delete it
	if provenantOnInput(userCommitInfo.Provenance) {
		return errors.Errorf("cannot delete the commit \"%s/%s\" because it has non-empty provenance", userCommit.Repo.Name, userCommit.ID)
	}
	deleteCommit(userCommitInfo.Commit, userCommitInfo.Commit)

	// 4) Delete all of the downstream commits of 'commit'
	for _, subv := range userCommitInfo.Subvenance {
		deleteCommit(subv.Lower, subv.Upper)
	}

	// 5) Remove the commits in 'deleted' from all remaining upstream commits'
	// subvenance.
	// While 'commit' is required to be an input commit (no provenance),
	// downstream commits from 'commit' may have multiple inputs, and those
	// other inputs must have their subvenance updated
	visited := make(map[string]bool) // visitied upstream (provenant) commits
	for _, deletedInfo := range deleted {
		for _, prov := range deletedInfo.Provenance {
			// Check if we've fixed provCommit already (or if it's deleted and
			// doesn't need to be fixed
			if _, isDeleted := deleted[prov.Commit.ID]; isDeleted || visited[prov.Commit.ID] {
				continue
			}
			visited[prov.Commit.ID] = true

			// fix provCommit's subvenance
			provCI := &pfs.CommitInfo{}
			if err := d.commits(prov.Commit.Repo.Name).ReadWrite(txnCtx.Stm).Update(prov.Commit.ID, provCI, func() error {
				subvTo := 0 // copy subvFrom to subvTo, excepting subv ranges to delete (so that they're overwritten)
			nextSubvRange:
				for subvFrom, subv := range provCI.Subvenance {
					// Compute path (of commit IDs) connecting subv.Upper to subv.Lower
					cur := subv.Upper.ID
					path := []string{cur}
					for cur != subv.Lower.ID {
						// Get CommitInfo for 'cur' (either in 'deleted' or from etcd)
						// and traverse parent
						curInfo, ok := deleted[cur]
						if !ok {
							curInfo = &pfs.CommitInfo{}
							if err := d.commits(subv.Lower.Repo.Name).ReadWrite(txnCtx.Stm).Get(cur, curInfo); err != nil {
								return errors.Wrapf(err, "error reading commitInfo for subvenant \"%s/%s\"", subv.Lower.Repo.Name, cur)
							}
						}
						if curInfo.ParentCommit == nil {
							break
						}
						cur = curInfo.ParentCommit.ID
						path = append(path, cur)
					}

					// move 'subv.Upper' through parents until it points to a non-deleted commit
					for j := range path {
						if _, ok := deleted[subv.Upper.ID]; !ok {
							break
						}
						if j+1 >= len(path) {
							// All commits in subvRange are deleted. Remove entire Range
							// from provCI.Subvenance
							continue nextSubvRange
						}
						subv.Upper.ID = path[j+1]
					}

					// move 'subv.Lower' through children until it points to a non-deleted commit
					for j := len(path) - 1; j >= 0; j-- {
						if _, ok := deleted[subv.Lower.ID]; !ok {
							break
						}
						// We'll eventually get to a non-deleted commit because the
						// 'upper' block didn't exit
						subv.Lower.ID = path[j-1]
					}
					provCI.Subvenance[subvTo] = provCI.Subvenance[subvFrom]
					subvTo++
				}
				provCI.Subvenance = provCI.Subvenance[:subvTo]
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err fixing subvenance of upstream commit %s/%s", prov.Commit.Repo.Name, prov.Commit.ID)
			}
		}
	}

	// 6) Rewrite ParentCommit of deleted commits' children, and
	// ChildCommits of deleted commits' parents
	visited = make(map[string]bool) // visited child/parent commits
	for deletedID, deletedInfo := range deleted {
		if visited[deletedID] {
			continue
		}

		// Traverse downwards until we find the lowest (most ancestral)
		// non-nil, deleted commit
		lowestCommitInfo := deletedInfo
		for {
			if lowestCommitInfo.ParentCommit == nil {
				break // parent is nil
			}
			parentInfo, ok := deleted[lowestCommitInfo.ParentCommit.ID]
			if !ok {
				break // parent is not deleted
			}
			lowestCommitInfo = parentInfo // parent exists and is deleted--go down
		}

		// BFS upwards through graph for all non-deleted children
		var next *pfs.Commit                            // next vertex to search
		queue := []*pfs.Commit{lowestCommitInfo.Commit} // queue of vertices to explore
		liveChildren := make(map[string]struct{})       // live children discovered so far
		for len(queue) > 0 {
			next, queue = queue[0], queue[1:]
			if visited[next.ID] {
				continue
			}
			visited[next.ID] = true
			nextInfo, ok := deleted[next.ID]
			if !ok {
				liveChildren[next.ID] = struct{}{}
				continue
			}
			queue = append(queue, nextInfo.ChildCommits...)
		}

		// Point all non-deleted children at the first valid parent (or nil),
		// and point first non-deleted parent at all non-deleted children
		commits := d.commits(deletedInfo.Commit.Repo.Name).ReadWrite(txnCtx.Stm)
		parent := lowestCommitInfo.ParentCommit
		for child := range liveChildren {
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Update(child, commitInfo, func() error {
				commitInfo.ParentCommit = parent
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err updating child commit %v", lowestCommitInfo.Commit)
			}
		}
		if parent != nil {
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Update(parent.ID, commitInfo, func() error {
				// Add existing live commits in commitInfo.ChildCommits to the
				// live children above lowestCommitInfo, then put them all in
				// 'parent'
				for _, child := range commitInfo.ChildCommits {
					if _, ok := deleted[child.ID]; ok {
						continue
					}
					liveChildren[child.ID] = struct{}{}
				}
				commitInfo.ChildCommits = make([]*pfs.Commit, 0, len(liveChildren))
				for child := range liveChildren {
					commitInfo.ChildCommits = append(commitInfo.ChildCommits, client.NewCommit(parent.Repo.Name, child))
				}
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err rewriting children of ancestor commit %v", lowestCommitInfo.Commit)
			}
		}
	}

	// 7) Traverse affected repos and rewrite all branches so that no branch
	// points to a deleted commit
	var affectedBranches []*pfs.BranchInfo
	repos := d.repos.ReadWrite(txnCtx.Stm)
	for repo := range affectedRepos {
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(repo, repoInfo); err != nil {
			return err
		}
		for _, brokenBranch := range repoInfo.Branches {
			// Traverse HEAD commit until we find a non-deleted parent or nil;
			// rewrite branch
			var branchInfo pfs.BranchInfo
			if err := d.branches(brokenBranch.Repo.Name).ReadWrite(txnCtx.Stm).Update(brokenBranch.Name, &branchInfo, func() error {
				prevHead := branchInfo.Head
				for {
					if branchInfo.Head == nil {
						return nil // no commits left in branch
					}
					headCommitInfo, headIsDeleted := deleted[branchInfo.Head.ID]
					if !headIsDeleted {
						break
					}
					branchInfo.Head = headCommitInfo.ParentCommit
				}
				if prevHead != nil && prevHead.ID != branchInfo.Head.ID {
					affectedBranches = append(affectedBranches, &branchInfo)
				}
				return err
			}); err != nil && !col.IsErrNotFound(err) {
				// If err is NotFound, branch is in downstream provenance but
				// doesn't exist yet--nothing to update
				return errors.Wrapf(err, "error updating branch %v/%v", brokenBranch.Repo.Name, brokenBranch.Name)
			}

			// Update repo size if this is the master branch
			if branchInfo.Name == "master" {
				if branchInfo.Head != nil {
					headCommitInfo, err := d.resolveCommit(txnCtx.Stm, branchInfo.Head)
					if err != nil {
						return err
					}
					repoInfo.SizeBytes = headCommitInfo.SizeBytes
				} else {
					// No HEAD commit, set the repo size to 0
					repoInfo.SizeBytes = 0
				}

				if err := repos.Put(repo, repoInfo); err != nil {
					return err
				}
			}
		}
	}

	// 8) propagate the changes to 'branch' and its subvenance. This may start
	// new HEAD commits downstream, if the new branch heads haven't been
	// processed yet
	for _, afBranch := range affectedBranches {
		if err := txnCtx.PropagateCommit(afBranch.Branch, false); err != nil {
			return err
		}
	}

	return nil
}

// this is a helper function to check if the given provenance has provenance on an input branch
func provenantOnInput(provenance []*pfs.CommitProvenance) bool {
	provenanceCount := len(provenance)
	for _, p := range provenance {
		// in particular, we want to exclude provenance on the spec repo (used e.g. for spouts)
		if p.Commit.Repo.Name == ppsconsts.SpecRepo {
			provenanceCount--
			break
		}
	}
	return provenanceCount > 0
}

func (d *driver) subscribeCommit(pachClient *client.APIClient, repo *pfs.Repo, branch string, prov *pfs.CommitProvenance, from *pfs.Commit, state pfs.CommitState, cb func(*pfs.CommitInfo) error) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}
	if from != nil && from.Repo.Name != repo.Name {
		return errors.Errorf("the `from` commit needs to be from repo %s", repo.Name)
	}

	commits := d.commits(repo.Name).ReadOnly(pachClient.Ctx())
	newCommitWatcher, err := commits.Watch(watch.WithSort(etcd.SortByCreateRevision, etcd.SortAscend))
	if err != nil {
		return err
	}
	defer newCommitWatcher.Close()
	// keep track of the commits that have been sent
	seen := make(map[string]bool)
	for {
		var commitID string
		commitInfo := &pfs.CommitInfo{}
		var event *watch.Event
		var ok bool
		event, ok = <-newCommitWatcher.Watch()
		if !ok {
			return nil
		}
		switch event.Type {
		case watch.EventError:
			return event.Err
		case watch.EventPut:
			if err := event.Unmarshal(&commitID, commitInfo); err != nil {
				return errors.Wrapf(err, "unmarshal")
			}
			if commitInfo == nil {
				return errors.Errorf("commit info is empty for id: %v", commitID)
			}

			// if provenance is provided, ensure that the returned commits have the commit in their provenance
			if prov != nil {
				valid := false
				for _, cProv := range commitInfo.Provenance {
					valid = valid || proto.Equal(cProv, prov)
				}
				if !valid {
					continue
				}
			}

			if commitInfo.Branch != nil {
				// if branch is provided, make sure the commit was created on that branch
				if branch != "" && commitInfo.Branch.Name != branch {
					continue
				}
				// For now, we don't want stats branches to have jobs triggered on them
				// and this is the simplest way to achieve that. Once we have labels,
				// we'll use those instead for a more principled approach.
				// TODO: Address this sooner rather than later...
				if commitInfo.Branch.Name == "stats" {
					continue
				}
			}

			// We don't want to include the `from` commit itself
			if !(seen[commitID] || (from != nil && from.ID == commitID)) {
				commitInfo, err := d.inspectCommit(pachClient, client.NewCommit(repo.Name, commitID), state)
				if err != nil {
					return err
				}
				if err := cb(commitInfo); err != nil {
					return err
				}
				seen[commitInfo.Commit.ID] = true
			}
		case watch.EventDelete:
			continue
		}
	}
}

func (d *driver) flushCommit(pachClient *client.APIClient, fromCommits []*pfs.Commit, toRepos []*pfs.Repo, cb func(*pfs.CommitInfo) error) error {
	if len(fromCommits) == 0 {
		return errors.Errorf("fromCommits cannot be empty")
	}

	// First compute intersection of the fromCommits subvenant commits, those
	// are the commits we're interested in. Iterate over all commits and keep a
	// running intersection (in commitsToWatch) of the subvenance of all commits
	// processed so far
	commitsToWatch := make(map[string]*pfs.Commit)
	for i, commit := range fromCommits {
		commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
		if err != nil {
			return err
		}
		if i == 0 {
			for _, subvCommit := range commitInfo.Subvenance {
				commitsToWatch[commitKey(subvCommit.Upper)] = subvCommit.Upper
			}
		} else {
			newCommitsToWatch := make(map[string]*pfs.Commit)
			for _, subvCommit := range commitInfo.Subvenance {
				if _, ok := commitsToWatch[commitKey(subvCommit.Upper)]; ok {
					newCommitsToWatch[commitKey(subvCommit.Upper)] = subvCommit.Upper
				}
			}
			commitsToWatch = newCommitsToWatch
		}
	}

	// Compute a map of repos we're flushing to.
	toRepoMap := make(map[string]*pfs.Repo)
	for _, toRepo := range toRepos {
		toRepoMap[toRepo.Name] = toRepo
	}

	// Wait for each of the commitsToWatch to be finished.

	// It's possible that downstream commits will create more downstream
	// commits when they finish due to a trigger firing. To deal with this we
	// loop while we watch commits and add newly discovered commits to the
	// commitsToWatch map.
	watchedCommits := make(map[string]bool)
	for {
		if len(watchedCommits) == len(commitsToWatch) {
			// We've watched every commit so it's time to break.
			break
		}
		additionalCommitsToWatch := make(map[string]*pfs.Commit)
		for key, commitToWatch := range commitsToWatch {
			if watchedCommits[key] {
				continue
			}
			watchedCommits[key] = true
			if len(toRepoMap) > 0 {
				if _, ok := toRepoMap[commitToWatch.Repo.Name]; !ok {
					continue
				}
			}
			finishedCommitInfo, err := d.inspectCommit(pachClient, commitToWatch, pfs.CommitState_FINISHED)
			if err != nil {
				if errors.As(err, &pfsserver.ErrCommitNotFound{}) {
					continue // just skip this
				} else if auth.IsErrNotAuthorized(err) {
					continue // again, just skip (we can't wait on commits we can't access)
				}
				return err
			}
			if err := cb(finishedCommitInfo); err != nil {
				return err
			}
			for _, subvCommit := range finishedCommitInfo.Subvenance {
				additionalCommitsToWatch[commitKey(subvCommit.Upper)] = subvCommit.Upper
			}
		}
		for key, additionalCommit := range additionalCommitsToWatch {
			commitsToWatch[key] = additionalCommit
		}
	}
	// Now wait for the root commits to finish. These are not passed to `f`
	// because it's expecting to just get downstream commits.
	for _, commit := range fromCommits {
		_, err := d.inspectCommit(pachClient, commit, pfs.CommitState_FINISHED)
		if err != nil {
			if errors.As(err, &pfsserver.ErrCommitNotFound{}) {
				continue // just skip this
			}
			return err
		}
	}

	return nil
}

func (d *driver) clearCommit(pachClient *client.APIClient, commit *pfs.Commit) error {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return errors.Errorf("cannot clear finished commit")
	}
	return d.commitStore.DropFilesets(ctx, commit)
}

// createBranch creates a new branch or updates an existing branch (must be one
// or the other). Most importantly, it sets 'branch.DirectProvenance' to
// 'provenance' and then for all (downstream) branches, restores the invariant:
//   ∀ b . b.Provenance = ∪ b'.Provenance (where b' ∈ b.DirectProvenance)
//
// This invariant is assumed to hold for all branches upstream of 'branch', but not
// for 'branch' itself once 'b.Provenance' has been set.
func (d *driver) createBranch(txnCtx *txnenv.TransactionContext, branch *pfs.Branch, commit *pfs.Commit, provenance []*pfs.Branch, trigger *pfs.Trigger) error {
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

	var err error
	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo.Name, auth.Permission_REPO_CREATE_BRANCH); err != nil {
		return err
	}
	// Validate request
	if err := ancestry.ValidateName(branch.Name); err != nil {
		return err
	}
	// The request must do exactly one of:
	// 1) updating 'branch's provenance (commit is nil OR commit == branch)
	// 2) re-pointing 'branch' at a new commit
	var ci *pfs.CommitInfo
	if commit != nil {
		// Determine if this is a provenance update
		sameTarget := branch.Repo.Name == commit.Repo.Name && branch.Name == commit.ID
		if !sameTarget && provenance != nil {
			ci, err = d.resolveCommit(txnCtx.Stm, commit)
			if err != nil {
				return err
			}
			for _, provBranch := range provenance {
				provBranchInfo := &pfs.BranchInfo{}
				if err := d.branches(provBranch.Repo.Name).ReadWrite(txnCtx.Stm).Get(provBranch.Name, provBranchInfo); err != nil {
					// If the branch doesn't exist no need to count it in provenance
					if col.IsErrNotFound(err) {
						continue
					}
					return err
				}
				for _, provC := range ci.Provenance {
					if proto.Equal(provBranch, provC.Branch) && !proto.Equal(provBranchInfo.Head, provC.Commit) {
						return errors.Errorf("cannot create branch %q with commit %q as head because commit has \"%s/%s\" as provenance but that commit is not the head of branch \"%s/%s\"", branch.Name, commit.ID, provC.Commit.Repo.Name, provC.Commit.ID, provC.Branch.Repo.Name, provC.Branch.Name)
					}
				}
			}
		}
	}

	// if 'commit' is a branch, resolve it
	if commit != nil {
		_, err = d.resolveCommit(txnCtx.Stm, commit) // if 'commit' is a branch, resolve it
		if err != nil {
			// possible that branch exists but has no head commit. This is fine, but
			// branchInfo.Head must also be nil
			if !isNoHeadErr(err) {
				return errors.Wrapf(err, "unable to inspect %s@%s", commit.Repo.Name, commit.ID)
			}
			commit = nil
		}
	}

	// Retrieve (and create, if necessary) the current version of this branch
	branches := d.branches(branch.Repo.Name).ReadWrite(txnCtx.Stm)
	branchInfo := &pfs.BranchInfo{}
	if err := branches.Upsert(branch.Name, branchInfo, func() error {
		branchInfo.Name = branch.Name // set in case 'branch' is new
		branchInfo.Branch = branch
		branchInfo.Head = commit
		branchInfo.DirectProvenance = nil
		for _, provBranch := range provenance {
			if provBranch.Repo.Name == branch.Repo.Name && provBranch.Name == branch.Name {
				return errors.Errorf("branch %s@%s cannot be in its own provenance", branch.Repo.Name, branch.Name)
			}
			add(&branchInfo.DirectProvenance, provBranch)
		}
		if trigger != nil && trigger.Branch != "" {
			branchInfo.Trigger = trigger
		}
		return nil
	}); err != nil {
		return err
	}
	repos := d.repos.ReadWrite(txnCtx.Stm)
	repoInfo := &pfs.RepoInfo{}
	if err := repos.Update(branch.Repo.Name, repoInfo, func() error {
		add(&repoInfo.Branches, branch)
		if branch.Name == "master" && commit != nil {
			ci, err := d.resolveCommit(txnCtx.Stm, commit)
			if err != nil {
				return err
			}
			repoInfo.SizeBytes = ci.SizeBytes
		}
		return nil
	}); err != nil {
		return err
	}

	// Update (or create)
	// 1) 'branch's Provenance
	// 2) the Provenance of all branches in 'branch's Subvenance (in the case of an update), and
	// 3) the Subvenance of all branches in the *old* provenance of 'branch's Subvenance
	toUpdate := []*pfs.BranchInfo{branchInfo}
	for _, subvBranch := range branchInfo.Subvenance {
		subvBranchInfo := &pfs.BranchInfo{}
		if err := d.branches(subvBranch.Repo.Name).ReadWrite(txnCtx.Stm).Get(subvBranch.Name, subvBranchInfo); err != nil {
			return err
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
			if err := d.addBranchProvenance(branchInfo, provBranch, txnCtx.Stm); err != nil {
				return err
			}
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches(provBranch.Repo.Name).ReadWrite(txnCtx.Stm).Get(provBranch.Name, provBranchInfo); err != nil {
				return errors.Wrapf(err, "error getting prov branch")
			}
			for _, provBranch := range provBranchInfo.Provenance {
				// add provBranch to branchInfo.Provenance, and branchInfo.Branch to
				// provBranch subvenance
				if err := d.addBranchProvenance(branchInfo, provBranch, txnCtx.Stm); err != nil {
					return err
				}
			}
		}
		// If we have a commit use it to set head of this branch info
		if ci != nil {
			for _, provC := range ci.Provenance {
				if proto.Equal(provC.Branch, branchInfo.Branch) {
					branchInfo.Head = provC.Commit
				}
			}
		}
		if err := d.branches(branchInfo.Branch.Repo.Name).ReadWrite(txnCtx.Stm).Put(branchInfo.Branch.Name, branchInfo); err != nil {
			return err
		}
		// Update Subvenance of 'branchInfo's Provenance (incl. all Subvenance)
		for _, oldProvBranch := range oldProvenance {
			if !has(&branchInfo.Provenance, oldProvBranch) {
				// Provenance was deleted, so we delete ourselves from their subvenance
				oldProvBranchInfo := &pfs.BranchInfo{}
				if err := d.branches(oldProvBranch.Repo.Name).ReadWrite(txnCtx.Stm).Update(oldProvBranch.Name, oldProvBranchInfo, func() error {
					del(&oldProvBranchInfo.Subvenance, branchInfo.Branch)
					return nil
				}); err != nil {
					return err
				}
			}
		}
	}

	// propagate the head commit to 'branch'. This may also modify 'branch', by
	// creating a new HEAD commit if 'branch's provenance was changed and its
	// current HEAD commit has old provenance
	var triggeredBranches []*pfs.Branch
	if commit != nil {
		if ci == nil {
			ci, err = d.resolveCommit(txnCtx.Stm, commit)
			if err != nil {
				return err
			}
		}
		if ci.Finished != nil {
			triggeredBranches, err = d.triggerCommit(txnCtx, ci.Commit)
			if err != nil {
				return err
			}
		}
	}
	for _, b := range append(triggeredBranches, branch) {
		if err := txnCtx.PropagateCommit(b, false); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) inspectBranch(txnCtx *txnenv.TransactionContext, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	// Validate arguments
	if branch == nil {
		return nil, errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return nil, errors.New("branch repo cannot be nil")
	}

	// Check that the user is logged in, but don't require any access level
	if _, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{}); err != nil {
		if !auth.IsErrNotActivated(err) {
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to run fsck)")
		}
	}

	result := &pfs.BranchInfo{}
	if err := d.branches(branch.Repo.Name).ReadWrite(txnCtx.Stm).Get(branch.Name, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) listBranch(pachClient *client.APIClient, repo *pfs.Repo, reverse bool) ([]*pfs.BranchInfo, error) {
	// Validate arguments
	if repo == nil {
		return nil, errors.New("repo cannot be nil")
	}

	if err := authserver.CheckRepoIsAuthorized(pachClient, repo.Name, auth.Permission_REPO_LIST_BRANCH); err != nil {
		return nil, err
	}

	// Make sure that the repo exists
	if repo.Name != "" {
		err := d.txnEnv.WithReadContext(pachClient.Ctx(), func(txnCtx *txnenv.TransactionContext) error {
			_, err := d.inspectRepo(txnCtx, repo, !includeAuth)
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	var result []*pfs.BranchInfo
	branchInfo := &pfs.BranchInfo{}
	branches := d.branches(repo.Name).ReadOnly(pachClient.Ctx())
	opts := *col.DefaultOptions // Note we dereference here so as to make a copy
	if reverse {
		opts.Order = etcd.SortAscend
	}
	var bis []*pfs.BranchInfo
	sendBis := func() {
		if !reverse {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) < len(bis[j].Provenance) })
		} else {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) > len(bis[j].Provenance) })
		}
		result = append(result, bis...)
		bis = nil
	}
	lastRev := int64(-1)
	if err := branches.ListRev(branchInfo, &opts, func(branch string, createRev int64) error {
		if createRev != lastRev {
			sendBis()
			lastRev = createRev
		}
		bis = append(bis, proto.Clone(branchInfo).(*pfs.BranchInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	sendBis()
	return result, nil
}

func (d *driver) deleteBranch(txnCtx *txnenv.TransactionContext, branch *pfs.Branch, force bool) error {
	// Validate arguments
	if branch == nil {
		return errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return errors.New("branch repo cannot be nil")
	}

	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, branch.Repo.Name, auth.Permission_REPO_DELETE_BRANCH); err != nil {
		return err
	}

	branches := d.branches(branch.Repo.Name).ReadWrite(txnCtx.Stm)
	branchInfo := &pfs.BranchInfo{}
	if err := branches.Get(branch.Name, branchInfo); err != nil {
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
		if err := branches.Delete(branch.Name); err != nil {
			return errors.Wrapf(err, "branches.Delete")
		}
		for _, provBranch := range branchInfo.Provenance {
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches(provBranch.Repo.Name).ReadWrite(txnCtx.Stm).Update(provBranch.Name, provBranchInfo, func() error {
				del(&provBranchInfo.Subvenance, branch)
				return nil
			}); err != nil && !isNotFoundErr(err) {
				return errors.Wrapf(err, "error deleting subvenance")
			}
		}
	}
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(txnCtx.Stm).Update(branch.Repo.Name, repoInfo, func() error {
		del(&repoInfo.Branches, branch)
		return nil
	}); err != nil {
		if !col.IsErrNotFound(err) || !force {
			return err
		}
	}
	return nil
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func isNoHeadErr(err error) bool {
	return errors.As(err, &pfsserver.ErrNoHead{})
}

func commitKey(commit *pfs.Commit) string {
	return fmt.Sprintf("%s/%s", commit.Repo.Name, commit.ID)
}

func branchKey(branch *pfs.Branch) string {
	return fmt.Sprintf("%s/%s", branch.Repo.Name, branch.Name)
}

func (d *driver) addBranchProvenance(branchInfo *pfs.BranchInfo, provBranch *pfs.Branch, stm col.STM) error {
	if provBranch.Repo.Name == branchInfo.Branch.Repo.Name && provBranch.Name == branchInfo.Branch.Name {
		return errors.Errorf("provenance loop, branch %s/%s cannot be provenance for itself", provBranch.Repo.Name, provBranch.Name)
	}
	add(&branchInfo.Provenance, provBranch)
	provBranchInfo := &pfs.BranchInfo{}
	if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Upsert(provBranch.Name, provBranchInfo, func() error {
		// Set provBranch, we may be creating this branch for the first time
		provBranchInfo.Name = provBranch.Name
		provBranchInfo.Branch = provBranch
		add(&provBranchInfo.Subvenance, branchInfo.Branch)
		return nil
	}); err != nil {
		return err
	}
	repoInfo := &pfs.RepoInfo{}
	return d.repos.ReadWrite(stm).Update(provBranch.Repo.Name, repoInfo, func() error {
		add(&repoInfo.Branches, provBranch)
		return nil
	})
}

func (d *driver) appendSubvenance(commitInfo *pfs.CommitInfo, subvCommitInfo *pfs.CommitInfo) {
	if subvCommitInfo.ParentCommit != nil {
		for _, subvCommitRange := range commitInfo.Subvenance {
			if subvCommitRange.Upper.ID == subvCommitInfo.ParentCommit.ID {
				subvCommitRange.Upper = subvCommitInfo.Commit
				return
			}
		}
	}
	commitInfo.Subvenance = append(commitInfo.Subvenance, &pfs.CommitRange{
		Lower: subvCommitInfo.Commit,
		Upper: subvCommitInfo.Commit,
	})
	if !d.env.DisableCommitProgressCounter {
		commitInfo.SubvenantCommitsTotal++
	}
}

func (d *driver) deleteAll(txnCtx *txnenv.TransactionContext) error {
	// Note: d.listRepo() doesn't return the 'spec' repo, so it doesn't get
	// deleted here. Instead, PPS is responsible for deleting and re-creating it
	repoInfos, err := d.listRepo(txnCtx.Client, !includeAuth)
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		if err := d.deleteRepo(txnCtx, repoInfo.Repo, true); err != nil && !auth.IsErrNotAuthorized(err) {
			return err
		}
	}
	return nil
}

// TODO: Is this really necessary?
type branchSet []*pfs.Branch

func (b *branchSet) search(branch *pfs.Branch) (int, bool) {
	key := branchKey(branch)
	i := sort.Search(len(*b), func(i int) bool {
		return branchKey((*b)[i]) >= key
	})
	if i == len(*b) {
		return i, false
	}
	return i, branchKey((*b)[i]) == branchKey(branch)
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
	if err != sql.ErrNoRows {
		return secret, err
	}
	if err == sql.ErrNoRows {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		logrus.Infof("generated new secret: %q", name)
		if err := keyStore.Create(ctx, name, secret); err != nil {
			return nil, err
		}
	}
	return keyStore.Get(ctx, name)
}
