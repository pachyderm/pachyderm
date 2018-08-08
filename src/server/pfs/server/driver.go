package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-collections/collections/stack"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/sql"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/golang-lru"
)

const (
	splitSuffixBase  = 16
	splitSuffixWidth = 64
	splitSuffixFmt   = "%016x"

	// Makes calls to ListRepo and InspectRepo more legible
	includeAuth = true

	// maxInt is the maximum value for 'int' (system-dependent). Not in 'math'!
	maxInt = int(^uint(0) >> 1)

	// Maximum number of concurrent put object calls.
	putObjectConcurrency = 100
)

// validateRepoName determines if a repo name is valid
func validateRepoName(name string) error {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if !match {
		return fmt.Errorf("repo name (%v) invalid: only alphanumeric characters, underscores, and dashes are allowed", name)
	}
	return nil
}

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
	// address is used to connect to achd's Object Store and Authorization API
	address string

	// pachClientOnce ensures that _pachClient is only initialized once
	pachClientOnce sync.Once

	// pachClient is a cached Pachd client, that connects to Pachyderm's object
	// store API and auth API. Instead of accessing it directly, functions should
	// call getPachClient()
	_pachClient *client.APIClient

	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	prefix     string

	// collections
	repos          col.Collection
	putFileRecords col.Collection
	commits        collectionFactory
	branches       collectionFactory
	openCommits    col.Collection

	// a cache for hashtrees
	treeCache *lru.Cache
}

// newDriver is used to create a new Driver instance
func newDriver(address string, etcdAddresses []string, etcdPrefix string, treeCache *lru.Cache) (*driver, error) {
	// Validate arguments
	if treeCache == nil {
		return nil, fmt.Errorf("cannot initialize driver with nil treeCache")
	}

	// Initialize etcd client
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddresses,
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to etcd: %v", err)
	}

	// Initialize driver
	d := &driver{
		address:        address,
		etcdClient:     etcdClient,
		prefix:         etcdPrefix,
		repos:          pfsdb.Repos(etcdClient, etcdPrefix),
		putFileRecords: pfsdb.PutFileRecords(etcdClient, etcdPrefix),
		commits: func(repo string) col.Collection {
			return pfsdb.Commits(etcdClient, etcdPrefix, repo)
		},
		branches: func(repo string) col.Collection {
			return pfsdb.Branches(etcdClient, etcdPrefix, repo)
		},
		openCommits: pfsdb.OpenCommits(etcdClient, etcdPrefix),
		treeCache:   treeCache,
	}
	go func() { d.getPachClient(context.Background()) }() // Begin dialing connection on startup
	return d, nil
}

// getPachClient() initializes the connection that the pfs driver has with the
// Pachyderm object API and auth API, and blocks until the connection is
// established
//
// TODO(msteffen): client initialization (both etcd and pachd) might be better
// placed happen in server.go, near main(), so that we only pay the dial cost
// once, and so that pps doesn't need to have its own initialization code
func (d *driver) getPachClient(ctx context.Context) *client.APIClient {
	d.pachClientOnce.Do(func() {
		var err error
		d._pachClient, err = client.NewFromAddress(d.address)
		if err != nil {
			panic(fmt.Sprintf("could not intiailize Pachyderm client in driver: %v", err))
		}
	})
	return d._pachClient.WithCtx(ctx)
}

// checkIsAuthorized returns an error if the current user (in 'ctx') has
// authorization scope 's' for repo 'r'
func (d *driver) checkIsAuthorized(pachClient *client.APIClient, r *pfs.Repo, s auth.Scope) error {
	ctx := pachClient.Ctx()
	me, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}
	resp, err := pachClient.AuthAPIClient.Authorize(ctx, &auth.AuthorizeRequest{
		Repo:  r.Name,
		Scope: s,
	})
	if err != nil {
		return fmt.Errorf("error during authorization check for operation on \"%s\": %v",
			r.Name, grpcutil.ScrubGRPC(err))
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

func present(key string) etcd.Cmp {
	return etcd.Compare(etcd.CreateRevision(key), ">", 0)
}

func absent(key string) etcd.Cmp {
	return etcd.Compare(etcd.CreateRevision(key), "=", 0)
}

func (d *driver) createRepo(ctx context.Context, repo *pfs.Repo, description string, update bool) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	// Check that the user is logged in (user doesn't need any access level to
	// create a repo, but they must be authenticated if auth is active)
	whoAmI, err := pachClient.AuthAPIClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	authIsActivated := !auth.IsErrNotActivated(err)
	if authIsActivated && err != nil {
		return fmt.Errorf("error authenticating (must log in to create a repo): %v",
			grpcutil.ScrubGRPC(err))
	}
	if err := validateRepoName(repo.Name); err != nil {
		return err
	}
	if update {
		return d.updateRepo(ctx, repo, description)
	}

	_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)

		// check if 'repo' already exists. If so, return that error. Otherwise,
		// proceed with ACL creation (avoids awkward "access denied" error when
		// calling "createRepo" on a repo that already exists)
		var existingRepoInfo pfs.RepoInfo
		err := repos.Get(repo.Name, &existingRepoInfo)
		if err != nil && !col.IsErrNotFound(err) {
			return fmt.Errorf("error checking whether \"%s\" exists: %v",
				repo.Name, err)
		} else if err == nil {
			return fmt.Errorf("cannot create \"%s\" as it already exists", repo.Name)
		}

		// Create ACL for new repo
		if authIsActivated {
			// auth is active, and user is logged in. Make user an owner of the new
			// repo (and clear any existing ACL under this name that might have been
			// created by accident)
			_, err := pachClient.AuthAPIClient.SetACL(ctx, &auth.SetACLRequest{
				Repo: repo.Name,
				Entries: []*auth.ACLEntry{{
					Username: whoAmI.Username,
					Scope:    auth.Scope_OWNER,
				}},
			})
			if err != nil {
				return fmt.Errorf("could not create ACL for new repo \"%s\": %v",
					repo.Name, grpcutil.ScrubGRPC(err))
			}
		}

		repoInfo := &pfs.RepoInfo{
			Repo:        repo,
			Created:     now(),
			Description: description,
		}
		return repos.Create(repo.Name, repoInfo)
	})
	return err
}

func (d *driver) updateRepo(ctx context.Context, repo *pfs.Repo, description string) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(repo.Name, repoInfo); err != nil {
			return fmt.Errorf("error updating repo: %v", err)
		}
		// Caller only neads to be a WRITER to call UpdatePipeline(), so caller only
		// needs to be a WRITER to update the provenance.
		if err := d.checkIsAuthorized(pachClient, repo, auth.Scope_WRITER); err != nil {
			return err
		}
		repoInfo.Description = description
		return repos.Put(repo.Name, repoInfo)
	})
	return err
}

func (d *driver) inspectRepo(ctx context.Context, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
	result := &pfs.RepoInfo{}
	if err := d.repos.ReadOnly(ctx).Get(repo.Name, result); err != nil {
		return nil, err
	}
	if includeAuth {
		accessLevel, err := d.getAccessLevel(ctx, repo)
		if err != nil {
			if auth.IsErrNotActivated(err) {
				return result, nil
			}
			return nil, fmt.Errorf("error getting access level for \"%s\": %v",
				repo.Name, grpcutil.ScrubGRPC(err))
		}
		result.AuthInfo = &pfs.RepoAuthInfo{AccessLevel: accessLevel}
	}
	return result, nil
}

func (d *driver) getAccessLevel(ctx context.Context, repo *pfs.Repo) (auth.Scope, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	who, err := pachClient.AuthAPIClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if err != nil {
		return auth.Scope_NONE, err
	}
	if who.IsAdmin {
		return auth.Scope_OWNER, nil
	}
	resp, err := pachClient.AuthAPIClient.GetScope(ctx, &auth.GetScopeRequest{Repos: []string{repo.Name}})
	if err != nil {
		return auth.Scope_NONE, err
	}
	if len(resp.Scopes) != 1 {
		return auth.Scope_NONE, fmt.Errorf("too many results from GetScope: %#v", resp)
	}
	return resp.Scopes[0], nil
}

func (d *driver) listRepo(ctx context.Context, includeAuth bool) (*pfs.ListRepoResponse, error) {
	repos := d.repos.ReadOnly(ctx)
	result := &pfs.ListRepoResponse{}
	authSeemsActive := true
	repoInfo := &pfs.RepoInfo{}
	if err := repos.List(repoInfo, col.DefaultOptions, func(repoName string) error {
		if repoName == ppsconsts.SpecRepo {
			return nil
		}
		if includeAuth && authSeemsActive {
			accessLevel, err := d.getAccessLevel(ctx, repoInfo.Repo)
			if err == nil {
				repoInfo.AuthInfo = &pfs.RepoAuthInfo{AccessLevel: accessLevel}
			} else if auth.IsErrNotActivated(err) {
				authSeemsActive = false
			} else {
				return fmt.Errorf("error getting access level for \"%s\": %v",
					repoName, grpcutil.ScrubGRPC(err))
			}
		}
		result.RepoInfo = append(result.RepoInfo, proto.Clone(repoInfo).(*pfs.RepoInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) deleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	// TODO(msteffen): Fix d.deleteAll() so that it doesn't need to delete and
	// recreate the PPS spec repo, then uncomment this block to prevent users from
	// deleting it and breaking their cluster
	// if repo.Name == ppsconsts.SpecRepo {
	// 	return fmt.Errorf("cannot delete the special PPS repo %s", ppsconsts.SpecRepo)
	// }
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		commits := d.commits(repo.Name).ReadWrite(stm)

		// check if 'repo' is already gone. If so, return that error. Otherwise,
		// proceed with auth check (avoids awkward "access denied" error when calling
		// "deleteRepo" on a repo that's already gone)
		var existingRepoInfo pfs.RepoInfo
		err := repos.Get(repo.Name, &existingRepoInfo)
		if err != nil {
			if !col.IsErrNotFound(err) {
				return fmt.Errorf("error checking whether \"%s\" exists: %v",
					repo.Name, err)
			}
		}

		// Check if the caller is authorized to delete this repo
		if err := d.checkIsAuthorized(pachClient, repo, auth.Scope_OWNER); err != nil {
			return err
		}

		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(repo.Name, repoInfo); err != nil {
			if !col.IsErrNotFound(err) {
				return fmt.Errorf("repos.Get: %v", err)
			}
		}
		commits.DeleteAll()
		var branchInfos []*pfs.BranchInfo
		for _, branch := range repoInfo.Branches {
			bi, err := d.inspectBranch(ctx, branch)
			if err != nil {
				return fmt.Errorf("error inspecting branch %s: %v", branch, err)
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
			if err := d.deleteBranchSTM(stm, branch, force); err != nil {
				return fmt.Errorf("delete branch %s: %v", branch, err)
			}
		}
		// Despite the fact that we already deleted each branch with
		// deleteBranchSTM we also do branches.DeleteAll(), this insulates us
		// against certain corruption situations where the RepoInfo doesn't
		// exist in etcd but branches do.
		branches := d.branches(repo.Name).ReadWrite(stm)
		branches.DeleteAll()
		if err := repos.Delete(repo.Name); err != nil && !col.IsErrNotFound(err) {
			return fmt.Errorf("repos.Delete: %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if _, err = pachClient.SetACL(ctx, &auth.SetACLRequest{
		Repo: repo.Name, // NewACL is unset, so this will clear the acl for 'repo'
	}); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	return nil
}

func (d *driver) startCommit(ctx context.Context, parent *pfs.Commit, branch string, provenance []*pfs.Commit, description string) (*pfs.Commit, error) {
	return d.makeCommit(ctx, "", parent, branch, provenance, nil, nil, nil, description)
}

func (d *driver) buildCommit(ctx context.Context, ID string, parent *pfs.Commit, branch string, provenance []*pfs.Commit, tree *pfs.Object) (*pfs.Commit, error) {
	return d.makeCommit(ctx, ID, parent, branch, provenance, tree, nil, nil, "")
}

// make commit makes a new commit in 'branch', with the parent 'parent' and the
// direct provenance 'provenance'. Note that
// - 'parent' must not be nil, but the only required field is 'parent.Repo'.
// - 'parent.ID' may be set to "", in which case the parent commit is inferred
//   from 'parent.Repo' and 'branch'.
// - If both 'parent.ID' and 'branch' are set, 'parent.ID' determines the parent
//   commit, but 'branch' is still moved to point at the new commit
//   to the new commit
// - If neither 'parent.ID' nor 'branch' are set, the new commit will have no
//   parent
func (d *driver) makeCommit(ctx context.Context, ID string, parent *pfs.Commit, branch string, provenance []*pfs.Commit, treeRef *pfs.Object, recordFiles []string, records []*pfs.PutFileRecords, description string) (*pfs.Commit, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	// Validate arguments:
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}

	// Check that caller is authorized
	if err := d.checkIsAuthorized(pachClient, parent.Repo, auth.Scope_WRITER); err != nil {
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
	newCommitInfo := &pfs.CommitInfo{
		Commit:      newCommit,
		Started:     now(),
		Description: description,
	}

	//  BuildCommit case: if the caller passed a tree reference with the commit
	// contents then retrieve the full tree so we can compute its size
	var tree hashtree.HashTree
	if treeRef != nil {
		var buf bytes.Buffer
		if err := pachClient.GetObject(treeRef.Hash, &buf); err != nil {
			return nil, err
		}
		var err error
		tree, err = hashtree.Deserialize(buf.Bytes())
		if err != nil {
			return nil, err
		}
	}

	// Txn: create the actual commit in etcd and update the branch + parent/child
	if _, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		// Clone the parent, as this stm modifies it and might wind up getting
		// run more than once (if there's a conflict.)
		parent := proto.Clone(parent).(*pfs.Commit)
		repos := d.repos.ReadWrite(stm)
		commits := d.commits(parent.Repo.Name).ReadWrite(stm)
		branches := d.branches(parent.Repo.Name).ReadWrite(stm)

		// Check if repo exists
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(parent.Repo.Name, repoInfo); err != nil {
			return err
		}

		// create/update 'branch' (if it was set) and set parent.ID (if, in addition,
		// 'parent.ID' was not set)
		if branch != "" {
			branchInfo := &pfs.BranchInfo{}
			if err := branches.Upsert(branch, branchInfo, func() error {
				if parent.ID == "" && branchInfo.Head != nil {
					parent.ID = branchInfo.Head.ID
				}
				// Point 'branch' at the new commit
				branchInfo.Name = branch // set in case 'branch' is new
				branchInfo.Head = newCommit
				branchInfo.Branch = client.NewBranch(newCommit.Repo.Name, branch)
				return nil
			}); err != nil {
				return err
			}
			add(&repoInfo.Branches, branchInfo.Branch)
			if len(branchInfo.Provenance) > 0 && treeRef == nil {
				return fmt.Errorf("cannot start a commit on an output branch")
			}
		}

		// Set newCommit.ParentCommit (if 'parent' and/or 'branch' was set) and add
		// newCommit to parent's ChildCommits
		if parent.ID != "" {
			// Resolve parent.ID if it's a branch that isn't 'branch' (which can
			// happen if 'branch' is new and diverges from the existing branch in
			// 'parent.ID')
			parentCommitInfo, err := d.resolveCommit(stm, parent)
			if err != nil {
				return fmt.Errorf("parent commit not found: %v", err)
			}
			// fail if the parent commit has not been finished
			if parentCommitInfo.Finished == nil {
				return fmt.Errorf("parent commit %s has not been finished", parent.ID)
			}
			if err := commits.Update(parent.ID, parentCommitInfo, func() error {
				newCommitInfo.ParentCommit = parent
				parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
				return nil
			}); err != nil {
				// Note: error is emitted if parent.ID is a missing/invalid branch OR a
				// missing/invalid commit ID
				return fmt.Errorf("could not resolve parent commit \"%s\": %v", parent.ID, err)
			}
		}

		// BuildCommit case: Now that 'parent' is resolved, read the parent commit's
		// tree (inside txn) and update the repo size
		if treeRef != nil || records != nil {
			parentTree, err := d.getTreeForCommit(ctx, parent)
			if err != nil {
				return err
			}
			if records != nil {
				openTree := parentTree.Open()
				for i, record := range records {
					if err := d.applyWrite(recordFiles[i], record, openTree); err != nil {
						return err
					}
				}
				tree, err = openTree.Finish()
				if err != nil {
					return err
				}
				data, err := hashtree.Serialize(tree)
				if err != nil {
					return err
				}
				treeRef, _, err = pachClient.PutObject(bytes.NewReader(data))
				if err != nil {
					return err
				}
			}
			repoInfo.SizeBytes += sizeChange(tree, parentTree)
		} else {
			if err := d.openCommits.ReadWrite(stm).Put(newCommit.ID, newCommit); err != nil {
				return err
			}
		}
		if treeRef != nil {
			newCommitInfo.Tree = treeRef
			newCommitInfo.SizeBytes = uint64(tree.FSSize())
			newCommitInfo.Finished = now()
		}

		if err := repos.Put(parent.Repo.Name, repoInfo); err != nil {
			return err
		}

		// 'newCommitProv' holds newCommit's provenance (use map for deduping).
		newCommitProv := make(map[string]*pfs.Commit)

		// Build newCommit's full provenance; my provenance's provenance is my
		// provenance (b/c provenance' is a transitive closure, there's no need to
		// explore full graph)
		for _, provCommit := range provenance {
			newCommitProv[provCommit.ID] = provCommit
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(provCommit.Repo.Name).ReadWrite(stm).Get(provCommit.ID, provCommitInfo); err != nil {
				return err
			}
			for _, c := range provCommitInfo.Provenance {
				newCommitProv[c.ID] = c
			}
		}

		// Copy newCommitProv into newCommitInfo.Provenance, and update upstream subv
		for _, provCommit := range newCommitProv {
			newCommitInfo.Provenance = append(newCommitInfo.Provenance, provCommit)
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(provCommit.Repo.Name).ReadWrite(stm).Update(provCommit.ID, provCommitInfo, func() error {
				appendSubvenance(provCommitInfo, newCommitInfo)
				return nil
			}); err != nil {
				return err
			}
		}

		// Finally, create the commit
		if err := commits.Create(newCommit.ID, newCommitInfo); err != nil {
			return err
		}
		// We propagate the branch last so propagateCommit can write to the
		// now-existing commit's subvenance
		if branch != "" {
			return d.propagateCommit(stm, client.NewBranch(newCommit.Repo.Name, branch))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return newCommit, nil
}

func (d *driver) finishCommit(ctx context.Context, commit *pfs.Commit, tree *pfs.Object, empty bool, description string) (retErr error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return fmt.Errorf("commit %s has already been finished", commit.FullID())
	}
	if description != "" {
		commitInfo.Description = description
	}

	scratchPrefix := d.scratchCommitPrefix(commit)
	defer func() {
		if retErr != nil {
			return
		}
		// only delete the scratch space if finishCommit() ran successfully and
		// won't need to be retried
		_, retErr = d.etcdClient.Delete(ctx, scratchPrefix, etcd.WithPrefix())
	}()

	var parentTree, finishedTree hashtree.HashTree
	if !empty {
		// Retrieve the parent commit's tree (to apply writes from etcd or just
		// compute the size change). If parentCommit.Tree == nil, walk up the branch
		// until we find a successful commit. Otherwise, require that the immediate
		// parent of 'commitInfo' is closed, as we use its contents
		parentCommit := commitInfo.ParentCommit
		for parentCommit != nil {
			parentCommitInfo, err := d.inspectCommit(ctx, parentCommit, pfs.CommitState_STARTED)
			if err != nil {
				return err
			}
			if parentCommitInfo.Tree != nil {
				break
			}
			parentCommit = parentCommitInfo.ParentCommit
		}
		parentTree, err = d.getTreeForCommit(ctx, parentCommit) // result is empty if parentCommit == nil
		if err != nil {
			return err
		}

		if tree == nil {
			var err error
			finishedTree, err = d.getTreeForOpenCommit(ctx, &pfs.File{Commit: commit}, parentTree)
			if err != nil {
				return err
			}
			// Serialize the tree
			data, err := hashtree.Serialize(finishedTree)
			if err != nil {
				return err
			}

			if len(data) > 0 {
				// Put the tree into the blob store
				obj, _, err := pachClient.PutObject(bytes.NewReader(data))
				if err != nil {
					return err
				}

				commitInfo.Tree = obj
			}
		} else {
			var buf bytes.Buffer
			if err := pachClient.GetObject(tree.Hash, &buf); err != nil {
				return err
			}
			var err error
			finishedTree, err = hashtree.Deserialize(buf.Bytes())
			if err != nil {
				return err
			}
			commitInfo.Tree = tree
		}

		commitInfo.SizeBytes = uint64(finishedTree.FSSize())
	}

	commitInfo.Finished = now()
	sizeChange := sizeChange(finishedTree, parentTree)
	_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		commits := d.commits(commit.Repo.Name).ReadWrite(stm)
		repos := d.repos.ReadWrite(stm)
		if err := commits.Put(commit.ID, commitInfo); err != nil {
			return err
		}
		if err := d.openCommits.ReadWrite(stm).Delete(commit.ID); err != nil {
			return fmt.Errorf("could not confirm that commit %s is open; this is likely a bug. err: %v", commit.ID, err)
		}
		if sizeChange > 0 {
			// update repo size
			repoInfo := new(pfs.RepoInfo)
			if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
				return err
			}

			// Increment the repo sizes by the sizes of the files that have
			// been added in this commit.
			repoInfo.SizeBytes += sizeChange
			if err := repos.Put(commit.Repo.Name, repoInfo); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// propagateCommit selectively starts commits in or downstream of 'branch' in
// order to restore the invariant that branch provenance matches HEAD commit
// provenance:
//   B.Head is provenant on A.Head <=>
//   branch B is provenant on branch A and A.Head != nil
// The implementation assumes that the invariant already holds for all branches
// upstream of 'branch', but not necessarily for 'branch' itself. Despite the
// name, 'branch' does not need a HEAD commit to propagate, though one may be
// created.
//
// In other words, propagateCommit scans all branches b_downstream that are
// equal to or downstream of 'branch', and if the HEAD of b_downstream isn't
// provenant on the HEADs of b_downstream's provenance, propagateCommit starts
// a new HEAD commit in b_downstream that is. For example, propagateCommit
// starts downstream output commits (which trigger PPS jobs) when new input
// commits arrive on 'branch', when 'branch's HEAD is deleted, or when 'branch'
// is newly created (i.e. in CreatePipeline).
func (d *driver) propagateCommit(stm col.STM, branch *pfs.Branch) error {
	if branch == nil {
		return fmt.Errorf("cannot propagate nil branch")
	}

	// 'subvBranchInfos' is the collection of downstream branches that may get a
	// new commit. Populate subvBranchInfo
	var subvBranchInfos []*pfs.BranchInfo
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches(branch.Repo.Name).ReadWrite(stm).Get(branch.Name, branchInfo); err != nil {
		return err
	}
	subvBranchInfos = append(subvBranchInfos, branchInfo) // add 'branch' itself
	for _, subvBranch := range branchInfo.Subvenance {
		subvBranchInfo := &pfs.BranchInfo{}
		if err := d.branches(subvBranch.Repo.Name).ReadWrite(stm).Get(subvBranch.Name, subvBranchInfo); err != nil {
			return err
		}
		subvBranchInfos = append(subvBranchInfos, subvBranchInfo)
	}

	// Sort subvBranchInfos so that upstream branches are processed before their
	// descendants. This guarantees that if branch B is provenant on branch A, we
	// create a new commit in A before creating a new commit in B provenant on the
	// (new) HEAD of A.
	sort.Slice(subvBranchInfos, func(i, j int) bool { return len(subvBranchInfos[i].Provenance) < len(subvBranchInfos[j].Provenance) })

	// Iterate through downstream branches and determine which need a new commit.
nextSubvBranch:
	for _, branchInfo := range subvBranchInfos {
		branch := branchInfo.Branch
		repo := branch.Repo
		commits := d.commits(repo.Name).ReadWrite(stm)
		branches := d.branches(repo.Name).ReadWrite(stm)

		// Compute the full provenance of hypothetical new output commit to decide
		// if we need it
		commitProvMap := make(map[string]*branchCommit)
		for _, provBranch := range branchInfo.Provenance {
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Get(provBranch.Name, provBranchInfo); err != nil && !col.IsErrNotFound(err) {
				return fmt.Errorf("could not read branch %s/%s: %v", provBranch.Repo.Name, provBranch.Name, err)
			}
			if provBranchInfo.Head == nil {
				continue
			}
			commitProvMap[provBranchInfo.Head.ID] = &branchCommit{
				commit: provBranchInfo.Head,
				branch: provBranchInfo.Branch,
			}
			// Because of the 'propagateCommit' invariant, we don't need to inspect
			// provBranchInfo.HEAD's provenance. Every commit in there will be the
			// HEAD of some other provBranchInfo.
		}
		if len(commitProvMap) == 0 {
			// no input commits to process; don't create a new output commit
			continue nextSubvBranch
		}

		// 'branch' may already have a HEAD commit, so compute whether the new
		// output commit would have the same provenance as the existing HEAD
		// commit. If so, a new output commit would be a duplicate, so don't create
		// it.
		if branchInfo.Head != nil {
			branchHeadInfo := &pfs.CommitInfo{}
			if err := commits.Get(branchInfo.Head.ID, branchHeadInfo); err != nil {
				return pfsserver.ErrCommitNotFound{branchInfo.Head}
			}
			headIsSubset := true
			for _, c := range branchHeadInfo.Provenance {
				if _, ok := commitProvMap[c.ID]; !ok {
					headIsSubset = false
					break
				}
			}
			if len(branchHeadInfo.Provenance) == len(commitProvMap) && headIsSubset {
				// existing HEAD commit is the same new output commit would be; don't
				// create new commit
				continue nextSubvBranch
			}
		}

		// If the only branches in the hypothetical output commit's provenance are
		// in the 'spec' repo, creating it would mean creating a confusing
		// "dummy" job with no non-spec input data. If this is the case, don't
		// create a new output commit
		allSpec := true
		for _, b := range commitProvMap {
			if b.branch.Repo.Name != ppsconsts.SpecRepo {
				allSpec = false
				break
			}
		}
		if allSpec {
			// Only input data is PipelineInfo; don't create new output commit
			continue nextSubvBranch
		}

		// *All checks passed* start a new output commit in 'subvBranch'
		newCommit := &pfs.Commit{
			Repo: branch.Repo,
			ID:   uuid.NewWithoutDashes(),
		}
		newCommitInfo := &pfs.CommitInfo{
			Commit:  newCommit,
			Started: now(),
		}

		// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
		newCommitInfo.ParentCommit = branchInfo.Head
		if branchInfo.Head != nil {
			parentCommitInfo := &pfs.CommitInfo{}
			if err := commits.Update(newCommitInfo.ParentCommit.ID, parentCommitInfo, func() error {
				parentCommitInfo.ChildCommits = append(parentCommitInfo.ChildCommits, newCommit)
				return nil
			}); err != nil {
				return err
			}
		}
		branchInfo.Head = newCommit
		branchInfo.Name = branch.Name // set in case 'branch' is new
		branchInfo.Branch = branch    // set in case 'branch' is new
		if err := branches.Put(branch.Name, branchInfo); err != nil {
			return err
		}

		// Set provenance and upstream subvenance (appendSubvenance needs
		// newCommitInfo.ParentCommit to extend the correct subvenance range)
		for _, prov := range commitProvMap {
			// set provenance of 'newCommit'
			newCommitInfo.Provenance = append(newCommitInfo.Provenance, prov.commit)
			newCommitInfo.BranchProvenance = append(newCommitInfo.BranchProvenance, prov.branch)

			// update subvenance of 'prov.commit'
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(prov.commit.Repo.Name).ReadWrite(stm).Update(prov.commit.ID, provCommitInfo, func() error {
				appendSubvenance(provCommitInfo, newCommitInfo)
				return nil
			}); err != nil {
				return err
			}
		}

		// finally create open 'commit'
		if err := commits.Create(newCommit.ID, newCommitInfo); err != nil {
			return err
		}
		if err := d.openCommits.ReadWrite(stm).Put(newCommit.ID, newCommit); err != nil {
			return err
		}
	}
	return nil
}

func sizeChange(tree hashtree.HashTree, parentTree hashtree.HashTree) uint64 {
	if tree == nil {
		return 0 // output commit from a failed job -- will be ignored
	}
	if parentTree == nil {
		return uint64(tree.FSSize())
	}
	var result uint64
	tree.Diff(parentTree, "", "", -1, func(path string, node *hashtree.NodeProto, new bool) error {
		if node.FileNode != nil && new {
			result += uint64(node.SubtreeSize)
		}
		return nil
	})
	return result
}

// inspectCommit takes a Commit and returns the corresponding CommitInfo.
//
// As a side effect, this function also replaces the ID in the given commit
// with a real commit ID.
func (d *driver) inspectCommit(ctx context.Context, commit *pfs.Commit, blockState pfs.CommitState) (*pfs.CommitInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if commit == nil {
		return nil, fmt.Errorf("cannot inspect nil commit")
	}
	if err := d.checkIsAuthorized(pachClient, commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}

	// Check if the commitID is a branch name
	var commitInfo *pfs.CommitInfo
	if _, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		var err error
		commitInfo, err = d.resolveCommit(stm, commit)
		return err
	}); err != nil {
		return nil, err
	}

	commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
	if blockState == pfs.CommitState_READY {
		// Wait for each provenant commit to be finished
		for _, commit := range commitInfo.Provenance {
			d.inspectCommit(ctx, commit, pfs.CommitState_FINISHED)
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
						return fmt.Errorf("Unmarshal: %v", err)
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
		return nil, fmt.Errorf("cannot resolve nil commit")
	}
	commit := userCommit // back up user commit, for error reporting
	// Extract any ancestor tokens from 'commit.ID' (i.e. ~ and ^)
	var ancestryLength int
	commit.ID, ancestryLength = parseCommitID(commit.ID)

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
		commit.ID = branchInfo.Head.ID
	}

	// Traverse commits' parents until you've reached the right ancestor
	commits := d.commits(commit.Repo.Name).ReadWrite(stm)
	commitInfo := &pfs.CommitInfo{}
	for i := 0; i <= ancestryLength; i++ {
		if commit == nil {
			return nil, pfsserver.ErrCommitNotFound{userCommit}
		}
		childCommit := commit // preserve child for error reporting
		if err := commits.Get(commit.ID, commitInfo); err != nil {
			if col.IsErrNotFound(err) {
				if i == 0 {
					return nil, pfsserver.ErrCommitNotFound{childCommit}
				}
				return nil, pfsserver.ErrParentCommitNotFound{childCommit}
			}
			return nil, err
		}
		commit = commitInfo.ParentCommit
	}
	return commitInfo, nil
}

// parseCommitID accepts a commit ID that might contain the Git ancestry
// syntax, such as "master^2", "master~~", "master^^", "master~5", etc.
// It then returns the ID component such as "master" and the depth of the
// ancestor.  For instance, for "master^2" it'd return "master" and 2.
func parseCommitID(commitID string) (string, int) {
	sepIndex := strings.IndexAny(commitID, "^~")
	if sepIndex == -1 {
		return commitID, 0
	}

	// Find the separator, which is either "^" or "~"
	sep := commitID[sepIndex]
	strAfterSep := commitID[sepIndex+1:]

	// Try convert the string after the separator to an int.
	intAfterSep, err := strconv.Atoi(strAfterSep)
	// If it works, return
	if err == nil {
		return commitID[:sepIndex], intAfterSep
	}

	// Otherwise, we check if there's a sequence of separators, as in
	// "master^^^^" or "master~~~~"
	for i := sepIndex + 1; i < len(commitID); i++ {
		if commitID[i] != sep {
			// If we find a character that's not the separator, as in
			// "master~whatever", then we return.
			return commitID, 0
		}
	}

	// Here we've confirmed that the commit ID ends with a sequence of
	// (the same) separators and therefore uses the correct ancestry
	// syntax.
	return commitID[:sepIndex], len(commitID) - sepIndex
}

func (d *driver) listCommit(ctx context.Context, repo *pfs.Repo, to *pfs.Commit, from *pfs.Commit, number uint64) ([]*pfs.CommitInfo, error) {
	var result []*pfs.CommitInfo
	if err := d.listCommitF(ctx, repo, to, from, number, func(ci *pfs.CommitInfo) error {
		result = append(result, ci)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) listCommitF(ctx context.Context, repo *pfs.Repo, to *pfs.Commit, from *pfs.Commit, number uint64, f func(*pfs.CommitInfo) error) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, repo, auth.Scope_READER); err != nil {
		return err
	}
	if from != nil && from.Repo.Name != repo.Name || to != nil && to.Repo.Name != repo.Name {
		return fmt.Errorf("`from` and `to` commits need to be from repo %s", repo.Name)
	}

	// Make sure that the repo exists
	_, err := d.inspectRepo(ctx, repo, !includeAuth)
	if err != nil {
		return err
	}

	// Make sure that both from and to are valid commits
	if from != nil {
		_, err = d.inspectCommit(ctx, from, pfs.CommitState_STARTED)
		if err != nil {
			return err
		}
	}
	if to != nil {
		_, err = d.inspectCommit(ctx, to, pfs.CommitState_STARTED)
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
		return fmt.Errorf("cannot use `from` commit without `to` commit")
	} else if from == nil && to == nil {
		// if neither from and to is given, we list all commits in
		// the repo, sorted by revision timestamp
		if err := commits.List(ci, col.DefaultOptions, func(commitID string) error {
			if number <= 0 {
				return errutil.ErrBreak
			}
			number--
			return f(proto.Clone(ci).(*pfs.CommitInfo))
		}); err != nil {
			return err
		}
	} else {
		cursor := to
		for number != 0 && cursor != nil && (from == nil || cursor.ID != from.ID) {
			var commitInfo pfs.CommitInfo
			if err := commits.Get(cursor.ID, &commitInfo); err != nil {
				return err
			}
			if err := f(&commitInfo); err != nil {
				if err == errutil.ErrBreak {
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

func (d *driver) subscribeCommit(ctx context.Context, repo *pfs.Repo, branch string, from *pfs.Commit, state pfs.CommitState, f func(*pfs.CommitInfo) error) error {
	if from != nil && from.Repo.Name != repo.Name {
		return fmt.Errorf("the `from` commit needs to be from repo %s", repo.Name)
	}

	branches := d.branches(repo.Name).ReadOnly(ctx)
	newCommitWatcher, err := branches.WatchOne(branch)
	if err != nil {
		return err
	}
	defer newCommitWatcher.Close()
	// keep track of the commits that have been sent
	seen := make(map[string]bool)
	// include all commits that are currently on the given branch,
	commitInfos, err := d.listCommit(ctx, repo, client.NewCommit(repo.Name, branch), from, 0)
	if err != nil {
		// We skip NotFound error because it's ok if the branch
		// doesn't exist yet, in which case ListCommit returns
		// a NotFound error.
		if !isNotFoundErr(err) {
			return err
		}
	}
	// ListCommit returns commits in newest-first order,
	// but SubscribeCommit should return commit in oldest-first
	// order, so we reverse the order.
	for i := range commitInfos {
		commitInfo := commitInfos[len(commitInfos)-i-1]
		commitInfo, err := d.inspectCommit(ctx, commitInfo.Commit, state)
		if err != nil {
			return err
		}
		if err := f(commitInfo); err != nil {
			return err
		}
		seen[commitInfo.Commit.ID] = true
	}
	for {
		var branchName string
		branchInfo := &pfs.BranchInfo{}
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
			if err := event.Unmarshal(&branchName, branchInfo); err != nil {
				return fmt.Errorf("Unmarshal: %v", err)
			}
			if branchInfo.Head == nil {
				continue // put event == new branch was created. No commits yet though
			}

			// TODO we check the branchName because right now WatchOne, like all
			// collection watch commands, returns all events matching a given prefix,
			// which means we'll get back events associated with `master-v1` if we're
			// watching `master`.  Once this is changed we should remove the
			// comparison between branchName and branch.

			// We don't want to include the `from` commit itself
			if branchName == branch && (!(seen[branchInfo.Head.ID] || (from != nil && from.ID == branchInfo.Head.ID))) {
				commitInfo, err := d.inspectCommit(ctx, branchInfo.Head, state)
				if err != nil {
					return err
				}
				if err := f(commitInfo); err != nil {
					return err
				}
				seen[commitInfo.Commit.ID] = true
			}
		case watch.EventDelete:
			continue
		}
	}
}

func (d *driver) flushCommit(ctx context.Context, fromCommits []*pfs.Commit, toRepos []*pfs.Repo, f func(*pfs.CommitInfo) error) error {
	if len(fromCommits) == 0 {
		return fmt.Errorf("fromCommits cannot be empty")
	}

	// First compute intersection of the fromCommits subvenant commits, those
	// are the commits we're interested in. Iterate over all commits and keep a
	// running intersection (in commitsToWatch) of the subvenance of all commits
	// processed so far
	commitsToWatch := make(map[string]*pfs.Commit)
	for i, commit := range fromCommits {
		commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
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
	for _, commitToWatch := range commitsToWatch {
		if len(toRepoMap) > 0 {
			if _, ok := toRepoMap[commitToWatch.Repo.Name]; !ok {
				continue
			}
		}
		finishedCommitInfo, err := d.inspectCommit(ctx, commitToWatch, pfs.CommitState_FINISHED)
		if err != nil {
			if _, ok := err.(pfsserver.ErrCommitNotFound); ok {
				continue // just skip this
			}
			return err
		}
		if err := f(finishedCommitInfo); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) deleteCommit(ctx context.Context, userCommit *pfs.Commit) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, userCommit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	// Main txn: Delete all downstream commits, and update subvenance of upstream commits
	// TODO update branches inside this txn, by storing a repo's branches in its
	// RepoInfo or its HEAD commit
	deleted := make(map[string]*pfs.CommitInfo) // deleted commits
	affectedRepos := make(map[string]struct{})  // repos containing deleted commits
	deleteScratch := false                      // only delete scratch if txn succeeds
	if _, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		// 1) re-read CommitInfo inside txn
		userCommitInfo, err := d.resolveCommit(stm, userCommit)
		if err != nil {
			return fmt.Errorf("resolveCommit: %v", err)
		}
		deleteScratch = userCommitInfo.Finished == nil

		// 2) Define helper for deleting commits. 'lower' corresponds to
		// pfs.CommitRange.Lower, and is an ancestor of 'upper'
		deleteCommit := func(lower, upper *pfs.Commit) error {
			// Validate arguments
			if lower.Repo.Name != upper.Repo.Name {
				return fmt.Errorf("cannot delete commit range with mismatched repos \"%s\" and \"%s\"", lower.Repo.Name, upper.Repo.Name)
			}
			affectedRepos[lower.Repo.Name] = struct{}{}
			commits := d.commits(lower.Repo.Name).ReadWrite(stm)

			// delete commits on path upper -> ... -> lower (traverse ParentCommits)
			commit := upper
			for {
				if commit == nil {
					return fmt.Errorf("encountered nil parent commit in %s/%s...%s", lower.Repo.Name, lower.ID, upper.ID)
				}
				// Store commitInfo in 'deleted' and remove commit from etcd
				commitInfo := &pfs.CommitInfo{}
				if err := commits.Get(commit.ID, commitInfo); err != nil {
					return err
				}
				deleted[commit.ID] = commitInfo
				if err := commits.Delete(commit.ID); err != nil {
					return err
				}

				// Update repo size
				// TODO this is basically wrong. Other commits could share data with the
				// commit being removed, in which case we're subtracting too much.
				// We could also modify makeCommit and FinishCommit so that
				// commitInfo.SizeBytes stores incremental size (which could cause
				// commits to have negative sizes)
				repoInfo := &pfs.RepoInfo{}
				if err := d.repos.ReadWrite(stm).Update(commit.Repo.Name, repoInfo, func() error {
					repoInfo.SizeBytes -= commitInfo.SizeBytes
					return nil
				}); err != nil {
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
		if len(userCommitInfo.Provenance) > 0 {
			return fmt.Errorf("cannot delete the commit \"%s/%s\" because it has non-empty provenance", userCommit.Repo.Name, userCommit.ID)
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
			for _, provCommit := range deletedInfo.Provenance {
				// Check if we've fixed provCommit already (or if it's deleted and
				// doesn't need to be fixed
				if _, isDeleted := deleted[provCommit.ID]; isDeleted || visited[provCommit.ID] {
					continue
				}
				visited[provCommit.ID] = true

				// fix provCommit's subvenance
				provCI := &pfs.CommitInfo{}
				if err := d.commits(provCommit.Repo.Name).ReadWrite(stm).Update(provCommit.ID, provCI, func() error {
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
								if err := d.commits(subv.Lower.Repo.Name).ReadWrite(stm).Get(cur, curInfo); err != nil {
									return fmt.Errorf("error reading commitInfo for subvenant \"%s/%s\": %v", subv.Lower.Repo.Name, cur, err)
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
					return fmt.Errorf("err fixing subvenance of upstream commit %s/%s: %v", provCommit.Repo.Name, provCommit.ID, err)
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
			commits := d.commits(deletedInfo.Commit.Repo.Name).ReadWrite(stm)
			parent := lowestCommitInfo.ParentCommit
			for child := range liveChildren {
				commitInfo := &pfs.CommitInfo{}
				if err := commits.Update(child, commitInfo, func() error {
					commitInfo.ParentCommit = parent
					return nil
				}); err != nil {
					return fmt.Errorf("err updating child commit %v: %v", lowestCommitInfo.Commit, err)
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
					return fmt.Errorf("err rewriting children of ancestor commit %v: %v", lowestCommitInfo.Commit, err)
				}
			}
		}

		// 7) Traverse affected repos and rewrite all branches so that no branch
		// points to a deleted commit
		var shortestBranch *pfs.Branch
		var shortestBranchLen = maxInt
		for repo := range affectedRepos {
			repoInfo := &pfs.RepoInfo{}
			if err := d.repos.ReadWrite(stm).Get(repo, repoInfo); err != nil {
				return err
			}
			for _, brokenBranch := range repoInfo.Branches {
				// Traverse HEAD commit until we find a non-deleted parent or nil;
				// rewrite branch
				var branchInfo pfs.BranchInfo
				if err := d.branches(brokenBranch.Repo.Name).ReadWrite(stm).Update(brokenBranch.Name, &branchInfo, func() error {
					if len(branchInfo.Provenance) < shortestBranchLen {
						shortestBranchLen = len(branchInfo.Provenance)
						shortestBranch = branchInfo.Branch
					}
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
					return err
				}); err != nil && !col.IsErrNotFound(err) {
					// If err is NotFound, branch is in downstream provenance but
					// doesn't exist yet--nothing to update
					return fmt.Errorf("error updating branch %v/%v: %v", brokenBranch.Repo.Name, brokenBranch.Name, err)
				}
			}
		}

		// 8) propagate the changes to 'branch' and its subvenance. This may start
		// new HEAD commits downstream, if the new branch heads haven't been
		// processed yet
		return d.propagateCommit(stm, shortestBranch)
	}); err != nil {
		return fmt.Errorf("error rewriting commit graph: %v", err)
	}

	// Delete the scratch space for this commit
	// TODO put scratch spaces in a collection and do this in the txn above
	if deleteScratch {
		if _, err := d.etcdClient.Delete(ctx, d.scratchCommitPrefix(userCommit), etcd.WithPrefix()); err != nil {
			return err
		}
	}
	return nil
}

// createBranch creates a new branch or updates an existing branch (must be one
// or the other). Most importantly, it sets 'branch.DirectProvenance' to
// 'provenance' and then for all (downstream) branches, restores the invariant:
//   ∀ b . b.Provenance = ∪ b'.Provenance (where b' ∈ b.DirectProvenance)
//
// This invariant is assumed to hold for all branches upstream of 'branch', but not
// for 'branch' itself once 'b.Provenance' has been set.
func (d *driver) createBranch(ctx context.Context, branch *pfs.Branch, commit *pfs.Commit, provenance []*pfs.Branch) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	// Validate request. The request must do exactly one of:
	// 1) updating 'branch's provenance (commit is nil OR commit == branch)
	// 2) re-pointing 'branch' at a new commit
	if commit != nil {
		// Determine if this is a provenance update
		sameTarget := branch.Repo.Name == commit.Repo.Name && branch.Name == commit.ID
		if !sameTarget && provenance != nil {
			return fmt.Errorf("cannot point branch \"%s\" at target commit \"%s/%s\" without clearing its provenance",
				branch.Name, commit.Repo.Name, commit.ID)
		}
	}

	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		// if 'commit' is a branch, resolve it
		var err error
		if commit != nil {
			_, err = d.resolveCommit(stm, commit) // if 'commit' is a branch, resolve it
			if err != nil {
				// possible that branch exists but has no head commit. This is fine, but
				// branchInfo.Head must also be nil
				if !isNoHeadErr(err) {
					return fmt.Errorf("unable to inspect %s/%s: %v", err, commit.Repo.Name, commit.ID)
				}
				commit = nil
			}
		}

		// Retrieve (and create, if necessary) the current version of this branch
		branches := d.branches(branch.Repo.Name).ReadWrite(stm)
		branchInfo := &pfs.BranchInfo{}
		if err := branches.Upsert(branch.Name, branchInfo, func() error {
			branchInfo.Name = branch.Name // set in case 'branch' is new
			branchInfo.Branch = branch
			branchInfo.Head = commit
			branchInfo.DirectProvenance = nil
			for _, provBranch := range provenance {
				add(&branchInfo.DirectProvenance, provBranch)
			}
			return nil
		}); err != nil {
			return err
		}
		repos := d.repos.ReadWrite(stm)
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Update(branch.Repo.Name, repoInfo, func() error {
			add(&repoInfo.Branches, branch)
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
			if err := d.branches(subvBranch.Repo.Name).ReadWrite(stm).Get(subvBranch.Name, subvBranchInfo); err != nil {
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
				if err := d.addBranchProvenance(branchInfo, provBranch, stm); err != nil {
					return err
				}
				provBranchInfo := &pfs.BranchInfo{}
				if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Get(provBranch.Name, provBranchInfo); err != nil {
					return err
				}
				for _, provBranch := range provBranchInfo.Provenance {
					if err := d.addBranchProvenance(branchInfo, provBranch, stm); err != nil {
						return err
					}
				}
			}
			if err := d.branches(branchInfo.Branch.Repo.Name).ReadWrite(stm).Put(branchInfo.Branch.Name, branchInfo); err != nil {
				return err
			}
			// Update Subvenance of 'branchInfo's Provenance (incl. all Subvenance)
			for _, oldProvBranch := range oldProvenance {
				if !has(&branchInfo.Provenance, oldProvBranch) {
					// Provenance was deleted, so we delete ourselves from their subvenance
					oldProvBranchInfo := &pfs.BranchInfo{}
					if err := d.branches(oldProvBranch.Repo.Name).ReadWrite(stm).Update(oldProvBranch.Name, oldProvBranchInfo, func() error {
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
		return d.propagateCommit(stm, branch)
	})
	return err
}

func (d *driver) inspectBranch(ctx context.Context, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	result := &pfs.BranchInfo{}
	if err := d.branches(branch.Repo.Name).ReadOnly(ctx).Get(branch.Name, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) listBranch(ctx context.Context, repo *pfs.Repo) ([]*pfs.BranchInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	var result []*pfs.BranchInfo
	branchInfo := &pfs.BranchInfo{}
	branches := d.branches(repo.Name).ReadOnly(ctx)
	if err := branches.List(branchInfo, col.DefaultOptions, func(string) error {
		result = append(result, proto.Clone(branchInfo).(*pfs.BranchInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) deleteBranch(ctx context.Context, branch *pfs.Branch, force bool) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		return d.deleteBranchSTM(stm, branch, force)
	})
	return err
}

func (d *driver) deleteBranchSTM(stm col.STM, branch *pfs.Branch, force bool) error {
	branches := d.branches(branch.Repo.Name).ReadWrite(stm)
	branchInfo := &pfs.BranchInfo{}
	if err := branches.Get(branch.Name, branchInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return fmt.Errorf("branches.Get: %v", err)
		}
	}
	if branchInfo.Branch != nil {
		if !force {
			if len(branchInfo.Subvenance) > 0 {
				return fmt.Errorf("branch %s has %v as subvenance, deleting it would break those branches", branch.Name, branchInfo.Subvenance)
			}
		}
		if err := branches.Delete(branch.Name); err != nil {
			return fmt.Errorf("branches.Delete: %v", err)
		}
		for _, provBranch := range branchInfo.Provenance {
			provBranchInfo := &pfs.BranchInfo{}
			if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Update(provBranch.Name, provBranchInfo, func() error {
				del(&provBranchInfo.Subvenance, branch)
				return nil
			}); err != nil && !isNotFoundErr(err) {
				return fmt.Errorf("error deleting subvenance: %v", err)
			}
		}
	}
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadWrite(stm).Update(branch.Repo.Name, repoInfo, func() error {
		del(&repoInfo.Branches, branch)
		return nil
	}); err != nil {
		if !col.IsErrNotFound(err) || !force {
			return err
		}
	}
	return nil
}

func (d *driver) scratchPrefix() string {
	return path.Join(d.prefix, "scratch")
}

// scratchCommitPrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchCommitPrefix(commit *pfs.Commit) string {
	// TODO(msteffen) this doesn't currenty (2018-2-4) use d.scratchPrefix(),
	// but probably should? If this is changed, filepathFromEtcdPath will also
	// need to change.
	return path.Join(commit.Repo.Name, commit.ID)
}

// scratchFilePrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchFilePrefix(ctx context.Context, file *pfs.File) (string, error) {
	return path.Join(d.scratchCommitPrefix(file.Commit), file.Path), nil
}

func (d *driver) filePathFromEtcdPath(etcdPath string) string {
	etcdPath = strings.TrimPrefix(etcdPath, d.prefix)
	// etcdPath looks like /putFileRecords/repo/commit/path/to/file
	split := strings.Split(etcdPath, "/")
	// we only want /path/to/file so we use index 4 (note that there's an "" at
	// the beginning of the slice because of the lead /)
	return path.Join(split[4:]...)
}

func (d *driver) putFile(ctx context.Context, file *pfs.File, delimiter pfs.Delimiter,
	targetFileDatums int64, targetFileBytes int64, overwriteIndex *pfs.OverwriteIndex, reader io.Reader, header *pfs.Metadata, footer *pfs.Metadata) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	// oneOff is true if we're creating the commit as part of this put-file
	oneOff := false
	// inspectCommit will replace file.Commit.ID with an actual commit ID if
	// it's a branch. So we want to save it first.
	branch := ""
	if !uuid.IsUUIDWithoutDashes(file.Commit.ID) {
		branch = file.Commit.ID
	}
	commitInfo, err := d.inspectCommit(ctx, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		if (!isNotFoundErr(err) && !isNoHeadErr(err)) || branch == "" {
			return err
		}
		oneOff = true
	}
	if commitInfo != nil && commitInfo.Finished != nil {
		if branch == "" {
			return pfsserver.ErrCommitFinished{file.Commit}
		}
		oneOff = true
	}

	records := &pfs.PutFileRecords{}
	if overwriteIndex != nil && overwriteIndex.Index == 0 {
		records.Tombstone = true
	}
	if err := hashtree.ValidatePath(file.Path); err != nil {
		return err
	}

	if delimiter == pfs.Delimiter_NONE {
		objects, size, err := pachClient.PutObjectSplit(reader)
		if err != nil {
			return err
		}

		// Here we use the invariant that every one but the last object
		// should have a size of ChunkSize.
		for i, object := range objects {
			record := &pfs.PutFileRecord{
				ObjectHash: object.Hash,
			}

			if size > pfs.ChunkSize {
				record.SizeBytes = pfs.ChunkSize
			} else {
				record.SizeBytes = size
			}
			size -= pfs.ChunkSize

			// The first record takes care of the overwriting
			if i == 0 && overwriteIndex != nil && overwriteIndex.Index != 0 {
				record.OverwriteIndex = overwriteIndex
			}

			records.Records = append(records.Records, record)
		}
	} else {
		buffer := &bytes.Buffer{}
		var datumsWritten int64
		var bytesWritten int64
		var filesPut int
		EOF := false
		var eg errgroup.Group
		decoder := json.NewDecoder(reader)
		bufioR := bufio.NewReader(reader)
		sqlReader := sql.NewPGDumpReader(bufioR)

		indexToRecord := make(map[int]*pfs.PutFileRecord)
		var mu sync.Mutex
		limiter := limit.New(putObjectConcurrency)
		for !EOF {
			var err error
			var value []byte
			switch delimiter {
			case pfs.Delimiter_JSON:
				var jsonValue json.RawMessage
				err = decoder.Decode(&jsonValue)
				value = jsonValue
			case pfs.Delimiter_LINE:
				value, err = bufioR.ReadBytes('\n')
			case pfs.Delimiter_SQL:
				value, err = sqlReader.ReadRow()
				if err == io.EOF {
					headerBuffer := &bytes.Buffer{}
					headerBuffer.Write(sqlReader.Header)
					limiter.Acquire()
					eg.Go(func() error {
						defer limiter.Release()
						object, size, err := pachClient.PutObject(headerBuffer)
						if err != nil {
							return err
						}
						mu.Lock()
						defer mu.Unlock()
						records.Header = &pfs.PutFileRecord{
							SizeBytes:  size,
							ObjectHash: object.Hash,
						}
						return nil
					})
					footerBuffer := &bytes.Buffer{}
					footerBuffer.Write(sqlReader.Footer)
					limiter.Acquire()
					eg.Go(func() error {
						defer limiter.Release()
						object, size, err := pachClient.PutObject(footerBuffer)
						if err != nil {
							return err
						}
						mu.Lock()
						defer mu.Unlock()
						records.Footer = &pfs.PutFileRecord{
							SizeBytes:  size,
							ObjectHash: object.Hash,
						}
						return nil
					})
				}
			default:
				return fmt.Errorf("unrecognized delimiter %s", delimiter.String())
			}
			if err != nil {
				if err == io.EOF {
					EOF = true
				} else {
					return err
				}
			}
			buffer.Write(value)
			bytesWritten += int64(len(value))
			datumsWritten++
			if buffer.Len() != 0 &&
				((targetFileBytes != 0 && bytesWritten >= targetFileBytes) ||
					(targetFileDatums != 0 && datumsWritten >= targetFileDatums) ||
					(targetFileBytes == 0 && targetFileDatums == 0) ||
					EOF) {
				_buffer := buffer
				index := filesPut
				limiter.Acquire()
				eg.Go(func() error {
					defer limiter.Release()
					object, size, err := pachClient.PutObject(_buffer)
					if err != nil {
						return err
					}
					mu.Lock()
					defer mu.Unlock()
					indexToRecord[index] = &pfs.PutFileRecord{
						SizeBytes:  size,
						ObjectHash: object.Hash,
					}
					return nil
				})
				datumsWritten = 0
				bytesWritten = 0
				buffer = &bytes.Buffer{}
				filesPut++
			}
		}
		if err := eg.Wait(); err != nil {
			return err
		}

		records.Split = true
		for i := 0; i < len(indexToRecord); i++ {
			records.Records = append(records.Records, indexToRecord[i])
		}
	}
	// NOTE: i dont use limiters here for simplicity...
	var mu sync.Mutex
	var eg errgroup.Group
	if header != nil {
		if len(header.Value) == 0 {
			records.Header = &pfs.PutFileRecord{}
		} else {
			eg.Go(func() error {
				object, size, err := pachClient.PutObject(bytes.NewReader(header.Value))
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				records.Header = &pfs.PutFileRecord{
					SizeBytes:  size,
					ObjectHash: object.Hash,
				}
				return nil
			})
		}
	}
	if footer != nil {
		if len(footer.Value) == 0 {
			records.Footer = &pfs.PutFileRecord{}
		} else {
			eg.Go(func() error {
				object, size, err := pachClient.PutObject(bytes.NewReader(footer.Value))
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				records.Footer = &pfs.PutFileRecord{
					SizeBytes:  size,
					ObjectHash: object.Hash,
				}
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	if oneOff {
		// oneOff puts only work on branches, so we know branch != "". We pass
		// a commit with no ID, that ID will be filled in with the head of
		// branch (if it exists).
		_, err := d.makeCommit(ctx, "", client.NewCommit(file.Commit.Repo.Name, ""), branch, nil, nil, []string{file.Path}, []*pfs.PutFileRecords{records}, "")
		return err
	}
	return d.upsertPutFileRecords(ctx, file, records)
}

func (d *driver) copyFile(ctx context.Context, src *pfs.File, dst *pfs.File, overwrite bool) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, src.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	if err := d.checkIsAuthorized(pachClient, dst.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	if err := hashtree.ValidatePath(dst.Path); err != nil {
		return err
	}
	branch := ""
	if !uuid.IsUUIDWithoutDashes(dst.Commit.ID) {
		branch = dst.Commit.ID
	}
	ci, err := d.inspectCommit(ctx, dst.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if ci.Finished != nil && branch == "" {
		return pfsserver.ErrCommitFinished{dst.Commit}
	}
	var paths []string
	var records []*pfs.PutFileRecords
	if overwrite {
		if ci.Finished == nil {
			if err := d.deleteFile(ctx, dst); err != nil {
				return err
			}
		} else {
			paths = append(paths, dst.Path)
			records = append(records, &pfs.PutFileRecords{Tombstone: true})
		}
	}
	srcTree, err := d.getTreeForFile(ctx, src)
	if err != nil {
		return err
	}
	// This is necessary so we can call filepath.Rel below
	if !strings.HasPrefix(src.Path, "/") {
		src.Path = "/" + src.Path
	}
	var recordsMu sync.Mutex
	var eg errgroup.Group
	if err := srcTree.Walk(src.Path, func(walkPath string, node *hashtree.NodeProto) error {
		if node.FileNode == nil {
			return nil
		}
		eg.Go(func() error {
			relPath, err := filepath.Rel(src.Path, walkPath)
			if err != nil {
				// This shouldn't be possible
				return fmt.Errorf("error from filepath.Rel: %+v (this is likely a bug)", err)
			}
			record := &pfs.PutFileRecords{}
			file := client.NewFile(dst.Commit.Repo.Name, dst.Commit.ID, path.Clean(path.Join(dst.Path, relPath)))
			for i, object := range node.FileNode.Objects {
				var size int64
				if i == 0 {
					size = node.SubtreeSize
				}
				record.Records = append(record.Records, &pfs.PutFileRecord{
					SizeBytes:  size,
					ObjectHash: object.Hash,
				})
			}
			if ci.Finished == nil {
				return d.upsertPutFileRecords(ctx, file, record)
			}
			recordsMu.Lock()
			defer recordsMu.Unlock()
			paths = append(paths, file.Path)
			records = append(records, record)
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	if ci.Finished != nil {
		_, err = d.makeCommit(ctx, "", client.NewCommit(dst.Commit.Repo.Name, ""), branch, nil, nil, paths, records, "")
		return err
	}
	return nil
}

func (d *driver) getTreeForCommit(ctx context.Context, commit *pfs.Commit) (hashtree.HashTree, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if commit == nil || commit.ID == "" {
		t, err := hashtree.NewHashTree().Finish()
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	tree, ok := d.treeCache.Get(commit.ID)
	if ok {
		h, ok := tree.(hashtree.HashTree)
		if ok {
			return h, nil
		}
		return nil, fmt.Errorf("corrupted cache: expected hashtree.Hashtree, found %v", tree)
	}

	if _, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED); err != nil {
		return nil, err
	}

	commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
	commitInfo := &pfs.CommitInfo{}
	if err := commits.Get(commit.ID, commitInfo); err != nil {
		return nil, err
	}
	if commitInfo.Finished == nil {
		return nil, fmt.Errorf("cannot read from an open commit")
	}
	treeRef := commitInfo.Tree

	if treeRef == nil {
		t, err := hashtree.NewHashTree().Finish()
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	// read the tree from the block store
	var buf bytes.Buffer
	if err := pachClient.GetObject(treeRef.Hash, &buf); err != nil {
		return nil, err
	}

	h, err := hashtree.Deserialize(buf.Bytes())
	if err != nil {
		return nil, err
	}

	d.treeCache.Add(commit.ID, h)

	return h, nil
}

// getTreeForFile is like getTreeForCommit except that it can handle open commits.
// It takes a file instead of a commit so that it can apply the changes for
// that path to the tree before it returns it.
func (d *driver) getTreeForFile(ctx context.Context, file *pfs.File) (hashtree.HashTree, error) {
	if file.Commit == nil {
		t, err := hashtree.NewHashTree().Finish()
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	commitInfo, err := d.inspectCommit(ctx, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished != nil {
		tree, err := d.getTreeForCommit(ctx, file.Commit)
		if err != nil {
			return nil, err
		}
		return tree, nil
	}
	parentTree, err := d.getTreeForCommit(ctx, commitInfo.ParentCommit)
	if err != nil {
		return nil, err
	}
	return d.getTreeForOpenCommit(ctx, file, parentTree)
}

func (d *driver) getTreeForOpenCommit(ctx context.Context, file *pfs.File, parentTree hashtree.HashTree) (hashtree.HashTree, error) {
	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return nil, err
	}
	var tree hashtree.OpenHashTree
	if _, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		tree = parentTree.Open()
		recordsCol := d.putFileRecords.ReadOnly(ctx)
		putFileRecords := &pfs.PutFileRecords{}
		opts := &col.Options{etcd.SortByModRevision, etcd.SortAscend, true}
		return recordsCol.ListPrefix(prefix, putFileRecords, opts, func(key string) error {
			return d.applyWrite(path.Join(file.Path, key), putFileRecords, tree)
		})
	}); err != nil {
		return nil, err
	}
	return tree.Finish()
}

func (d *driver) getFile(ctx context.Context, file *pfs.File, offset int64, size int64) (io.Reader, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}

	tree, err := d.getTreeForFile(ctx, client.NewFile(file.Commit.Repo.Name, file.Commit.ID, ""))
	if err != nil {
		return nil, err
	}
	paths, err := tree.Glob(file.Path)
	if err != nil {
		return nil, err
	}
	var objects []*pfs.Object
	var totalSize int64
	var sortedPaths []string
	foundDirectoryNode := false
	for path, node := range paths {
		sortedPaths = append(sortedPaths, path)
		if node.DirNode != nil {
			foundDirectoryNode = true
		}
	}
	if !foundDirectoryNode {
		// e.g. this was a request for a single file e.g. /foo/bar.txt
		dirNodePaths, err := tree.Glob(filepath.Dir(file.Path))
		if err != nil {
			return nil, err
		}
		if len(dirNodePaths) != 1 {
			return nil, fmt.Errorf("invalid set of directory node paths - found %v expected 1 - this is probably a bug", len(dirNodePaths))
		}
		for key, node := range dirNodePaths {
			paths[key] = node
			sortedPaths = append(sortedPaths, key)
		}
	}
	sort.Strings(sortedPaths)
	footers := stack.New()
	directories := stack.New()
	for _, path := range sortedPaths {
		node := paths[path]
		thisDir := directories.Peek()
		if thisDir != nil && !strings.HasPrefix(path, thisDir.(string)) {
			// We've proceeded past the current directory
			footer := footers.Pop().(*pfs.Object)
			if footer != nil {
				objects = append(objects, footer)
			}
			directories.Pop()
		}

		if node.DirNode != nil {
			header := node.DirNode.Header
			if header != nil {
				objects = append(objects, header)
			}
			footers.Push(node.DirNode.Footer)
			directories.Push(path + "/") // Need trailing slash to differentiate dir from other lexigraphical matches
		} else {
			objects = append(objects, node.FileNode.Objects...)
		}

		totalSize += node.SubtreeSize

	}
	for footers.Len() > 0 {
		footer := footers.Pop().(*pfs.Object)
		if footer != nil {
			objects = append(objects, footer)
		}
	}

	if len(paths) <= 0 {
		return nil, fmt.Errorf("no file(s) found that match %v", file.Path)
	}

	if foundDirectoryNode && len(objects) == 0 {
		// Dir has no header/footer, return err
		return nil, fmt.Errorf("cannot read directory, no header or footer")
	}

	getObjectsClient, err := pachClient.ObjectAPIClient.GetObjects(
		ctx,
		&pfs.GetObjectsRequest{
			Objects:     objects,
			OffsetBytes: uint64(offset),
			SizeBytes:   uint64(size),
			TotalSize:   uint64(totalSize),
		})
	if err != nil {
		return nil, err
	}

	return grpcutil.NewStreamingBytesReader(getObjectsClient), nil
}

// If full is false, exclude potentially large fields such as `Objects`
// and `Children`
func nodeToFileInfo(commit *pfs.Commit, path string, node *hashtree.NodeProto, full bool) *pfs.FileInfo {
	fileInfo := &pfs.FileInfo{
		File: &pfs.File{
			Commit: commit,
			Path:   path,
		},
		SizeBytes: uint64(node.SubtreeSize),
		Hash:      node.Hash,
	}
	if node.FileNode != nil {
		fileInfo.FileType = pfs.FileType_FILE
		if full {
			fileInfo.Objects = node.FileNode.Objects
		}
	} else if node.DirNode != nil {
		fileInfo.FileType = pfs.FileType_DIR
		if full {
			fileInfo.Children = node.DirNode.Children
		}
	}
	return fileInfo
}

func (d *driver) inspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	tree, err := d.getTreeForFile(ctx, file)
	if err != nil {
		return nil, err
	}

	node, err := tree.Get(file.Path)
	if err != nil {
		return nil, pfsserver.ErrFileNotFound{file}
	}

	// TODO(bryce) Add file path cleaning to this return
	return nodeToFileInfo(file.Commit, file.Path, node, true), nil
}

func (d *driver) listFile(ctx context.Context, file *pfs.File, full bool) ([]*pfs.FileInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}

	tree, err := d.getTreeForFile(ctx, client.NewFile(file.Commit.Repo.Name, file.Commit.ID, ""))
	if err != nil {
		return nil, err
	}

	rootPaths, err := tree.Glob(file.Path)
	if err != nil {
		return nil, err
	}
	seenPaths := make(map[string]struct{}, len(rootPaths))
	var fileInfos []*pfs.FileInfo
	for rootPath, rootNode := range rootPaths {
		nodes, err := tree.List(rootPath)
		if err != nil {
			if hashtree.Code(err) == hashtree.PathConflict {
				if _, ok := seenPaths[rootPath]; !ok {
					fileInfos = append(fileInfos, nodeToFileInfo(file.Commit, rootPath, rootNode, full))
					seenPaths[rootPath] = struct{}{}
				}
				continue
			}
			return nil, err
		}
		for _, node := range nodes {
			path := filepath.Join(rootPath, node.Name)
			if _, ok := seenPaths[path]; !ok {
				fileInfos = append(fileInfos, nodeToFileInfo(file.Commit, path, node, full))
				seenPaths[path] = struct{}{}
			}
		}
	}
	return fileInfos, nil
}

func (d *driver) globFile(ctx context.Context, commit *pfs.Commit, pattern string) ([]*pfs.FileInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}

	tree, err := d.getTreeForFile(ctx, client.NewFile(commit.Repo.Name, commit.ID, ""))
	if err != nil {
		return nil, err
	}

	paths, err := tree.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for path, node := range paths {
		fileInfos = append(fileInfos, nodeToFileInfo(commit, path, node, false))
	}
	return fileInfos, nil
}

func (d *driver) diffFile(ctx context.Context, newFile *pfs.File, oldFile *pfs.File, shallow bool) ([]*pfs.FileInfo, []*pfs.FileInfo, error) {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	// Do READER authorization check for both newFile and oldFile
	if oldFile != nil && oldFile.Commit != nil {
		//	if oldFile != nil {
		if err := d.checkIsAuthorized(pachClient, oldFile.Commit.Repo, auth.Scope_READER); err != nil {
			return nil, nil, err
		}
	}
	if newFile != nil && newFile.Commit != nil {
		//	if newFile != nil {
		if err := d.checkIsAuthorized(pachClient, newFile.Commit.Repo, auth.Scope_READER); err != nil {
			return nil, nil, err
		}
	}
	newTree, err := d.getTreeForFile(ctx, newFile)
	if err != nil {
		return nil, nil, err
	}
	// if oldFile is new we use the parent of newFile
	if oldFile == nil {
		oldFile = &pfs.File{}
		newCommitInfo, err := d.inspectCommit(ctx, newFile.Commit, pfs.CommitState_STARTED)
		if err != nil {
			return nil, nil, err
		}
		// ParentCommit may be nil, that's fine because getTreeForCommit
		// handles nil
		oldFile.Commit = newCommitInfo.ParentCommit
		oldFile.Path = newFile.Path
	}
	oldTree, err := d.getTreeForFile(ctx, oldFile)
	if err != nil {
		return nil, nil, err
	}
	var newFileInfos []*pfs.FileInfo
	var oldFileInfos []*pfs.FileInfo
	recursiveDepth := -1
	if shallow {
		recursiveDepth = 1
	}
	if err := newTree.Diff(oldTree, newFile.Path, oldFile.Path, int64(recursiveDepth), func(path string, node *hashtree.NodeProto, new bool) error {
		if new {
			newFileInfos = append(newFileInfos, nodeToFileInfo(newFile.Commit, path, node, false))
		} else {
			oldFileInfos = append(oldFileInfos, nodeToFileInfo(oldFile.Commit, path, node, false))
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return newFileInfos, oldFileInfos, nil
}

func (d *driver) deleteFile(ctx context.Context, file *pfs.File) error {
	pachClient := d.getPachClient(ctx)
	ctx = pachClient.Ctx()
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	branch := ""
	if !uuid.IsUUIDWithoutDashes(file.Commit.ID) {
		branch = file.Commit.ID
	}
	commitInfo, err := d.inspectCommit(ctx, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		if branch == "" {
			return pfsserver.ErrCommitFinished{file.Commit}
		}
		_, err := d.makeCommit(ctx, "", client.NewCommit(file.Commit.Repo.Name, ""), branch, nil, nil, []string{file.Path}, []*pfs.PutFileRecords{&pfs.PutFileRecords{Tombstone: true}}, "")
		return err
	}
	return d.upsertPutFileRecords(ctx, file, &pfs.PutFileRecords{Tombstone: true})
}

func (d *driver) deleteAll(ctx context.Context) error {
	// Note: d.listRepo() doesn't return the 'spec' repo, so it doesn't get
	// deleted here. Instead, PPS is responsible for deleting and re-creating it
	repoInfos, err := d.listRepo(ctx, !includeAuth)
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		if err := d.deleteRepo(ctx, repoInfo.Repo, true); err != nil && !auth.IsErrNotAuthorized(err) {
			return err
		}
	}
	return nil
}

// Put the tree into the blob store
// Only write the records to etcd if the commit does exist and is open.
// To check that a key exists in etcd, we assert that its CreateRevision
// is greater than zero.
func (d *driver) upsertPutFileRecords(ctx context.Context, file *pfs.File, newRecords *pfs.PutFileRecords) error {
	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		commitsCol := d.openCommits.ReadOnly(ctx)
		var commit pfs.Commit
		err := commitsCol.Get(file.Commit.ID, &commit)
		if err != nil {
			return err
		}
		// Dumb check to make sure the unmarshalled value exists (and matches the current ID)
		// to denote that the current commit is indeed open
		if commit.ID != file.Commit.ID {
			return fmt.Errorf("commit %v is not open", file.Commit.ID)
		}
		recordsCol := d.putFileRecords.ReadWrite(stm)
		var existingRecords pfs.PutFileRecords
		return recordsCol.Upsert(prefix, &existingRecords, func() error {
			if newRecords.Tombstone {
				existingRecords.Tombstone = true
				existingRecords.Records = nil
			}
			existingRecords.Split = newRecords.Split
			existingRecords.Records = append(existingRecords.Records, newRecords.Records...)
			existingRecords.Header = newRecords.Header
			existingRecords.Footer = newRecords.Footer
			return nil
		})
	})
	if err != nil {
		return err
	}

	return err
}

func (d *driver) applyWrite(key string, records *pfs.PutFileRecords, tree hashtree.OpenHashTree) error {
	// a map that keeps track of the sizes of objects
	sizeMap := make(map[string]int64)

	if records.Tombstone {
		if err := tree.DeleteFile(key); err != nil {
			return err
		}
	}
	if !records.Split {
		if len(records.Records) == 0 {
			return nil
		}
		for _, record := range records.Records {
			sizeMap[record.ObjectHash] = record.SizeBytes
			if record.OverwriteIndex != nil {
				// Computing size delta
				delta := record.SizeBytes
				fileNode, err := tree.Get(key)
				if err == nil {
					// If we can't find the file, that's fine.
					for i := record.OverwriteIndex.Index; int(i) < len(fileNode.FileNode.Objects); i++ {
						delta -= sizeMap[fileNode.FileNode.Objects[i].Hash]
					}
				}

				if err := tree.PutFileOverwrite(key, []*pfs.Object{{Hash: record.ObjectHash}}, record.OverwriteIndex, delta); err != nil {
					return err
				}
			} else {
				if err := tree.PutFile(key, []*pfs.Object{{Hash: record.ObjectHash}}, record.SizeBytes); err != nil {
					return err
				}
			}
		}
	} else {
		nodes, err := tree.List(key)
		if err != nil && hashtree.Code(err) != hashtree.PathNotFound {
			return err
		}
		var indexOffset int64
		if len(nodes) > 0 {
			indexOffset, err = strconv.ParseInt(path.Base(nodes[len(nodes)-1].Name), splitSuffixBase, splitSuffixWidth)
			if err != nil {
				return fmt.Errorf("error parsing filename %s as int, this likely means you're "+
					"using split on a directory which contains other data that wasn't put with split",
					path.Base(nodes[len(nodes)-1].Name))
			}
			indexOffset++ // start writing to the file after the last file
		}
		var header *pfs.Object
		var footer *pfs.Object
		headerFooterSize := int64(0)
		emptyRecord := pfs.PutFileRecord{}
		if records.Header != nil {
			if *records.Header == emptyRecord {
				header = &pfs.Object{}
			} else {
				header = &pfs.Object{Hash: records.Header.ObjectHash}
				headerFooterSize += records.Header.SizeBytes
			}
		}
		if records.Footer != nil {
			if *records.Footer == emptyRecord {
				footer = &pfs.Object{}
			} else {
				footer = &pfs.Object{Hash: records.Footer.ObjectHash}
				headerFooterSize += records.Footer.SizeBytes
			}
		}
		if err := tree.PutFileSplit(
			key,
			nil,
			0,
			header,
			footer,
			headerFooterSize,
			false,
		); err != nil {
			return err
		}
		nodes, err = tree.List(key)
		for i, record := range records.Records {
			if err := tree.PutFileSplit(
				path.Join(key, fmt.Sprintf(splitSuffixFmt, i+int(indexOffset))),
				[]*pfs.Object{{Hash: record.ObjectHash}},
				record.SizeBytes,
				header,
				footer,
				headerFooterSize,
				false,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func isNoHeadErr(err error) bool {
	_, ok := err.(pfsserver.ErrNoHead)
	return ok
}

func commitKey(commit *pfs.Commit) string {
	return fmt.Sprintf("%s/%s", commit.Repo.Name, commit.ID)
}

func branchKey(branch *pfs.Branch) string {
	return fmt.Sprintf("%s/%s", branch.Repo.Name, branch.Name)
}

func (d *driver) addBranchProvenance(branchInfo *pfs.BranchInfo, provBranch *pfs.Branch, stm col.STM) error {
	if provBranch.Repo.Name == branchInfo.Branch.Repo.Name && provBranch.Name == branchInfo.Branch.Name {
		return fmt.Errorf("provenance loop, branch %s/%s cannot be provenance for itself", provBranch.Repo.Name, provBranch.Name)
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

type branchCommit struct {
	commit *pfs.Commit
	branch *pfs.Branch
}

func appendSubvenance(commitInfo *pfs.CommitInfo, subvCommitInfo *pfs.CommitInfo) {
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
}

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

func search(bs *[]*pfs.Branch, branch *pfs.Branch) (int, bool) {
	return (*branchSet)(bs).search(branch)
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
