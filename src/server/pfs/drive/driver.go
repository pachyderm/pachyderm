package drive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
			provRepo := &pfs.RepoInfo{}
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
	repoInfo := &pfs.RepoInfo{}
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		return d.repos(stm).Get(repo.Name, repoInfo)
	})
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	var result []*pfs.RepoInfo
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		result = nil
		repos := d.repos(stm)
		// Ensure that all provenance repos exist
		for _, prov := range provenance {
			repoInfo := &pfs.RepoInfo{}
			if err := repos.Get(prov.Name, repoInfo); err != nil {
				return err
			}
		}

		iterate, err := repos.List()
		if err != nil {
			return err
		}
	nextRepo:
		for {
			repoName, repoInfo := "", pfs.RepoInfo{}
			ok, err := iterate(&repoName, &repoInfo)
			if err != nil {
				return err
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
			result = append(result, &repoInfo)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		repos := d.repos(stm)
		repoRefCounts := d.repoRefCounts(stm)
		commits := d.commits(stm)(repo.Name)
		refs := d.refs(stm)(repo.Name)

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

		repoInfo := &pfs.RepoInfo{}
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
		refs.DeleteAll()
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
		refs := d.refs(stm)(parent.Repo.Name)

		// Check if repo exists
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(parent.Repo.Name, repoInfo); err != nil {
			return err
		}

		commitInfo := &pfs.CommitInfo{
			Commit:     commit,
			Started:    now(),
			Provenance: provenance,
		}

		if parent.ID != "" {
			ref := &pfs.Ref{}
			// See if we are given a ref
			if err := refs.Get(parent.ID, ref); err != nil {
				if _, ok := err.(ErrNotFound); !ok {
					return err
				}
				// If parent is not a ref, it needs to be a commit
				// Check that the parent commit exists
				parentCommitInfo := &pfs.CommitInfo{}
				if err := commits.Get(parent.ID, parentCommitInfo); err != nil {
					return err
				}
				commitInfo.ParentCommit = parent
			} else {
				commitInfo.ParentCommit = &pfs.Commit{
					Repo: parent.Repo,
					ID:   ref.Commit.ID,
				}
				ref.Commit = commit
				refs.Put(ref.Name, ref)
			}
		}
		return commits.Create(commit.ID, commitInfo)
	}); err != nil {
		return nil, err
	}

	return commit, nil
}

func (d *driver) FinishCommit(ctx context.Context, commit *pfs.Commit) error {
	if _, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		if err := d.resolveRef(ctx, commit); err != nil {
			return err
		}

		prefix, err := d.scratchCommitPrefix(ctx, commit)
		if err != nil {
			return err
		}

		// Read everything under the scratch space for this commit
		// TODO: lock the scratch space to prevent concurrent PutFile
		resp, err := d.etcdClient.Get(ctx, prefix, etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortAscend))
		if err != nil {
			return err
		}

		commits := d.commits(stm)(commit.Repo.Name)
		commitInfo := &pfs.CommitInfo{}
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

		blockrefs, err := blockClient.PutBlock(pfs.Delimiter_NONE, bytes.NewReader(data))
		if err != nil {
			return err
		}

		commitInfo.Tree = blockrefs.BlockRef[0]
		commitInfo.Finished = now()
		commits.Put(commit.ID, commitInfo)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Squash merges the content of fromCommits into a single commit with
// the given parent.
func (d *driver) SquashCommit(ctx context.Context, fromCommits []*pfs.Commit, parent *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) InspectCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) ListCommit(ctx context.Context, repo *pfs.Repo, from *pfs.Commit, to *pfs.Commit, number uint64) ([]*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) FlushCommit(ctx context.Context, fromCommits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) DeleteCommit(ctx context.Context, commit *pfs.Commit) error {
	return nil
}

func (d *driver) ListBranch(ctx context.Context, repo *pfs.Repo) ([]string, error) {
	return nil, nil
}

func (d *driver) SetBranch(ctx context.Context, commit *pfs.Commit, name string) error {
	return nil
}

func (d *driver) RenameBranch(ctx context.Context, repo *pfs.Repo, from string, to string) error {
	return nil
}

// resolveRef replaces a reference with a real commit ID, e.g. "master" ->
// UUID.
// If the given commit already contains a real commit ID, then this
// function does nothing.
func (d *driver) resolveRef(ctx context.Context, commit *pfs.Commit) error {
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		refs := d.refs(stm)(commit.Repo.Name)

		ref := &pfs.Ref{}
		// See if we are given a ref
		if err := refs.Get(commit.ID, ref); err != nil {
			if _, ok := err.(ErrNotFound); !ok {
				return err
			}
			// If it's not a ref, use it as it is
			return nil
		}
		commit.ID = ref.Commit.ID
		return nil
	})
	return err
}

// scratchCommitPrefix returns an etcd prefix that's used to temporarily
// store the state of a file in an open commit.  Once the commit is finished,
// the scratch space is removed.
func (d *driver) scratchCommitPrefix(ctx context.Context, commit *pfs.Commit) (string, error) {
	if err := d.resolveRef(ctx, commit); err != nil {
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

func (d *driver) PutFile(ctx context.Context, file *pfs.File, reader io.Reader) error {
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

	if err := d.resolveRef(ctx, file.Commit); err != nil {
		return err
	}

	prefix, err := d.scratchFilePrefix(ctx, file)
	if err != nil {
		return err
	}

	_, err = d.newSequentialKV(ctx, prefix, buffer.String())
	return err
}
func (d *driver) MakeDirectory(ctx context.Context, file *pfs.File) error {
	return nil
}

func (d *driver) getTreeForCommit(ctx context.Context, commit *pfs.Commit) (*hashtree.HashTreeProto, error) {
	if commit == nil {
		return &hashtree.HashTreeProto{}, nil
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

func (d *driver) InspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	return nil, nil
}
func (d *driver) ListFile(ctx context.Context, file *pfs.File) ([]*pfs.FileInfo, error) {
	return nil, nil
}
func (d *driver) DeleteFile(ctx context.Context, file *pfs.File) error {
	return nil
}

func (d *driver) DeleteAll(ctx context.Context) error {
	return nil
}
func (d *driver) Dump(ctx context.Context) {
}
