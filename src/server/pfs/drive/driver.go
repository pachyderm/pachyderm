package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"google.golang.org/grpc"
)

type driver struct {
	blockAddress    string
	blockClientOnce sync.Once
	blockClient     *client.APIClient
	etcdClient      *etcd.Client
	prefix          string
}

const (
	TOMBSTONE = "delete"
)

// NewDriver is used to create a new Driver instance
func NewDriver(blockAddress string, etcdAddresses []string, etcdPrefix string) (Driver, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddresses,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &driver{
		blockAddress: blockAddress,
		etcdClient:   etcdClient,
		prefix:       etcdPrefix,
	}, nil
}

// NewLocalDriver creates a driver using an local etcd instance.  This
// function is intended for testing purposes
func NewLocalDriver(blockAddress string, etcdPrefix string) (Driver, error) {
	return NewDriver(blockAddress, []string{"localhost:2379"}, etcdPrefix)
}

func (d *driver) getBlockClient() (*client.APIClient, error) {
	if d.blockClient == nil {
		var onceErr error
		// Be thread safe
		d.blockClientOnce.Do(func() {
			clientConn, err := grpc.Dial(d.blockAddress, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			d.blockClient = &client.APIClient{BlockAPIClient: pfs.NewBlockAPIClient(clientConn)}
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return d.blockClient, nil
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

	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		repos := d.repos(stm)
		repoRefCounts := d.repoRefCounts(stm)

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
	if err := d.reposReadonly(ctx).Get(repo.Name, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	var result []*pfs.RepoInfo
	repos := d.reposReadonly(ctx)
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
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		repos := d.repos(stm)
		repoRefCounts := d.repoRefCounts(stm)
		commits := d.commits(stm)(repo.Name)
		branches := d.branches(stm)(repo.Name)

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
	commit := &pfs.Commit{
		Repo: parent.Repo,
		ID:   uuid.NewWithoutDashes(),
	}
	if _, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		repos := d.repos(stm)
		commits := d.commits(stm)(parent.Repo.Name)
		branches := d.branches(stm)(parent.Repo.Name)

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
			provCommits := d.commits(stm)(prov.Repo.Name)
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
				if _, ok := err.(ErrNotFound); !ok {
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

	if _, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		commits := d.commits(stm)(commit.Repo.Name)
		repos := d.repos(stm)

		commitInfo := new(pfs.CommitInfo)
		if err := commits.Get(commit.ID, commitInfo); err != nil {
			return err
		}

		if commitInfo.Finished != nil {
			return fmt.Errorf("commit %s has already been finished", commit.FullID())
		}

		tree, err := d.getTreeForCommit(ctx, commitInfo.ParentCommit)
		if err != nil {
			return err
		}

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

		// Serialize the tree
		data, err := proto.Marshal(tree)
		if err != nil {
			return err
		}

		// Put the tree into the blob store
		blockClient, err := d.getBlockClient()
		if err != nil {
			return err
		}

		blockRefs, err := blockClient.PutBlock(pfs.Delimiter_NONE, bytes.NewReader(data))
		if err != nil {
			return err
		}

		if len(blockRefs.BlockRef) > 0 {
			// TODO: the block store might break up the tree into multiple
			// blocks if the tree is big enough, in which case the following
			// code won't work.  This shouldn't be a problem once we migrate
			// to use the tag store, which puts everything as one block.
			commitInfo.Tree = blockRefs.BlockRef[0]
		}

		// update commit size
		root, err := tree.Get("/")
		// the tree might be empty (if the commit is empty), in which case
		// the library returns a PathNotFound error
		if err != nil && hashtree.Code(err) != hashtree.PathNotFound {
			return err
		}
		if root != nil {
			commitInfo.SizeBytes = uint64(root.SubtreeSize)
		}
		commitInfo.Finished = now()
		commits.Put(commit.ID, commitInfo)

		// update repo size
		repoInfo := new(pfs.RepoInfo)
		if err := repos.Get(commit.Repo.Name, repoInfo); err != nil {
			return err
		}
		repoInfo.SizeBytes += commitInfo.SizeBytes
		repos.Put(commit.Repo.Name, repoInfo)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *driver) InspectCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	if err := d.resolveBranch(ctx, commit); err != nil {
		return nil, err
	}

	commits := d.commitsReadonly(ctx)(commit.Repo.Name)
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
	commits := d.commitsReadonly(ctx)(repo.Name)

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
	// an iterator that receives new commits
	newCommitsIter IterateCloser
	// record whether a commit has been seen
	seen map[string]bool
}

func (c *commitInfoIterator) Next() (*pfs.CommitInfo, error) {
	if len(c.buffer) == 0 {
		var commitID string
		commit := new(pfs.Commit)
		for {
			ok, err := c.newCommitsIter.Next(&commitID, commit)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, nil
			}
			if !c.seen[commitID] {
				break
			}
		}
		// Now we watch the CommitInfo until the commit has been finished
		commits := c.driver.commitsReadonly(c.ctx)(commit.Repo.Name)
		commitInfoIter, err := commits.WatchOne(commit.ID)
		if err != nil {
			return nil, err
		}
		for {
			var commitID string
			commitInfo := new(pfs.CommitInfo)
			ok, err := commitInfoIter.Next(&commitID, commitInfo)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("unable to wait until commit %s finishes", commit.ID)
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

func (c *commitInfoIterator) Close() error {
	return c.newCommitsIter.Close()
}

func (d *driver) SubscribeCommit(ctx context.Context, repo *pfs.Repo, branch string, from *pfs.Commit) (CommitInfoIterator, error) {
	if from != nil && from.Repo.Name != repo.Name {
		return nil, fmt.Errorf("the `from` commit needs to be from repo %s", repo.Name)
	}

	// We need to watch for new commits before we start listing commits,
	// because otherwise we might miss some commits in between when we
	// finish listing and when we start watching.
	branches := d.branchesReadonly(ctx)(repo.Name)
	newCommitsIter, err := branches.WatchOne(branch)
	if err != nil {
		return nil, err
	}

	iterator := &commitInfoIterator{
		ctx:            ctx,
		driver:         d,
		newCommitsIter: newCommitsIter,
		seen:           make(map[string]bool),
	}
	commitInfos, err := d.ListCommit(ctx, repo, &pfs.Commit{
		Repo: repo,
		ID:   branch,
	}, from, 0)
	if err != nil {
		return nil, err
	}
	iterator.buffer = commitInfos
	return iterator, nil
}

func (d *driver) FlushCommit(ctx context.Context, fromCommits []*pfs.Commit, toRepos []*pfs.Repo) (CommitInfoIterator, error) {
	// watch /prov/repo/commit/commit
	return nil, nil
}

func (d *driver) DeleteCommit(ctx context.Context, commit *pfs.Commit) error {
	return nil
}

func (d *driver) ListBranch(ctx context.Context, repo *pfs.Repo) ([]*pfs.Branch, error) {
	branches := d.branchesReadonly(ctx)(repo.Name)
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
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		commits := d.commits(stm)(commit.Repo.Name)
		branches := d.branches(stm)(commit.Repo.Name)

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
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		branches := d.branches(stm)(repo.Name)
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
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		branches := d.branches(stm)(commit.Repo.Name)

		head := new(pfs.Commit)
		// See if we are given a branch
		if err := branches.Get(commit.ID, head); err != nil {
			if _, ok := err.(ErrNotFound); !ok {
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

func (d *driver) getTreeForCommit(ctx context.Context, commit *pfs.Commit) (*hashtree.HashTreeProto, error) {
	if commit == nil {
		return &hashtree.HashTreeProto{}, nil
	}

	if err := d.resolveBranch(ctx, commit); err != nil {
		return nil, err
	}

	// TODO: get the tree from a cache
	var treeRef *pfs.BlockRef
	if _, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		commits := d.commits(stm)(commit.Repo.Name)
		commitInfo := &pfs.CommitInfo{}
		if err := commits.Get(commit.ID, commitInfo); err != nil {
			return err
		}
		if commitInfo.Finished == nil {
			return fmt.Errorf("cannot read from an open commit")
		}
		treeRef = commitInfo.Tree
		return nil
	}); err != nil {
		return nil, err
	}

	if treeRef == nil {
		return &hashtree.HashTreeProto{}, nil
	}

	// read the tree from the block store
	blockClient, err := d.getBlockClient()
	if err != nil {
		return nil, err
	}

	obj, err := blockClient.GetBlock(treeRef.Block.Hash, 0, 0)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	h := &hashtree.HashTreeProto{}
	if err := proto.Unmarshal(bytes, h); err != nil {
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

func nodeToFileInfo(file *pfs.File, node *hashtree.NodeProto) *pfs.FileInfo {
	fileInfo := &pfs.FileInfo{
		File: &pfs.File{
			Commit: file.Commit,
			Path:   path.Join(file.Path, node.Name),
		},
		SizeBytes: uint64(node.SubtreeSize),
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

	return nodeToFileInfo(file, node), nil
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
		fileInfos = append(fileInfos, nodeToFileInfo(file, node))
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
