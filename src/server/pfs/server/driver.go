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

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/golang-lru"
	"google.golang.org/grpc"
)

const (
	splitSuffixBase  = 16
	splitSuffixWidth = 64
	splitSuffixFmt   = "%016x"

	// Makes calls to ListRepo and InspectRepo more legible
	includeAuth = true
)

// ValidateRepoName determines if a repo name is valid
func ValidateRepoName(name string) error {
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
	// address, and pachConn are used to connect back to Pachd's Object Store API
	// and Authorization API
	address      string
	pachConnOnce sync.Once
	onceErr      error
	pachConn     *grpc.ClientConn

	// pachClient is a cached Pachd client, that connects to Pachyderm's object
	// store API and auth API
	pachClient *client.APIClient

	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	prefix     string

	// collections
	repos         col.Collection
	repoRefCounts col.Collection
	commits       collectionFactory
	branches      collectionFactory
	openCommits   col.Collection

	// a cache for hashtrees
	treeCache *lru.Cache
}

const (
	tombstone = "delete"
)

const (
	defaultTreeCacheSize = 128
)

// newDriver is used to create a new Driver instance
func newDriver(address string, etcdAddresses []string, etcdPrefix string, treeCacheSize int64) (*driver, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddresses,
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to etcd: %v", err)
	}
	if treeCacheSize <= 0 {
		treeCacheSize = defaultTreeCacheSize
	}
	treeCache, err := lru.New(int(treeCacheSize))
	if err != nil {
		return nil, fmt.Errorf("could not initialize treeCache: %v", err)
	}

	d := &driver{
		address:       address,
		etcdClient:    etcdClient,
		prefix:        etcdPrefix,
		repos:         pfsdb.Repos(etcdClient, etcdPrefix),
		repoRefCounts: pfsdb.RepoRefCounts(etcdClient, etcdPrefix),
		commits: func(repo string) col.Collection {
			return pfsdb.Commits(etcdClient, etcdPrefix, repo)
		},
		branches: func(repo string) col.Collection {
			return pfsdb.Branches(etcdClient, etcdPrefix, repo)
		},
		openCommits: pfsdb.OpenCommits(etcdClient, etcdPrefix),
		treeCache:   treeCache,
	}
	go func() { d.initializePachConn() }() // Begin dialing connection on startup
	return d, nil
}

// newLocalDriver creates a driver using an local etcd instance.  This
// function is intended for testing purposes
func newLocalDriver(blockAddress string, etcdPrefix string) (*driver, error) {
	return newDriver(blockAddress, []string{"localhost:32379"}, etcdPrefix, defaultTreeCacheSize)
}

// initializePachConn initializes the connects that the pfs driver has with the
// Pachyderm object API and auth API, and blocks until the connection is
// established
//
// TODO(msteffen): client initialization (both etcd and pachd) might be better
// placed happen in server.go, near main(), so that we only pay the dial cost
// once, and so that pps doesn't need to have its own initialization code
func (d *driver) initializePachConn() error {
	d.pachConnOnce.Do(func() {
		d.pachConn, d.onceErr = grpc.Dial(d.address, client.PachDialOptions()...)
		d.pachClient = &client.APIClient{
			AuthAPIClient:   auth.NewAPIClient(d.pachConn),
			ObjectAPIClient: pfs.NewObjectAPIClient(d.pachConn),
		}
	})
	return d.onceErr
}

// checkIsAuthorized returns an error if the current user (in 'ctx') has
// authorization scope 's' for repo 'r'
func (d *driver) checkIsAuthorized(ctx context.Context, r *pfs.Repo, s auth.Scope) error {
	d.initializePachConn()
	resp, err := d.pachClient.AuthAPIClient.Authorize(auth.In2Out(ctx), &auth.AuthorizeRequest{
		Repo:  r.Name,
		Scope: s,
	})
	if err == nil && !resp.Authorized {
		return &auth.NotAuthorizedError{Repo: r.Name, Required: s}
	} else if err != nil && !auth.IsNotActivatedError(err) {
		return fmt.Errorf("error during authorization check for operation on \"%s\": %v",
			r.Name, grpcutil.ScrubGRPC(err))
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

func (d *driver) createRepo(ctx context.Context, repo *pfs.Repo, provenance []*pfs.Repo, description string, update bool) error {
	if err := ValidateRepoName(repo.Name); err != nil {
		return err
	}
	if err := d.initializePachConn(); err != nil {
		return err
	}
	if update {
		return d.updateRepo(ctx, repo, provenance, description)
	}

	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoRefCounts := d.repoRefCounts.ReadWriteInt(stm)

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
		whoAmI, err := d.pachClient.AuthAPIClient.WhoAmI(auth.In2Out(ctx),
			&auth.WhoAmIRequest{})
		if err != nil && !auth.IsNotActivatedError(err) {
			return fmt.Errorf("error while creating repo \"%s\": %v",
				repo.Name, grpcutil.ScrubGRPC(err))
		} else if err == nil {
			// auth is active, and user is logged in. Make user an owner of the new
			// repo (and clear any existing ACL under this name that might have been
			// created by accident)
			_, err := d.pachClient.AuthAPIClient.SetACL(auth.In2Out(ctx), &auth.SetACLRequest{
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

		// compute the full provenance of this repo
		fullProv := make(map[string]bool)
		for _, prov := range provenance {
			fullProv[prov.Name] = true
			provRepo := new(pfs.RepoInfo)
			if err := repos.Get(prov.Name, provRepo); err != nil {
				return err
			}
			// the provenance of my provenance is my provenance
			for _, prov := range provRepo.Provenance {
				fullProv[prov.Name] = true
			}
		}

		var fullProvRepos []*pfs.Repo
		for prov := range fullProv {
			fullProvRepos = append(fullProvRepos, &pfs.Repo{prov})
			if err := repoRefCounts.Increment(prov); err != nil {
				return err
			}
		}
		if err := repoRefCounts.Create(repo.Name, 0); err != nil {
			return err
		}
		repoInfo := &pfs.RepoInfo{
			Repo:        repo,
			Created:     now(),
			Provenance:  fullProvRepos,
			Description: description,
		}
		return repos.Create(repo.Name, repoInfo)
	})
	return err
}

func (d *driver) updateRepo(ctx context.Context, repo *pfs.Repo, provenance []*pfs.Repo, description string) error {
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoRefCounts := d.repoRefCounts.ReadWriteInt(stm)

		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(repo.Name, repoInfo); err != nil {
			return fmt.Errorf("error updating repo: %v", err)
		}
		// Caller only neads to be a WRITER to call UpdatePipeline(), so caller only
		// needs to be a WRITER to update the provenance.
		if err := d.checkIsAuthorized(ctx, repo, auth.Scope_WRITER); err != nil {
			return err
		}

		provToAdd := make(map[string]bool)
		provToRemove := make(map[string]bool)
		for _, newProv := range provenance {
			provToAdd[newProv.Name] = true
		}
		for _, oldProv := range repoInfo.Provenance {
			delete(provToAdd, oldProv.Name)
			provToRemove[oldProv.Name] = true
		}
		for _, newProv := range provenance {
			delete(provToRemove, newProv.Name)
		}

		// For each new provenance repo, we increase its ref count
		// by N where N is this repo's ref count.
		// For each old provenance repo we do the opposite.
		myRefCount, err := repoRefCounts.Get(repo.Name)
		if err != nil {
			return err
		}
		// +1 because we need to include ourselves.
		myRefCount++

		for newProv := range provToAdd {
			if err := repoRefCounts.IncrementBy(newProv, myRefCount); err != nil {
				return err
			}
		}

		for oldProv := range provToRemove {
			if err := repoRefCounts.DecrementBy(oldProv, myRefCount); err != nil {
				return err
			}
		}

		// We also add the new provenance repos to the provenance
		// of all downstream repos, and remove the old provenance
		// repos from their provenance.
		downstreamRepos, err := d.listRepo(ctx, []*pfs.Repo{repo}, !includeAuth)
		if err != nil {
			return err
		}

		for _, repoInfo := range downstreamRepos.RepoInfo {
		nextNewProv:
			for newProv := range provToAdd {
				for _, prov := range repoInfo.Provenance {
					if newProv == prov.Name {
						continue nextNewProv
					}
				}
				repoInfo.Provenance = append(repoInfo.Provenance, &pfs.Repo{newProv})
			}
		nextOldProv:
			for oldProv := range provToRemove {
				for i, prov := range repoInfo.Provenance {
					if oldProv == prov.Name {
						repoInfo.Provenance = append(repoInfo.Provenance[:i], repoInfo.Provenance[i+1:]...)
						continue nextOldProv
					}
				}
			}
			repos.Put(repoInfo.Repo.Name, repoInfo)
		}

		repoInfo.Description = description
		repoInfo.Provenance = provenance
		repos.Put(repo.Name, repoInfo)
		return nil
	})
	return err
}

func (d *driver) inspectRepo(ctx context.Context, repo *pfs.Repo, includeAuth bool) (*pfs.RepoInfo, error) {
	d.initializePachConn()
	result := &pfs.RepoInfo{}
	if err := d.repos.ReadOnly(ctx).Get(repo.Name, result); err != nil {
		return nil, err
	}
	if includeAuth {
		accessLevel, err := d.getAccessLevel(ctx, repo)
		if err != nil {
			if auth.IsNotActivatedError(err) {
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
	who, err := d.pachClient.AuthAPIClient.WhoAmI(auth.In2Out(ctx),
		&auth.WhoAmIRequest{})
	if err != nil {
		return auth.Scope_NONE, err
	}
	if who.IsAdmin {
		return auth.Scope_OWNER, nil
	}
	resp, err := d.pachClient.AuthAPIClient.GetScope(auth.In2Out(ctx),
		&auth.GetScopeRequest{Repos: []string{repo.Name}})
	if err != nil {
		return auth.Scope_NONE, err
	}
	if len(resp.Scopes) != 1 {
		return auth.Scope_NONE, fmt.Errorf("too many results from GetScope: %#v", resp)
	}
	return resp.Scopes[0], nil
}

func (d *driver) listRepo(ctx context.Context, provenance []*pfs.Repo, includeAuth bool) (*pfs.ListRepoResponse, error) {
	repos := d.repos.ReadOnly(ctx)
	// Ensure that all provenance repos exist
	for _, prov := range provenance {
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(prov.Name, repoInfo); err != nil {
			return nil, err
		}
	}

	iterator, err := repos.List()
	if err != nil {
		return nil, err
	}
	result := new(pfs.ListRepoResponse)
	authSeemsActive := true
nextRepo:
	for {
		repoName, repoInfo := "", new(pfs.RepoInfo)
		ok, err := iterator.Next(&repoName, repoInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if repoName == ppsconsts.SpecRepo {
			continue nextRepo
		}
		// A repo needs to have *all* the given repos as provenance
		// in order to be included in the result.
		for _, reqProv := range provenance {
			var matched bool
			for _, prov := range repoInfo.Provenance {
				if reqProv.Name == prov.Name {
					matched = true
				}
			}
			if !matched {
				continue nextRepo
			}
		}
		if includeAuth && authSeemsActive {
			accessLevel, err := d.getAccessLevel(ctx, repoInfo.Repo)
			if err == nil {
				repoInfo.AuthInfo = &pfs.RepoAuthInfo{AccessLevel: accessLevel}
			} else if auth.IsNotActivatedError(err) {
				authSeemsActive = false
			} else {
				return nil, fmt.Errorf("error getting access level for \"%s\": %v",
					repoName, grpcutil.ScrubGRPC(err))
			}
		}
		result.RepoInfo = append(result.RepoInfo, repoInfo)
	}
	return result, nil
}

func (d *driver) deleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	// TODO(msteffen): Fix d.deleteAll() so that it doesn't need to delete and
	// recreate the PPS spec repo, then uncomment this block to prevent users from
	// deleting it and breaking their cluster
	// if repo.Name == ppsconsts.SpecRepo {
	// 	return fmt.Errorf("cannot delete the special PPS repo %s", ppsconsts.SpecRepo)
	// }
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoRefCounts := d.repoRefCounts.ReadWriteInt(stm)
		commits := d.commits(repo.Name).ReadWrite(stm)
		branches := d.branches(repo.Name).ReadWrite(stm)

		// check if 'repo' is already gone. If so, return that error. Otherwise,
		// proceed with auth check (avoids awkward "access denied" error when calling
		// "deleteRepo" on a repo that's already gone)
		var existingRepoInfo pfs.RepoInfo
		err := repos.Get(repo.Name, &existingRepoInfo)
		if err != nil {
			if col.IsErrNotFound(err) {
				return fmt.Errorf("cannot delete \"%s\" as it does not exist", repo.Name)
			}
			return fmt.Errorf("error checking whether \"%s\" exists: %v",
				repo.Name, err)
		}

		// Check if the caller is authorized to delete this repo
		if err := d.checkIsAuthorized(ctx, repo, auth.Scope_OWNER); err != nil {
			return err
		}

		// Check if this repo is the provenance of some other repos
		if !force {
			refCount, err := repoRefCounts.Get(repo.Name)
			if err != nil {
				return err
			}
			if refCount != 0 {
				return fmt.Errorf("cannot delete the provenance of other repos")
			}
		}

		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(repo.Name, repoInfo); err != nil {
			return err
		}
		for _, prov := range repoInfo.Provenance {
			if err := repoRefCounts.Decrement(prov.Name); err != nil && !col.IsErrNotFound(err) {
				// Skip NotFound error, because it's possible that the
				// provenance repo has been deleted via --force.
				return err
			}
		}
		if err := repos.Delete(repo.Name); err != nil {
			return err
		}
		if err := repoRefCounts.Delete(repo.Name); err != nil {
			return err
		}
		commits.DeleteAll()
		branches.DeleteAll()
		return nil
	})
	if err != nil {
		return err
	}

	if _, err = d.pachClient.SetACL(auth.In2Out(ctx), &auth.SetACLRequest{
		Repo: repo.Name, // NewACL is unset, so this will clear the acl for 'repo'
	}); err != nil && !auth.IsNotActivatedError(err) {
		return grpcutil.ScrubGRPC(err)
	}
	return nil
}

func (d *driver) startCommit(ctx context.Context, parent *pfs.Commit, branch string, provenance []*pfs.Commit, description string) (*pfs.Commit, error) {
	return d.makeCommit(ctx, parent, branch, provenance, nil, description)
}

func (d *driver) buildCommit(ctx context.Context, parent *pfs.Commit, branch string, provenance []*pfs.Commit, tree *pfs.Object) (*pfs.Commit, error) {
	return d.makeCommit(ctx, parent, branch, provenance, tree, "")
}

func (d *driver) makeCommit(ctx context.Context, parent *pfs.Commit, branch string, provenance []*pfs.Commit, treeRef *pfs.Object, description string) (*pfs.Commit, error) {
	if err := d.checkIsAuthorized(ctx, parent.Repo, auth.Scope_WRITER); err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}
	commit := &pfs.Commit{
		Repo: parent.Repo,
		ID:   uuid.NewWithoutDashes(),
	}
	var tree hashtree.HashTree
	if treeRef != nil {
		var buf bytes.Buffer
		if err := d.pachClient.GetObject(treeRef.Hash, &buf); err != nil {
			return nil, err
		}
		_tree, err := hashtree.Deserialize(buf.Bytes())
		if err != nil {
			return nil, err
		}
		tree = _tree
	}
	if _, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		commits := d.commits(parent.Repo.Name).ReadWrite(stm)
		branches := d.branches(parent.Repo.Name).ReadWrite(stm)

		// Check if repo exists
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(parent.Repo.Name, repoInfo); err != nil {
			return err
		}

		commitInfo := &pfs.CommitInfo{
			Commit:      commit,
			Started:     now(),
			Description: description,
		}
		if branch != "" {
			// If we don't have an explicit parent we use the previous head of
			// branch as the parent, if it exists.
			branchInfo := new(pfs.BranchInfo)
			if err := branches.Get(branch, branchInfo); err != nil {
				if _, ok := err.(col.ErrNotFound); !ok {
					return err
				}
			} else if parent.ID == "" && branchInfo.Head != nil {
				parent.ID = branchInfo.Head.ID
			}
			branchInfo.Name = branch
			branchInfo.Head = commit
			branchInfo.Branch = client.NewBranch(commit.Repo.Name, branch)
			// Make commit the new head of the branch
			if err := branches.Put(branch, branchInfo); err != nil {
				return err
			}
		}
		if parent.ID != "" {
			parentCommitInfo, err := d.inspectCommit(ctx, parent, false)
			if err != nil {
				return err
			}
			// fail if the parent commit has not been finished
			if parentCommitInfo.Finished == nil {
				return fmt.Errorf("parent commit %s has not been finished", parent.ID)
			}
			commitInfo.ParentCommit = parent
		}

		// Use a map to de-dup provenance
		provenanceMap := make(map[string]*pfs.Commit)
		// Build the full provenance; my provenance's provenance is
		// my provenance
		for _, provCommit := range provenance {
			provCommits := d.commits(provCommit.Repo.Name).ReadWrite(stm)
			provCommitInfo := &pfs.CommitInfo{}
			if err := provCommits.Get(provCommit.ID, provCommitInfo); err != nil {
				return err
			}
			for _, c := range provCommitInfo.Provenance {
				provenanceMap[c.ID] = c
			}
		}
		// finally include the given provenance
		for _, provCommit := range provenance {
			provenanceMap[provCommit.ID] = provCommit
		}
		for _, provCommit := range provenanceMap {
			commitInfo.Provenance = append(commitInfo.Provenance, provCommit)
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(provCommit.Repo.Name).ReadWrite(stm).Update(provCommit.ID, provCommitInfo, func() error {
				appendSubvenance(provCommitInfo, commitInfo)
				return nil
			}); err != nil {
				return err
			}
		}
		parentTree, err := d.getTreeForCommit(ctx, parent)
		if err != nil {
			return err
		}
		if treeRef != nil {
			commitInfo.Tree = treeRef
			commitInfo.SizeBytes = uint64(tree.FSSize())
			commitInfo.Finished = now()
			repoInfo.SizeBytes += sizeChange(tree, parentTree)
			repos.Put(parent.Repo.Name, repoInfo)
		} else {
			d.openCommits.ReadWrite(stm).Put(commit.ID, commit)
		}
		if err := commits.Create(commit.ID, commitInfo); err != nil {
			return err
		}
		// We propagate the branch last so propagateCommit can write to the
		// now-existing commit's subvenance
		if branch != "" {
			return d.propagateCommit(ctx, client.NewBranch(commit.Repo.Name, branch), commitInfo, stm)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return commit, nil
}

func (d *driver) finishCommit(ctx context.Context, commit *pfs.Commit, tree *pfs.Object, empty bool, description string) (retErr error) {
	if err := d.checkIsAuthorized(ctx, commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	commitInfo, err := d.inspectCommit(ctx, commit, false)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return fmt.Errorf("commit %s has already been finished", commit.FullID())
	}
	if description != "" {
		commitInfo.Description = description
	}

	prefix, err := d.scratchCommitPrefix(ctx, commit)
	if err != nil {
		return err
	}
	defer func() {
		// Delete the scratch space for this commit
		_, err = d.etcdClient.Delete(ctx, prefix, etcd.WithPrefix())
		if err != nil && retErr == nil {
			retErr = err
		}
	}()

	var parentTree, finishedTree hashtree.HashTree
	if !empty {
		// Retrieve commit tree from parent commit (to apply writes from etcd or just
		// compute the size change). If parentCommit.Tree == nil, walk up the branch
		// until we find a successful commit. Otherwise, require that the immediate
		// parent of 'commitInfo' is closed, as we use its contents
		parentCommit := commitInfo.ParentCommit
		for parentCommit != nil {
			parentCommitInfo, err := d.inspectCommit(ctx, parentCommit, false)
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
			// Read everything under the scratch space for this commit
			resp, err := d.etcdClient.Get(ctx, prefix, etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortAscend))
			if err != nil {
				return err
			}

			tree := parentTree.Open()

			if err := d.applyWrites(resp, tree); err != nil {
				return err
			}

			finishedTree, err = tree.Finish()
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
				obj, _, err := d.pachClient.PutObject(bytes.NewReader(data))
				if err != nil {
					return err
				}

				commitInfo.Tree = obj
			}
		} else {
			var buf bytes.Buffer
			if err := d.pachClient.GetObject(tree.Hash, &buf); err != nil {
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
		// TODO this isn't part of the transaction because we can't do lists in transactions.
		// We need to maintain a list of branch refcounts.
		repos := d.repos.ReadWrite(stm)

		commits.Put(commit.ID, commitInfo)
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
			repos.Put(commit.Repo.Name, repoInfo)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return err
}

func (d *driver) propagateCommit(ctx context.Context, branch *pfs.Branch, commitInfo *pfs.CommitInfo, stm col.STM) error {
	commit := commitInfo.Commit
	var subvBranchInfos []*pfs.BranchInfo
	branchInfo := &pfs.BranchInfo{}
	if err := d.branches(branch.Repo.Name).ReadWrite(stm).Get(branch.Name, branchInfo); err != nil {
		return err
	}
	for _, subvBranch := range branchInfo.Subvenance {
		subvBranchInfo := &pfs.BranchInfo{}
		if err := d.branches(subvBranch.Repo.Name).ReadWrite(stm).Get(subvBranch.Name, subvBranchInfo); err != nil {
			return err
		}
		subvBranchInfos = append(subvBranchInfos, subvBranchInfo)
	}
	// Sort subvBranchInfos so that upstream branches are processed before their descendants
	sort.Slice(subvBranchInfos, func(i, j int) bool { return len(subvBranchInfos[i].Provenance) < len(subvBranchInfos[j].Provenance) })

	// C is provenant on B
	// B is a 'spec branch' in some formal sense
	// the commit that we would create in C would have only 'commit' in its provenance
	// don't create that commit, or any commits downstream of C.

	// Iterate through downstream branches, and create a new HEAD commit in all of
	// them.
	// 'branchToCommit' contains the commit we're using for each branch (subvBranchInfos
	// may have repeats)
	branchToCommit := make(map[string]*branchCommit)
	branchToCommit[branchKey(branch)] = &branchCommit{
		branch: branch,
		commit: commit,
	}
	for i, provCommit := range commitInfo.Provenance {
		provBranch := commitInfo.BranchProvenance[i]
		branchToCommit[branchKey(provBranch)] = &branchCommit{
			branch: provBranch,
			commit: provCommit,
		}
	}
	for _, subvBranchInfo := range subvBranchInfos {
		branch := subvBranchInfo.Branch
		repo := branch.Repo
		branches := d.branches(repo.Name).ReadWrite(stm)
		commits := d.commits(repo.Name).ReadWrite(stm)
		// new downstream commit
		commit := &pfs.Commit{
			Repo: repo,
			ID:   uuid.NewWithoutDashes(),
		}

		// Compute provenance of 'commit' (new downstream commit). One member
		// of its provenance will be the top-level commit that we're propagating,
		// but the head commit of other upstream branches will be in its provenance
		// too. Also look those up and add them.
		var provenance []*pfs.Commit
		var branchProvenance []*pfs.Branch
		for _, provBranch := range subvBranchInfo.Provenance {
			// TODO we store retrieved values in branchToCommit, but d.branches is
			// is already cached. Remove branchToCommit and just read from d.branches
			// in every iteration (simplifying this code)
			_, ok := branchToCommit[branchKey(provBranch)]
			provBranchInfo := &pfs.BranchInfo{}
			if !ok {
				if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Get(provBranch.Name, provBranchInfo); err != nil {
					if col.IsErrNotFound(err) {
						branchToCommit[branchKey(provBranch)] = nil
					} else {
						return err
					}
				}
				if provBranchInfo.Head != nil {
					branchToCommit[branchKey(provBranch)] = &branchCommit{
						commit: provBranchInfo.Head,
						branch: provBranch,
					}
				}
			}
			if branchToCommit[branchKey(provBranch)] != nil {
				provenance = append(provenance, branchToCommit[branchKey(provBranch)].commit)
				branchProvenance = append(branchProvenance, branchToCommit[branchKey(provBranch)].branch)
			}
		}
		// If the only branches in branchProvenance are in the 'spec' repo, this
		// output commit would create a confusing "dummy" job with no input
		// data--skip it
		allSpec := true
		for _, branch := range branchProvenance {
			if branch.Repo.Name != ppsconsts.SpecRepo {
				allSpec = false
				break
			}
		}
		if allSpec {
			continue
		}
		branchToCommit[branchKey(branch)] = &branchCommit{
			branch: branch, // downstream branch
			commit: commit, // new downstream commit
		}
		commitInfo := &pfs.CommitInfo{ // metadata for 'commit'
			Commit:           commit,
			Started:          now(),
			Provenance:       provenance,
			BranchProvenance: branchProvenance,
		}
		// Get 'commit's parent commit ('commit's branch head)
		var branchInfo pfs.BranchInfo
		if err := branches.Get(branch.Name, &branchInfo); err != nil {
			if _, ok := err.(col.ErrNotFound); !ok {
				return err
			}
		} else {
			commitInfo.ParentCommit = branchInfo.Head
		}
		// finally create the commit
		if err := commits.Create(commit.ID, commitInfo); err != nil {
			return err
		}
		if err := d.openCommits.ReadWrite(stm).Put(commit.ID, commit); err != nil {
			return err
		}
		branchInfo.Head = commit
		if err := branches.Put("master", &branchInfo); err != nil {
			return err
		}
		for _, provCommit := range commitInfo.Provenance {
			provCommitInfo := &pfs.CommitInfo{}
			if err := d.commits(provCommit.Repo.Name).ReadWrite(stm).Update(provCommit.ID, provCommitInfo, func() error {
				appendSubvenance(provCommitInfo, commitInfo)
				return nil
			}); err != nil {
				return err
			}
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
func (d *driver) inspectCommit(ctx context.Context, commit *pfs.Commit, block bool) (*pfs.CommitInfo, error) {
	if commit == nil {
		return nil, fmt.Errorf("cannot inspect nil commit")
	}
	if err := d.checkIsAuthorized(ctx, commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}

	commitID, ancestryLength := parseCommitID(commit.ID)

	// Check if the commitID is a branch name
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		branches := d.branches(commit.Repo.Name).ReadWrite(stm)

		branchInfo := new(pfs.BranchInfo)
		// See if we are given a branch
		if err := branches.Get(commitID, branchInfo); err != nil {
			if _, ok := err.(col.ErrNotFound); !ok {
				return err
			}
			// If it's not a branch, use it as it is
			return nil
		}
		if branchInfo.Head == nil {
			return pfsserver.ErrNoHead{branchInfo.Branch}
		}
		commitID = branchInfo.Head.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	var commitInfo *pfs.CommitInfo
	nextCommit := &pfs.Commit{
		Repo: commit.Repo,
		ID:   commitID,
	}
	for i := 0; i <= ancestryLength; i++ {
		if nextCommit == nil {
			return nil, pfsserver.ErrCommitNotFound{commit}
		}
		commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
		commitInfo = new(pfs.CommitInfo)
		if err := commits.Get(nextCommit.ID, commitInfo); err != nil {
			return nil, pfsserver.ErrCommitNotFound{nextCommit}
		}
		nextCommit = commitInfo.ParentCommit
	}

	commit.ID = commitInfo.Commit.ID
	if block {
		// Watch the CommitInfo until the commit has been finished
		commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
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
					return fmt.Errorf("commit %s deleted", commit.ID)
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
	if err := d.checkIsAuthorized(ctx, repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	if from != nil && from.Repo.Name != repo.Name || to != nil && to.Repo.Name != repo.Name {
		return nil, fmt.Errorf("`from` and `to` commits need to be from repo %s", repo.Name)
	}

	// Make sure that the repo exists
	_, err := d.inspectRepo(ctx, repo, !includeAuth)
	if err != nil {
		return nil, err
	}

	// Make sure that both from and to are valid commits
	if from != nil {
		_, err = d.inspectCommit(ctx, from, false)
		if err != nil {
			return nil, err
		}
	}
	if to != nil {
		_, err = d.inspectCommit(ctx, to, false)
		if err != nil {
			if _, ok := err.(pfsserver.ErrNoHead); ok {
				return nil, nil
			}
			return nil, err
		}
	}

	// if number is 0, we return all commits that match the criteria
	if number == 0 {
		number = math.MaxUint64
	}
	var commitInfos []*pfs.CommitInfo
	commits := d.commits(repo.Name).ReadOnly(ctx)

	if from != nil && to == nil {
		return nil, fmt.Errorf("cannot use `from` commit without `to` commit")
	} else if from == nil && to == nil {
		// if neither from and to is given, we list all commits in
		// the repo, sorted by revision timestamp
		iterator, err := commits.List()
		if err != nil {
			return nil, err
		}
		var commitID string
		for number != 0 {
			var commitInfo pfs.CommitInfo
			ok, err := iterator.Next(&commitID, &commitInfo)
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			commitInfos = append(commitInfos, &commitInfo)
			number--
		}
	} else {
		cursor := to
		for number != 0 && cursor != nil && (from == nil || cursor.ID != from.ID) {
			var commitInfo pfs.CommitInfo
			if err := commits.Get(cursor.ID, &commitInfo); err != nil {
				return nil, err
			}
			commitInfos = append(commitInfos, &commitInfo)
			cursor = commitInfo.ParentCommit
			number--
		}
	}
	return commitInfos, nil
}

type commitStream struct {
	stream chan CommitEvent
	done   chan struct{}
}

func (c *commitStream) Stream() <-chan CommitEvent {
	return c.stream
}

func (c *commitStream) Close() {
	close(c.done)
}

func (d *driver) subscribeCommit(ctx context.Context, repo *pfs.Repo, branch string, from *pfs.Commit) (CommitStream, error) {
	d.initializePachConn()
	if from != nil && from.Repo.Name != repo.Name {
		return nil, fmt.Errorf("the `from` commit needs to be from repo %s", repo.Name)
	}

	// We need to watch for new commits before we start listing commits,
	// because otherwise we might miss some commits in between when we
	// finish listing and when we start watching.
	branches := d.branches(repo.Name).ReadOnly(ctx)
	newCommitWatcher, err := branches.WatchOne(branch)
	if err != nil {
		return nil, err
	}

	stream := make(chan CommitEvent)
	done := make(chan struct{})

	go func() (retErr error) {
		defer newCommitWatcher.Close()
		defer func() {
			if retErr != nil {
				select {
				case stream <- CommitEvent{
					Err: retErr,
				}:
				case <-done:
				}
			}
			close(stream)
		}()
		// keep track of the commits that have been sent
		seen := make(map[string]bool)
		// include all commits that are currently on the given branch,
		// but only the ones that have been finished
		commitInfos, err := d.listCommit(ctx, repo, &pfs.Commit{
			Repo: repo,
			ID:   branch,
		}, from, 0)
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
			select {
			case stream <- CommitEvent{
				Value: commitInfo,
			}:
				seen[commitInfo.Commit.ID] = true
			case <-done:
				return nil
			}
		}

		for {
			var branchName string
			branchInfo := &pfs.BranchInfo{}
			var event *watch.Event
			var ok bool
			select {
			case event, ok = <-newCommitWatcher.Watch():
			case <-done:
				return nil
			}
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
			case watch.EventDelete:
				continue
			}
			if branchInfo.Head == nil {
				continue // put event == new branch was created. No commits yet though
			}

			// We don't want to include the `from` commit itself

			// TODO we're check the branchName because right now WatchOne,
			// like all collection watching commands returns prefixes which
			// means we'll get back `master-v1` if we're looking for
			// `master` once this is changed we should remove the
			// comparison between branchName and branch.
			if branchName == branch && (!(seen[branchInfo.Head.ID] || (from != nil && from.ID == branchInfo.Head.ID))) {
				commitInfo, err := d.inspectCommit(ctx, branchInfo.Head, false)
				if err != nil {
					return err
				}
				select {
				case stream <- CommitEvent{
					Value: commitInfo,
				}:
					seen[commitInfo.Commit.ID] = true
				case <-done:
					return nil
				}
			}
		}
	}()

	return &commitStream{
		stream: stream,
		done:   done,
	}, nil
}

func (d *driver) flushCommit(ctx context.Context, fromCommits []*pfs.Commit, toRepos []*pfs.Repo, f func(commitInfo *pfs.CommitInfo) error) error {
	if len(fromCommits) == 0 {
		return fmt.Errorf("fromCommits cannot be empty")
	}
	d.initializePachConn()

	// First compute intersection of the fromCommits subvenant commits, those
	// are the commits we're interested in. Iterate over all commits and keep a
	// running intersection (in commitsToWatch) of the subvenance of all commits
	// processed so far
	commitsToWatch := make(map[string]*pfs.Commit)
	for i, commit := range fromCommits {
		commitInfo, err := d.inspectCommit(ctx, commit, false)
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
		finishedCommitInfo, err := d.inspectCommit(ctx, commitToWatch, true)
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

func (d *driver) flushRepo(ctx context.Context, repo *pfs.Repo) ([]*pfs.RepoInfo, error) {
	iter, err := d.repos.ReadOnly(ctx).GetByIndex(pfsdb.ProvenanceIndex, repo)
	if err != nil {
		return nil, err
	}
	var repoInfos []*pfs.RepoInfo
	for {
		var repoName string
		repoInfo := new(pfs.RepoInfo)
		ok, err := iter.Next(&repoName, repoInfo)
		if !ok {
			return repoInfos, nil
		}
		if err != nil {
			return nil, err
		}
		repoInfos = append(repoInfos, repoInfo)
	}
}

func (d *driver) deleteCommit(ctx context.Context, commit *pfs.Commit) error {
	if err := d.checkIsAuthorized(ctx, commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	commitInfo, err := d.inspectCommit(ctx, commit, false)
	if err != nil {
		return err
	}

	if commitInfo.Finished != nil {
		return fmt.Errorf("cannot delete finished commit")
	}

	// Delete the scratch space for this commit
	prefix, err := d.scratchCommitPrefix(ctx, commit)
	if err != nil {
		return err
	}
	_, err = d.etcdClient.Delete(ctx, prefix, etcd.WithPrefix())
	if err != nil {
		return err
	}

	// If this commit is the head of a branch, make the commit's parent
	// the head instead.
	branchInfos, err := d.listBranch(ctx, commit.Repo)
	if err != nil {
		return err
	}

	for _, branchInfo := range branchInfos {
		if branchInfo.Head != nil && branchInfo.Head.ID == commitInfo.Commit.ID {
			var provenance []*pfs.Branch
			for _, provBranch := range branchInfo.DirectProvenance {
				provenance = append(provenance, provBranch)
			}
			if err := d.createBranch(ctx, branchInfo.Branch, commitInfo.ParentCommit, provenance); err != nil {
				return err
			}
		}
	}

	// Delete the commit itself and subtract the size of the commit
	// from repo size.
	_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
			return err
		}
		repoInfo.SizeBytes -= commitInfo.SizeBytes
		repos.Put(commit.Repo.Name, repoInfo)

		commits := d.commits(commit.Repo.Name).ReadWrite(stm)
		return commits.Delete(commit.ID)
	})

	return err
}

func (d *driver) createBranch(ctx context.Context, branch *pfs.Branch, commit *pfs.Commit, provenance []*pfs.Branch) error {
	if err := d.checkIsAuthorized(ctx, branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		// if 'commit' is a branch, resolve it
		var commitInfo *pfs.CommitInfo
		var err error
		if commit != nil {
			commitInfo, err = d.inspectCommit(ctx, commit, false)
			if err != nil {
				// possible that branch exists but has no head commit. This is fine, but
				// branchInfo.Head must also be nil
				if _, ok := err.(pfsserver.ErrNoHead); !ok {
					return fmt.Errorf("inspectCommit: %v", err)
				}
				commit = nil
			} else {
				commit = commitInfo.Commit
			}
		}
		branchInfo := &pfs.BranchInfo{
			Branch: branch,
			Name:   branch.Name,
		}
		// Get the previous version of this branch if there is one, we do this
		// so that we don't lose pre existing subvenance.
		if err := d.branches(branch.Repo.Name).ReadWrite(stm).Get(branch.Name, branchInfo); err != nil && !col.IsErrNotFound(err) {
			return err
		}
		branchInfo.Head = commit
		branchInfo.DirectProvenance = nil
		for _, provBranch := range provenance {
			(*branchSet)(&branchInfo.DirectProvenance).add(provBranch)
		}
		// We don't just need to update (or create) branch's Provenance but
		// also the Provenance of all our Subvenance (in the case of an update)
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
			for _, oldProvBranch := range oldProvenance {
				if !(*branchSet)(&branchInfo.Provenance).has(oldProvBranch) {
					// Provenance was deleted, so we delete ourselves from their subvenance
					oldProvBranchInfo := &pfs.BranchInfo{}
					if err := d.branches(oldProvBranch.Repo.Name).ReadWrite(stm).Upsert(oldProvBranch.Name, oldProvBranchInfo, func() error {
						(*branchSet)(&oldProvBranchInfo.Subvenance).del(branchInfo.Branch)
						return nil
					}); err != nil {
						return err
					}
				}
			}
		}
		if len(branchInfo.Provenance) > 0 {
			commitProvMap := make(map[string]*branchCommit)
			// Update branches to have this new branch as subvenance, and possibly
			// create a new HEAD commit in this new branch
			for _, provBranch := range branchInfo.Provenance {
				provBranchInfo := &pfs.BranchInfo{}
				// record the fact that we are subvenance for all of our provenant branches
				if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Get(provBranch.Name, provBranchInfo); err != nil {
					return err
				}
				// provBranchInfo contains valid data re:provBranch following Upsert.
				// Check if there are any commits in provBranch (might need to create a
				// HEAD commit in this branch)
				if provBranchInfo.Head != nil {
					provCommitInfo := &pfs.CommitInfo{}
					if err := d.commits(provBranchInfo.Head.Repo.Name).ReadWrite(stm).Get(provBranchInfo.Head.ID, provCommitInfo); err != nil {
						return err
					}
					commitProvMap[commitKey(provBranchInfo.Head)] = &branchCommit{
						commit: provBranchInfo.Head,
						branch: provBranchInfo.Branch,
					}
					for i, provCommit := range provCommitInfo.Provenance {
						commitProvMap[commitKey(provCommit)] = &branchCommit{
							commit: provCommit,
							branch: provCommitInfo.BranchProvenance[i],
						}
					}
				}
			}
			if len(commitProvMap) > 0 {
				// we're updating the branch, but new HEAD commit has exact same
				// provenance as existing head commit--skip it
				if commitInfo != nil {
					headIsSubset := true
					for _, c := range commitInfo.Provenance {
						if _, ok := commitProvMap[commitKey(c)]; !ok {
							headIsSubset = false
							break
						}
					}
					if len(commitInfo.Provenance) == len(commitProvMap) && headIsSubset {
						return nil // existing head commit is the same as the one we want to create--nothing new to do
					}
				}
				// If the only branches in the output commit's provenance are in the
				// 'spec' repo, this output commit would create a confusing "dummy" job
				// with no input data--skip it
				allSpec := true
				for _, b := range commitProvMap {
					if b.branch.Repo.Name != ppsconsts.SpecRepo {
						allSpec = false
						break
					}
				}
				if !allSpec {
					commit := &pfs.Commit{
						Repo: branch.Repo,
						ID:   uuid.NewWithoutDashes(),
					}
					commitInfo := &pfs.CommitInfo{
						Commit:  commit,
						Started: now(),
					}
					branchInfo.Head = commitInfo.Commit
					for _, provCommit := range commitProvMap {
						commitInfo.Provenance = append(commitInfo.Provenance, provCommit.commit)
						commitInfo.BranchProvenance = append(commitInfo.BranchProvenance, provCommit.branch)
						provCommitInfo := &pfs.CommitInfo{}
						if err := d.commits(provCommit.commit.Repo.Name).ReadWrite(stm).Upsert(provCommit.commit.ID, provCommitInfo, func() error {
							appendSubvenance(provCommitInfo, commitInfo)
							return nil
						}); err != nil {
							return err
						}
					}
					if err := d.commits(commit.Repo.Name).ReadWrite(stm).Create(commit.ID, commitInfo); err != nil {
						return err
					}
					if err := d.openCommits.ReadWrite(stm).Put(commit.ID, commit); err != nil {
						return err
					}
					if err := d.propagateCommit(ctx, branch, commitInfo, stm); err != nil {
						return err
					}
				}
			}
		}
		branches := d.branches(branch.Repo.Name).ReadWrite(stm)
		return branches.Put(branch.Name, branchInfo)
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
	if err := d.checkIsAuthorized(ctx, repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	branches := d.branches(repo.Name).ReadOnly(ctx)
	iterator, err := branches.List()
	if err != nil {
		return nil, err
	}

	var res []*pfs.BranchInfo
	for {
		var branchName string
		branchInfo := new(pfs.BranchInfo)
		ok, err := iterator.Next(&branchName, branchInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		res = append(res, branchInfo)
	}
	return res, nil
}

func (d *driver) deleteBranch(ctx context.Context, branch *pfs.Branch) error {
	if err := d.checkIsAuthorized(ctx, branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		branches := d.branches(branch.Repo.Name).ReadWrite(stm)
		return branches.Delete(branch.Name)
	})
	return err
}

func (d *driver) scratchPrefix() string {
	return path.Join(d.prefix, "scratch")
}

// scratchCommitPrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchCommitPrefix(ctx context.Context, commit *pfs.Commit) (string, error) {
	if _, err := d.inspectCommit(ctx, commit, false); err != nil {
		return "", err
	}
	return path.Join(d.scratchPrefix(), commit.Repo.Name, commit.ID), nil
}

// scratchFilePrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchFilePrefix(ctx context.Context, file *pfs.File) (string, error) {
	return path.Join(d.scratchPrefix(), file.Commit.Repo.Name, file.Commit.ID, file.Path), nil
}

func (d *driver) filePathFromEtcdPath(etcdPath string) string {
	trimmed := strings.TrimPrefix(etcdPath, d.scratchPrefix())
	// trimmed looks like /repo/commit/path/to/file
	split := strings.Split(trimmed, "/")
	// we only want /path/to/file so we use index 3 (note that there's an "" at
	// the beginning of the slice because of the lead /)
	return path.Join(split[3:]...)
}

// checkPath checks if a file path is legal
func checkPath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("filename cannot contain null character: %s", path)
	}
	return nil
}

func (d *driver) putFile(ctx context.Context, file *pfs.File, delimiter pfs.Delimiter,
	targetFileDatums int64, targetFileBytes int64, overwriteIndex *pfs.OverwriteIndex, reader io.Reader) error {
	if err := d.checkIsAuthorized(ctx, file.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	// Check if the commit ID is a branch name.  If so, we have to
	// get the real commit ID in order to check if the commit does exist
	// and is open.
	// Since we use UUIDv4 for commit IDs, the 13th character would be 4 if
	// this is a commit ID.
	if len(file.Commit.ID) != uuid.UUIDWithoutDashesLength || file.Commit.ID[12] != '4' {
		commitInfo, err := d.inspectCommit(ctx, file.Commit, false)
		if err != nil {
			return err
		}
		file.Commit = commitInfo.Commit
	}

	if overwriteIndex != nil && overwriteIndex.Index == 0 {
		if err := d.deleteFile(ctx, file); err != nil {
			return err
		}
	}

	records := &pfs.PutFileRecords{}
	if err := checkPath(file.Path); err != nil {
		return err
	}
	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	// Put the tree into the blob store
	// Only write the records to etcd if the commit does exist and is open.
	// To check that a key exists in etcd, we assert that its CreateRevision
	// is greater than zero.
	putRecords := func() error {
		marshalledRecords, err := records.Marshal()
		if err != nil {
			return err
		}
		kvc := etcd.NewKV(d.etcdClient)

		txnResp, err := kvc.Txn(ctx).
			If(etcd.Compare(etcd.CreateRevision(d.openCommits.Path(file.Commit.ID)), ">", 0)).Then(etcd.OpPut(path.Join(prefix, uuid.NewWithoutDashes()), string(marshalledRecords))).Commit()
		if err != nil {
			return err
		}
		if !txnResp.Succeeded {
			return fmt.Errorf("commit %v is not open", file.Commit.ID)
		}
		return nil
	}

	if delimiter == pfs.Delimiter_NONE {
		objects, size, err := d.pachClient.PutObjectSplit(reader)
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

		return putRecords()
	}
	buffer := &bytes.Buffer{}
	var datumsWritten int64
	var bytesWritten int64
	var filesPut int
	EOF := false
	var eg errgroup.Group
	decoder := json.NewDecoder(reader)
	bufioR := bufio.NewReader(reader)

	indexToRecord := make(map[int]*pfs.PutFileRecord)
	var mu sync.Mutex
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
			eg.Go(func() error {
				object, size, err := d.pachClient.PutObject(_buffer)
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

	return putRecords()
}

func (d *driver) copyFile(ctx context.Context, src *pfs.File, dst *pfs.File, overwrite bool) error {
	if err := d.checkIsAuthorized(ctx, src.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	if err := d.checkIsAuthorized(ctx, dst.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	if err := checkPath(dst.Path); err != nil {
		return err
	}
	// Check if the commit ID is a branch name.  If so, we have to
	// get the real commit ID in order to check if the commit does exist
	// and is open.
	// Since we use UUIDv4 for commit IDs, the 13th character would be 4 if
	// this is a commit ID.
	if len(dst.Commit.ID) != uuid.UUIDWithoutDashesLength || dst.Commit.ID[12] != '4' {
		commitInfo, err := d.inspectCommit(ctx, dst.Commit, false)
		if err != nil {
			return err
		}
		dst.Commit = commitInfo.Commit
	}
	if overwrite {
		if err := d.deleteFile(ctx, dst); err != nil {
			return err
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
			records := &pfs.PutFileRecords{}
			file := client.NewFile(dst.Commit.Repo.Name, dst.Commit.ID, path.Clean(path.Join(dst.Path, relPath)))
			prefix, err := d.scratchFilePrefix(ctx, file)
			if err != nil {
				return err
			}
			for i, object := range node.FileNode.Objects {
				var size int64
				if i == 0 {
					size = node.SubtreeSize
				}
				records.Records = append(records.Records, &pfs.PutFileRecord{
					SizeBytes:  size,
					ObjectHash: object.Hash,
				})
			}
			marshalledRecords, err := records.Marshal()
			if err != nil {
				return err
			}
			kvc := etcd.NewKV(d.etcdClient)
			txnResp, err := kvc.Txn(ctx).
				If(etcd.Compare(etcd.CreateRevision(d.openCommits.Path(file.Commit.ID)), ">", 0)).Then(etcd.OpPut(path.Join(prefix, uuid.NewWithoutDashes()), string(marshalledRecords))).Commit()
			if err != nil {
				return err
			}
			if !txnResp.Succeeded {
				return fmt.Errorf("commit %v is not open", file.Commit.ID)
			}
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}

func (d *driver) getTreeForCommit(ctx context.Context, commit *pfs.Commit) (hashtree.HashTree, error) {
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

	if _, err := d.inspectCommit(ctx, commit, false); err != nil {
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
	if err := d.pachClient.GetObject(treeRef.Hash, &buf); err != nil {
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
	commitInfo, err := d.inspectCommit(ctx, file.Commit, false)
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
	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return nil, err
	}
	// Read everything under the scratch space for this commit
	resp, err := d.etcdClient.Get(ctx, prefix, etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortAscend))
	if err != nil {
		return nil, err
	}

	parentTree, err := d.getTreeForCommit(ctx, commitInfo.ParentCommit)
	if err != nil {
		return nil, err
	}
	openTree := parentTree.Open()
	if err := d.applyWrites(resp, openTree); err != nil {
		return nil, err
	}
	tree, err := openTree.Finish()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (d *driver) getFile(ctx context.Context, file *pfs.File, offset int64, size int64) (io.Reader, error) {
	if err := d.checkIsAuthorized(ctx, file.Commit.Repo, auth.Scope_READER); err != nil {
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

	if node.FileNode == nil {
		return nil, fmt.Errorf("%s is a directory", file.Path)
	}

	getObjectsClient, err := d.pachClient.ObjectAPIClient.GetObjects(
		ctx,
		&pfs.GetObjectsRequest{
			Objects:     node.FileNode.Objects,
			OffsetBytes: uint64(offset),
			SizeBytes:   uint64(size),
			TotalSize:   uint64(node.SubtreeSize),
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
	if err := d.checkIsAuthorized(ctx, file.Commit.Repo, auth.Scope_READER); err != nil {
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

	return nodeToFileInfo(file.Commit, file.Path, node, true), nil
}

func (d *driver) listFile(ctx context.Context, file *pfs.File, full bool) ([]*pfs.FileInfo, error) {
	if err := d.checkIsAuthorized(ctx, file.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	tree, err := d.getTreeForFile(ctx, file)
	if err != nil {
		return nil, err
	}

	nodes, err := tree.List(file.Path)
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for _, node := range nodes {
		fileInfos = append(fileInfos, nodeToFileInfo(file.Commit, path.Join(file.Path, node.Name), node, full))
	}
	return fileInfos, nil
}

func (d *driver) globFile(ctx context.Context, commit *pfs.Commit, pattern string) ([]*pfs.FileInfo, error) {
	if err := d.checkIsAuthorized(ctx, commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	tree, err := d.getTreeForFile(ctx, client.NewFile(commit.Repo.Name, commit.ID, ""))
	if err != nil {
		return nil, err
	}

	nodes, err := tree.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for _, node := range nodes {
		fileInfos = append(fileInfos, nodeToFileInfo(commit, node.Name, node, false))
	}
	return fileInfos, nil
}

func (d *driver) diffFile(ctx context.Context, newFile *pfs.File, oldFile *pfs.File, shallow bool) ([]*pfs.FileInfo, []*pfs.FileInfo, error) {
	// Do READER authorization check for both newFile and oldFile
	if oldFile != nil && oldFile.Commit != nil {
		//	if oldFile != nil {
		if err := d.checkIsAuthorized(ctx, oldFile.Commit.Repo, auth.Scope_READER); err != nil {
			return nil, nil, err
		}
	}
	if newFile != nil && newFile.Commit != nil {
		//	if newFile != nil {
		if err := d.checkIsAuthorized(ctx, newFile.Commit.Repo, auth.Scope_READER); err != nil {
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
		newCommitInfo, err := d.inspectCommit(ctx, newFile.Commit, false)
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
	if err := d.checkIsAuthorized(ctx, file.Commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	commitInfo, err := d.inspectCommit(ctx, file.Commit, false)
	if err != nil {
		return err
	}

	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{file.Commit}
	}

	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	_, err = d.etcdClient.Put(ctx, path.Join(prefix, uuid.NewWithoutDashes()), tombstone)
	return err
}

func (d *driver) deleteAll(ctx context.Context) error {
	repoInfos, err := d.listRepo(ctx, nil, !includeAuth)
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		// Hack to keep 'DeleteAll()' from breaking a cluster by deleting the
		// pipeline spec repo. It would be nice if this didn't need to exist.
		if repoInfo.Repo.Name == ppsconsts.SpecRepo {
			continue
		}
		if err := d.deleteRepo(ctx, repoInfo.Repo, true); err != nil && !auth.IsNotAuthorizedError(err) {
			return err
		}
	}

	// Second hack: Normally PPS would delete all commits in the spec repo, but
	// DeleteCommit() can't yet delete commits that have been finished. Instead,
	// PFS Deletes the spec repo at the end of DeleteAll() and recreates it
	// (which PPS can't do, as it doesn't know when PFS.DeleteAll() will run)
	// TODO(msteffen): Delete this, and either hide PFS.DeleteAll() from users and
	// put this block in PPS.DeleteAll, or implement PFS.DeleteCommit() for
	// finished commits and have PPS.DeleteAll() delete all commits in the spec
	// repo.
	if err := d.deleteRepo(ctx, client.NewRepo(ppsconsts.SpecRepo), true); err != nil {
		return err
	}
	return d.createRepo(ctx, client.NewRepo(ppsconsts.SpecRepo), nil, "", false)
}

func (d *driver) applyWrites(resp *etcd.GetResponse, tree hashtree.OpenHashTree) error {
	// a map that keeps track of the sizes of objects
	sizeMap := make(map[string]int64)
	for _, kv := range resp.Kvs {
		// fileStr is going to look like "some/path/UUID"
		fileStr := d.filePathFromEtcdPath(string(kv.Key))
		// the last element of `parts` is going to be UUID
		parts := strings.Split(fileStr, "/")
		// filePath should look like "some/path"
		filePath := strings.Join(parts[:len(parts)-1], "/")

		if string(kv.Value) == tombstone {
			if err := tree.DeleteFile(filePath); err != nil {
				// Deleting a non-existent file in an open commit should
				// be a no-op
				if hashtree.Code(err) != hashtree.PathNotFound {
					return err
				}
			}
		} else {
			records := &pfs.PutFileRecords{}
			if err := records.Unmarshal(kv.Value); err != nil {
				return err
			}
			if !records.Split {
				if len(records.Records) == 0 {
					return fmt.Errorf("unexpect %d length pfs.PutFileRecord (this is likely a bug)", len(records.Records))
				}
				for _, record := range records.Records {
					sizeMap[record.ObjectHash] = record.SizeBytes
					if record.OverwriteIndex != nil {
						// Computing size delta
						delta := record.SizeBytes
						fileNode, err := tree.Get(filePath)
						if err == nil {
							// If we can't find the file, that's fine.
							for i := record.OverwriteIndex.Index; int(i) < len(fileNode.FileNode.Objects); i++ {
								delta -= sizeMap[fileNode.FileNode.Objects[i].Hash]
							}
						}

						if err := tree.PutFileOverwrite(filePath, []*pfs.Object{{Hash: record.ObjectHash}}, record.OverwriteIndex, delta); err != nil {
							return err
						}
					} else {
						if err := tree.PutFile(filePath, []*pfs.Object{{Hash: record.ObjectHash}}, record.SizeBytes); err != nil {
							return err
						}
					}
				}
			} else {
				nodes, err := tree.List(filePath)
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
				for i, record := range records.Records {
					if err := tree.PutFile(path.Join(filePath, fmt.Sprintf(splitSuffixFmt, i+int(indexOffset))), []*pfs.Object{{Hash: record.ObjectHash}}, record.SizeBytes); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func commitKey(commit *pfs.Commit) string {
	return fmt.Sprintf("%s/%s", commit.Repo.Name, commit.ID)
}

func branchKey(branch *pfs.Branch) string {
	return fmt.Sprintf("%s/%s", branch.Repo.Name, branch.Name)
}

func (d *driver) addBranchProvenance(branchInfo *pfs.BranchInfo, provBranch *pfs.Branch, stm col.STM) error {
	(*branchSet)(&branchInfo.Provenance).add(provBranch)
	provBranchInfo := &pfs.BranchInfo{}
	return d.branches(provBranch.Repo.Name).ReadWrite(stm).Upsert(provBranch.Name, provBranchInfo, func() error {
		// Set provBranch, we may be creating this branch for the first time
		provBranchInfo.Branch = provBranch
		(*branchSet)(&provBranchInfo.Subvenance).add(branchInfo.Branch)
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

func (b *branchSet) add(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if !ok {
		*b = append(*b, nil)
		copy((*b)[i+1:], (*b)[i:])
		(*b)[i] = branch
	}
}

func (b *branchSet) del(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if ok {
		copy((*b)[i:], (*b)[i+1:])
		(*b)[len((*b))-1] = nil
		*b = (*b)[:len((*b))-1]
	}
}

func (b *branchSet) has(branch *pfs.Branch) bool {
	_, ok := b.search(branch)
	return ok
}
