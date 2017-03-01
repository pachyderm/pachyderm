package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"google.golang.org/grpc"
)

type collectionFactory func(string) col.Collection

type driver struct {
	address      string
	pachConnOnce sync.Once
	pachConn     *grpc.ClientConn
	etcdClient   *etcd.Client
	prefix       string

	// collections
	repos         col.Collection
	repoRefCounts col.Collection
	commits       collectionFactory
	branches      collectionFactory
}

const (
	TOMBSTONE = "delete"
)

// collection prefixes
const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	branchesPrefix      = "/branches"
)

var (
	provenanceIndex = col.Index{
		Field: "Provenance",
		Multi: true,
	}
)

// NewDriver is used to create a new Driver instance
func NewDriver(address string, etcdAddresses []string, etcdPrefix string) (Driver, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddresses,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &driver{
		address:    address,
		etcdClient: etcdClient,
		prefix:     etcdPrefix,
		repos: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, reposPrefix),
			[]col.Index{provenanceIndex},
			&pfs.RepoInfo{},
		),
		repoRefCounts: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, repoRefCountsPrefix),
			nil,
			nil,
		),
		commits: func(repo string) col.Collection {
			return col.NewCollection(
				etcdClient,
				path.Join(etcdPrefix, commitsPrefix, repo),
				[]col.Index{provenanceIndex},
				&pfs.CommitInfo{},
			)
		},
		branches: func(repo string) col.Collection {
			return col.NewCollection(
				etcdClient,
				path.Join(etcdPrefix, branchesPrefix, repo),
				nil,
				&pfs.Commit{},
			)
		},
	}, nil
}

// NewLocalDriver creates a driver using an local etcd instance.  This
// function is intended for testing purposes
func NewLocalDriver(blockAddress string, etcdPrefix string) (Driver, error) {
	return NewDriver(blockAddress, []string{"localhost:2379"}, etcdPrefix)
}

func (d *driver) getBlockClient() (*client.APIClient, error) {
	if d.pachConn == nil {
		var onceErr error
		d.pachConnOnce.Do(func() {
			pachConn, err := grpc.Dial(d.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			d.pachConn = pachConn
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return &client.APIClient{BlockAPIClient: pfs.NewBlockAPIClient(d.pachConn)}, nil
}

func (d *driver) getObjectClient() (*client.APIClient, error) {
	if d.pachConn == nil {
		var onceErr error
		d.pachConnOnce.Do(func() {
			pachConn, err := grpc.Dial(d.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			d.pachConn = pachConn
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return &client.APIClient{ObjectAPIClient: pfs.NewObjectAPIClient(d.pachConn)}, nil
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

func (d *driver) CreateRepo(ctx context.Context, repo *pfs.Repo, provenance []*pfs.Repo) error {
	if err := ValidateRepoName(repo.Name); err != nil {
		return err
	}

	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoRefCounts := d.repoRefCounts.ReadWriteInt(stm)

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
			Repo:       repo,
			Created:    now(),
			Provenance: fullProvRepos,
		}
		return repos.Create(repo.Name, repoInfo)
	})
	return err
}

func (d *driver) InspectRepo(ctx context.Context, repo *pfs.Repo) (*pfs.RepoInfo, error) {
	repoInfo := new(pfs.RepoInfo)
	if err := d.repos.ReadOnly(ctx).Get(repo.Name, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	var result []*pfs.RepoInfo
	repos := d.repos.ReadOnly(ctx)
	// Ensure that all provenance repos exist
	for _, prov := range provenance {
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(prov.Name, repoInfo); err != nil {
			return nil, err
		}
	}

	iterator, err := repos.List()
	if err != nil {
		return nil, err
	}
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
		result = append(result, repoInfo)
	}
	return result, nil
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		repos := d.repos.ReadWrite(stm)
		repoRefCounts := d.repoRefCounts.ReadWriteInt(stm)
		commits := d.commits(repo.Name).ReadWrite(stm)
		branches := d.branches(repo.Name).ReadWrite(stm)

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
			if err := repoRefCounts.Decrement(prov.Name); err != nil {
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
	return err
}

func (d *driver) StartCommit(ctx context.Context, parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error) {
	return d.makeCommit(ctx, parent, provenance, nil)
}

func (d *driver) BuildCommit(ctx context.Context, parent *pfs.Commit, provenance []*pfs.Commit, tree *pfs.Object) (*pfs.Commit, error) {
	return d.makeCommit(ctx, parent, provenance, tree)
}

func (d *driver) makeCommit(ctx context.Context, parent *pfs.Commit, provenance []*pfs.Commit, treeRef *pfs.Object) (*pfs.Commit, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}
	commit := &pfs.Commit{
		Repo: parent.Repo,
		ID:   uuid.NewWithoutDashes(),
	}
	var commitSize uint64
	if treeRef != nil {
		objClient, err := d.getObjectClient()
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := objClient.GetObject(treeRef.Hash, &buf); err != nil {
			return nil, err
		}
		tree, err := hashtree.Deserialize(buf.Bytes())
		if err != nil {
			return nil, err
		}
		commitSize = uint64(tree.Size())
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
			Commit:  commit,
			Started: now(),
		}

		// Build the full provenance; my provenance's provenance is
		// my provenance
		for _, prov := range provenance {
			provCommits := d.commits(prov.Repo.Name).ReadWrite(stm)
			provCommitInfo := new(pfs.CommitInfo)
			if err := provCommits.Get(prov.ID, provCommitInfo); err != nil {
				return err
			}
			commitInfo.Provenance = append(commitInfo.Provenance, provCommitInfo.Provenance...)
		}
		// finally include the given provenance
		commitInfo.Provenance = append(commitInfo.Provenance, provenance...)

		if parent.ID != "" {
			head := new(pfs.Commit)
			// See if we are given a branch
			if err := branches.Get(parent.ID, head); err != nil {
				if _, ok := err.(col.ErrNotFound); !ok {
					return err
				}
			} else {
				// if parent.ID is a branch, make myself the head
				branches.Put(parent.ID, commit)
				parent.ID = head.ID
			}
			// Check that the parent commit exists
			parentCommitInfo := new(pfs.CommitInfo)
			if err := commits.Get(parent.ID, parentCommitInfo); err != nil {
				return err
			}
			// fail if the parent commit has not been finished
			if parentCommitInfo.Finished == nil {
				return fmt.Errorf("parent commit %s has not been finished", parent.ID)
			}
			commitInfo.ParentCommit = parent
		}
		if treeRef != nil {
			commitInfo.Tree = treeRef
			commitInfo.SizeBytes = commitSize
			commitInfo.Finished = now()
			repoInfo.SizeBytes += commitSize
			repos.Put(parent.Repo.Name, repoInfo)
		}
		return commits.Create(commit.ID, commitInfo)
	}); err != nil {
		return nil, err
	}

	return commit, nil
}

func (d *driver) FinishCommit(ctx context.Context, commit *pfs.Commit) error {
	if err := d.resolveBranch(ctx, commit); err != nil {
		return err
	}

	prefix, err := d.scratchCommitPrefix(ctx, commit)
	if err != nil {
		return err
	}

	// Read everything under the scratch space for this commit
	// TODO: lock the scratch space to prevent concurrent PutFile
	resp, err := d.etcdClient.Get(ctx, prefix, etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortAscend))
	if err != nil {
		return err
	}

	commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
	commitInfo := new(pfs.CommitInfo)
	if err := commits.Get(commit.ID, commitInfo); err != nil {
		return err
	}

	if commitInfo.Finished != nil {
		return fmt.Errorf("commit %s has already been finished", commit.FullID())
	}

	_tree, err := d.getTreeForCommit(ctx, commitInfo.ParentCommit)
	if err != nil {
		return err
	}
	tree := _tree.Open()

	for _, kv := range resp.Kvs {
		// fileStr is going to look like "some/path/0"
		fileStr := strings.TrimPrefix(string(kv.Key), prefix)
		// the last element of `parts` is going to be 0
		parts := strings.Split(fileStr, "/")
		// filePath should look like "some/path"
		filePath := strings.Join(parts[:len(parts)-1], "/")

		if string(kv.Value) == TOMBSTONE {
			if err := tree.DeleteFile(filePath); err != nil {
				return err
			}
		} else {
			// The serialized data contains multiple blockrefs; read them all
			var blockRefs []*pfs.BlockRef
			data := bytes.NewReader(kv.Value)
			for {
				blockRef := &pfs.BlockRef{}
				if _, err := pbutil.ReadDelimited(data, blockRef); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				blockRefs = append(blockRefs, blockRef)
			}

			if err := tree.PutFile(filePath, blockRefs); err != nil {
				return err
			}
		}
	}

	finishedTree, err := tree.Finish()
	// Serialize the tree
	data, err := hashtree.Serialize(finishedTree)
	if err != nil {
		return err
	}

	if len(data) > 0 {
		// Put the tree into the blob store
		objClient, err := d.getObjectClient()
		if err != nil {
			return err
		}

		obj, err := objClient.PutObject(bytes.NewReader(data))
		if err != nil {
			return err
		}

		commitInfo.Tree = obj
	}

	commitInfo.SizeBytes = uint64(finishedTree.Size())
	commitInfo.Finished = now()

	_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		commits := d.commits(commit.Repo.Name).ReadWrite(stm)
		repos := d.repos.ReadWrite(stm)

		commits.Put(commit.ID, commitInfo)
		// update repo size
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
			return err
		}
		repoInfo.SizeBytes += commitInfo.SizeBytes
		repos.Put(commit.Repo.Name, repoInfo)
		return nil
	})
	return err
}

func (d *driver) InspectCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	if err := d.resolveBranch(ctx, commit); err != nil {
		return nil, err
	}

	commits := d.commits(commit.Repo.Name).ReadOnly(ctx)
	commitInfo := &pfs.CommitInfo{}
	if err := commits.Get(commit.ID, commitInfo); err != nil {
		return nil, err
	}
	return commitInfo, nil
}

func (d *driver) ListCommit(ctx context.Context, repo *pfs.Repo, to *pfs.Commit, from *pfs.Commit, number uint64) ([]*pfs.CommitInfo, error) {
	if from != nil && from.Repo.Name != repo.Name || to != nil && to.Repo.Name != repo.Name {
		return nil, fmt.Errorf("`from` and `to` commits need to be from repo %s", repo.Name)
	}

	if err := d.resolveBranch(ctx, from); err != nil {
		return nil, err
	}
	if err := d.resolveBranch(ctx, to); err != nil {
		return nil, err
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
			number -= 1
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
			number -= 1
		}
	}
	return commitInfos, nil
}

type commitInfoIterator struct {
	ctx    context.Context
	driver *driver
	buffer []*pfs.CommitInfo
	// a watcher that watches for updates to a branch
	newCommitEvents watch.EventChan
	// record whether a commit has been seen
	seen map[string]bool
}

func (c *commitInfoIterator) Next() (*pfs.CommitInfo, error) {
	if len(c.buffer) == 0 {
	receiveNewCommit:
		var branchName string
		commit := new(pfs.Commit)
		for {
			event := <-c.newCommitEvents
			switch event.Type {
			case watch.EventError:
				return nil, event.Err
			case watch.EventPut:
				event.Unmarshal(&branchName, commit)
			case watch.EventDelete:
				continue
			}
			if !c.seen[commit.ID] {
				break
			}
		}
		// Now we watch the CommitInfo until the commit has been finished
		commits := c.driver.commits(commit.Repo.Name).ReadOnly(c.ctx)
		commitInfoEvents := commits.WatchOne(commit.ID)
		for {
			var commitID string
			commitInfo := new(pfs.CommitInfo)
			event := <-commitInfoEvents
			switch event.Type {
			case watch.EventError:
				return nil, event.Err
			case watch.EventPut:
				event.Unmarshal(&commitID, commitInfo)
			case watch.EventDelete:
				// if this commit that we are waiting for is deleted, then
				// we go back to watch the branch to get a new commit
				goto receiveNewCommit
			}
			if commitInfo.Finished != nil {
				c.buffer = append(c.buffer, commitInfo)
				break
			}
		}
	}
	// We pop the buffer from the end because the buffer is ordered such
	// that the newer commits come first, but we want to return the older
	// commits first.  The buffered is ordered this way because that's how
	// ListCommit() orders the returned commits, and we use ListCommit to
	// populate the initial buffer.
	commitInfo := c.buffer[len(c.buffer)-1]
	c.buffer = c.buffer[:len(c.buffer)-1]
	c.seen[commitInfo.Commit.ID] = true
	return commitInfo, nil
}

func (d *driver) SubscribeCommit(ctx context.Context, repo *pfs.Repo, branch string, from *pfs.Commit) (CommitInfoIterator, error) {
	if from != nil && from.Repo.Name != repo.Name {
		return nil, fmt.Errorf("the `from` commit needs to be from repo %s", repo.Name)
	}

	// We need to watch for new commits before we start listing commits,
	// because otherwise we might miss some commits in between when we
	// finish listing and when we start watching.
	branches := d.branches(repo.Name).ReadOnly(ctx)
	newCommitEvents := branches.WatchOne(branch)

	iterator := &commitInfoIterator{
		ctx:             ctx,
		driver:          d,
		newCommitEvents: newCommitEvents,
		seen:            make(map[string]bool),
	}

	// include all commits that are currently on the given branch, but only
	// the ones that have been finished
	commitInfos, err := d.ListCommit(ctx, repo, &pfs.Commit{
		Repo: repo,
		ID:   branch,
	}, from, 0)
	if err != nil {
		return nil, err
	}
	for _, commitInfo := range commitInfos {
		if commitInfo.Finished != nil {
			iterator.buffer = append(iterator.buffer, commitInfo)
		}
	}

	return iterator, nil
}

func (d *driver) FlushCommit(ctx context.Context, fromCommits []*pfs.Commit, toRepos []*pfs.Repo) (watch.EventChan, error) {
	var repos []*pfs.Repo
	if toRepos != nil {
		repos = toRepos
	} else {
		var downstreamRepos []*pfs.Repo
		// keep track of how many times a repo appears downstream of
		// a repo in fromCommits.
		repoCounts := make(map[string]int)
		// Find the repos that have *all* the given repos as provenance
		for _, commit := range fromCommits {
			// get repos that have the commit's repo as provenance
			repoInfos, err := d.flushRepo(ctx, commit.Repo)
			if err != nil {
				return nil, err
			}

		NextRepoInfo:
			for _, repoInfo := range repoInfos {
				repoCounts[repoInfo.Repo.Name]++
				for _, repo := range downstreamRepos {
					if repoInfo.Repo.Name == repo.Name {
						// Already in the list; skip it
						continue NextRepoInfo
					}
				}
				downstreamRepos = append(downstreamRepos, repoInfo.Repo)
			}
		}
		for _, repo := range downstreamRepos {
			// Only the repos that showed up as a downstream repo for
			// len(fromCommits) repos will contain commits that are
			// downstream of all fromCommits.
			if repoCounts[repo.Name] == len(fromCommits) {
				repos = append(repos, repo)
			}
		}
	}

	// A commit needs to show up len(fromCommits) times in order to
	// prove that it indeed has all the fromCommits as provenance.
	commitCounts := make(map[string]int)
	var commitCountsLock sync.Mutex
	// When we've sent len(repos) commits, we are done
	var numCommitsSent int
	commitChan := make(watch.EventChan, len(repos))
	for _, commit := range fromCommits {
		for _, repo := range repos {
			commitEvents := d.commits(repo.Name).ReadOnly(ctx).WatchByIndex(provenanceIndex, commit)
			go func(commit *pfs.Commit) {
				ev, ok := <-commitEvents
				if !ok {
					commitChan <- &watch.Event{
						Type: watch.EventError,
						Err:  fmt.Errorf("stream for commits that have %v as provenance closed unexpectedly", commit),
					}
					return
				}
				commitID := string(ev.Key)
				commitCountsLock.Lock()
				defer commitCountsLock.Unlock()
				commitCounts[commitID]++
				if commitCounts[commitID] == len(fromCommits) {
					// The reason we only receive one event from the
					// cannel is that We only expect to receive one
					// commit per repo, because usually there's only
					// one commit per repo that has a given commit as
					// provenance.
					// Wrapping it in a select in case a send fails
					// because the other side closed.
					commitChan <- ev
					numCommitsSent++
					if numCommitsSent == len(repos) {
						close(commitChan)
					}
				}
			}(commit)
		}
	}
	return commitChan, nil
}

func (d *driver) flushRepo(ctx context.Context, repo *pfs.Repo) ([]*pfs.RepoInfo, error) {
	iter, err := d.repos.ReadOnly(ctx).GetByIndex(provenanceIndex, repo)
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
	panic("unreachable")
}

func (d *driver) DeleteCommit(ctx context.Context, commit *pfs.Commit) error {
	return nil
}

func (d *driver) ListBranch(ctx context.Context, repo *pfs.Repo) ([]*pfs.Branch, error) {
	branches := d.branches(repo.Name).ReadOnly(ctx)
	iterator, err := branches.List()
	if err != nil {
		return nil, err
	}

	var res []*pfs.Branch
	for {
		var branchName string
		head := new(pfs.Commit)
		ok, err := iterator.Next(&branchName, head)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		res = append(res, &pfs.Branch{
			Name: path.Base(branchName),
			Head: head,
		})
	}
	return res, nil
}

func (d *driver) SetBranch(ctx context.Context, commit *pfs.Commit, name string) error {
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		commits := d.commits(commit.Repo.Name).ReadWrite(stm)
		branches := d.branches(commit.Repo.Name).ReadWrite(stm)

		// Make sure that the commit exists
		var commitInfo pfs.CommitInfo
		if err := commits.Get(commit.ID, &commitInfo); err != nil {
			return err
		}

		branches.Put(name, commit)
		return nil
	})
	return err
}

func (d *driver) DeleteBranch(ctx context.Context, repo *pfs.Repo, name string) error {
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		branches := d.branches(repo.Name).ReadWrite(stm)
		return branches.Delete(name)
	})
	return err
}

// resolveBranch replaces a branch with a real commit ID, e.g. "master" ->
// UUID.
// If the given commit already contains a real commit ID, then this
// function does nothing.
func (d *driver) resolveBranch(ctx context.Context, commit *pfs.Commit) error {
	if commit == nil {
		return nil
	}
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		branches := d.branches(commit.Repo.Name).ReadWrite(stm)

		head := new(pfs.Commit)
		// See if we are given a branch
		if err := branches.Get(commit.ID, head); err != nil {
			if _, ok := err.(col.ErrNotFound); !ok {
				return err
			}
			// If it's not a branch, use it as it is
			return nil
		}
		commit.ID = head.ID
		return nil
	})
	return err
}

// scratchCommitPrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchCommitPrefix(ctx context.Context, commit *pfs.Commit) (string, error) {
	if err := d.resolveBranch(ctx, commit); err != nil {
		return "", err
	}
	return path.Join(d.prefix, "scratch", commit.Repo.Name, commit.ID), nil
}

// scratchFilePrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchFilePrefix(ctx context.Context, file *pfs.File) (string, error) {
	return path.Join(d.prefix, "scratch", file.Commit.Repo.Name, file.Commit.ID, file.Path), nil
}

// checkPath checks if a file path is legal
func checkPath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("filename cannot contain null character: %s", path)
	}
	return nil
}

func (d *driver) PutFile(ctx context.Context, file *pfs.File, reader io.Reader) error {
	if err := checkPath(file.Path); err != nil {
		return err
	}

	// Put the tree into the blob store
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}

	blockRefs, err := blockClient.PutBlock(pfs.Delimiter_NONE, reader)
	if err != nil {
		return err
	}

	// Serialize the blockrefs.
	// Since we are serializing multiple blockrefs, we use WriteDelimited
	// to write them into the same object, since protobuf messages are not
	// self-delimiting.
	var buffer bytes.Buffer
	for _, blockRef := range blockRefs.BlockRef {
		if _, err := pbutil.WriteDelimited(&buffer, blockRef); err != nil {
			return err
		}
	}

	if err := d.resolveBranch(ctx, file.Commit); err != nil {
		return err
	}

	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	_, err = d.etcdClient.Put(ctx, path.Join(prefix, uuid.NewWithoutDashes()), buffer.String())
	return err
}
func (d *driver) MakeDirectory(ctx context.Context, file *pfs.File) error {
	return nil
}

func (d *driver) getTreeForCommit(ctx context.Context, commit *pfs.Commit) (hashtree.HashTree, error) {
	if commit == nil {
		t, err := hashtree.NewHashTree().Finish()
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	if err := d.resolveBranch(ctx, commit); err != nil {
		return nil, err
	}

	// TODO: get the tree from a cache
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
	objClient, err := d.getObjectClient()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := objClient.GetObject(treeRef.Hash, &buf); err != nil {
		return nil, err
	}

	h, err := hashtree.Deserialize(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (d *driver) GetFile(ctx context.Context, file *pfs.File, offset int64, size int64) (io.ReadCloser, error) {
	tree, err := d.getTreeForCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}

	node, err := tree.Get(file.Path)
	if err != nil {
		return nil, err
	}

	if node.FileNode == nil {
		return nil, fmt.Errorf("%s is a directory", file.Path)
	}

	return d.newFileReader(node.FileNode.BlockRefs, file, offset, size)
}

type fileReader struct {
	blockClient *client.APIClient
	reader      io.Reader
	offset      int64
	size        int64 // how much data to read
	sizeRead    int64 // how much data has been read
	blockRefs   []*pfs.BlockRef
	file        *pfs.File
}

func (d *driver) newFileReader(blockRefs []*pfs.BlockRef, file *pfs.File, offset int64, size int64) (*fileReader, error) {
	blockClient, err := d.getBlockClient()
	if err != nil {
		return nil, err
	}

	return &fileReader{
		blockClient: blockClient,
		blockRefs:   blockRefs,
		offset:      offset,
		size:        size,
		file:        file,
	}, nil
}

func (r *fileReader) Read(data []byte) (int, error) {
	var err error
	if r.reader == nil {
		var blockRef *pfs.BlockRef
		for {
			if len(r.blockRefs) == 0 {
				return 0, io.EOF
			}
			blockRef = r.blockRefs[0]
			r.blockRefs = r.blockRefs[1:]
			blockSize := int64(blockRef.Range.Upper - blockRef.Range.Lower)
			if r.offset >= blockSize {
				r.offset -= blockSize
				continue
			}
			break
		}
		sizeLeft := r.size
		// e.g. sometimes a reader is constructed of size 0
		if sizeLeft != 0 {
			sizeLeft -= r.sizeRead
		}
		r.reader, err = r.blockClient.GetBlock(blockRef.Block.Hash, uint64(r.offset), uint64(sizeLeft))
		if err != nil {
			return 0, err
		}
		r.offset = 0
	}
	size, err := r.reader.Read(data)
	if err != nil && err != io.EOF {
		return size, err
	}
	if err == io.EOF {
		r.reader = nil
	}
	r.sizeRead += int64(size)
	if r.sizeRead == r.size {
		return size, io.EOF
	}
	if r.size > 0 && r.sizeRead > r.size {
		return 0, fmt.Errorf("read more than we need; this is likely a bug")
	}
	return size, nil
}

func (r *fileReader) Close() error {
	return nil
}

func nodeToFileInfo(commit *pfs.Commit, path string, node *hashtree.NodeProto) *pfs.FileInfo {
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
	} else if node.DirNode != nil {
		fileInfo.FileType = pfs.FileType_DIR
		fileInfo.Children = node.DirNode.Children
	}
	return fileInfo
}

func (d *driver) InspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	tree, err := d.getTreeForCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}

	node, err := tree.Get(file.Path)
	if err != nil {
		return nil, err
	}

	return nodeToFileInfo(file.Commit, file.Path, node), nil
}

func (d *driver) ListFile(ctx context.Context, file *pfs.File) ([]*pfs.FileInfo, error) {
	tree, err := d.getTreeForCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}

	nodes, err := tree.List(file.Path)
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for _, node := range nodes {
		fileInfos = append(fileInfos, nodeToFileInfo(file.Commit, path.Join(file.Path, node.Name), node))
	}
	return fileInfos, nil
}

func (d *driver) GlobFile(ctx context.Context, commit *pfs.Commit, pattern string) ([]*pfs.FileInfo, error) {
	tree, err := d.getTreeForCommit(ctx, commit)
	if err != nil {
		return nil, err
	}

	nodes, err := tree.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for _, node := range nodes {
		fileInfos = append(fileInfos, nodeToFileInfo(commit, node.Name, node))
	}
	return fileInfos, nil
}

func (d *driver) DeleteFile(ctx context.Context, file *pfs.File) error {
	if err := d.resolveBranch(ctx, file.Commit); err != nil {
		return err
	}

	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	_, err = d.etcdClient.Put(ctx, path.Join(prefix, uuid.NewWithoutDashes()), TOMBSTONE)
	return err
}

func (d *driver) DeleteAll(ctx context.Context) error {
	return nil
}
func (d *driver) Dump(ctx context.Context) {
}
