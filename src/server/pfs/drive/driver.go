package drive

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

type driver struct {
	blockClient pfs.BlockAPIClient
	etcdClient  *etcd.Client
	prefix      string
}

// NewDriver is used to create a new Driver instance
func NewDriver(blockAddress string, etcdAddresses []string, etcdPrefix string) (Driver, error) {
	clientConn, err := grpc.Dial(blockAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   etcdAddresses,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &driver{
		blockClient: pfs.NewBlockAPIClient(clientConn),
		etcdClient:  etcdClient,
		prefix:      etcdPrefix,
	}, nil
}

// NewLocalDriver creates a driver using an local etcd instance.  This
// function is intended for testing purposes
func NewLocalDriver(blockAddress string, etcdPrefix string) (Driver, error) {
	clientConn, err := grpc.Dial(blockAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &driver{
		blockClient: pfs.NewBlockAPIClient(clientConn),
		etcdClient:  etcdClient,
		prefix:      etcdPrefix,
	}, nil
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

	repos := d.repos(ctx)

	if err := repos.Lock(); err != nil {
		return err
	}
	defer repos.Unlock()

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
	}

	repoInfo := &pfs.RepoInfo{
		Repo:       repo,
		Created:    now(),
		Provenance: fullProvRepos,
	}
	return repos.Create(repo.Name, repoInfo)
}

func (d *driver) InspectRepo(ctx context.Context, repo *pfs.Repo) (*pfs.RepoInfo, error) {
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos(ctx).Get(repo.Name, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	repos := d.repos(ctx)
	// Ensure that all provenance repos exist
	for _, prov := range provenance {
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(prov.Name, repoInfo); err != nil {
			return nil, err
		}
	}

	var result []*pfs.RepoInfo
	iterate, err := d.repos(ctx).List()
	if err != nil {
		return nil, err
	}
nextRepo:
	for {
		repoName, repoInfo := "", pfs.RepoInfo{}
		ok, err := iterate(&repoName, &repoInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
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
	return result, nil
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	repos := d.repos(ctx)

	// To understand why a lock is necessary here, consider the following
	// order of execution:
	//
	// 1. DeleteRepo(foo) begins.  It scans all repos and concludes that
	// foo is not the provenance of any repo, therefore it's safe to delete
	// this foo
	// 2. CreateRepo(bar, provenance=foo) executes.  It creates bar with
	// foo as the provenance.
	// 3. DeleteRepo(foo) finishes by deleting foo.
	//
	// Now we end up with a repo bar with a nonexistent provenance foo.
	if err := repos.Lock(); err != nil {
		return err
	}
	defer repos.Unlock()

	// Check if this repo is the provenance of some other repos
	iterate, err := repos.List()
	if err != nil {
		return err
	}
	if !force {
		for {
			key, repoInfo := "", pfs.RepoInfo{}
			ok, err := iterate(&key, &repoInfo)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			for _, prov := range repoInfo.Provenance {
				if prov.Name == repo.Name {
					return fmt.Errorf("cannot delete repo %s because it's the provenance of repo %s", repo.Name, prov.Name)
				}
			}
		}
	}

	repoKey := repos.path(repo.Name)
	resp, err := d.etcdClient.Txn(ctx).If(present(repoKey)).Then(
		etcd.OpDelete(repoKey),
		etcd.OpDelete(d.commits(ctx)(repo.Name).path(""), etcd.WithPrefix()),
		etcd.OpDelete(d.refs(ctx)(repo.Name).path(""), etcd.WithPrefix())).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("repo %s doesn't exist", repo.Name)
	}

	return nil
}

func (d *driver) StartCommit(ctx context.Context, parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error) {
	commit := &pfs.Commit{
		Repo: parent.Repo,
		ID:   uuid.NewWithoutDashes(),
	}
	_, err := concurrency.NewSTMRepeatable(ctx, d.etcdClient, func(stm concurrency.STM) error {
		repos := d.repos(ctx).stm(stm)
		commits := d.commits(ctx)(parent.Repo.Name).stm(stm)
		refs := d.refs(ctx)(parent.Repo.Name).stm(stm)

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
			if err := refs.Put(ref.Name, ref); err != nil {
				return err
			}
		}
		return commits.Create(commit.ID, commitInfo)
	})
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (d *driver) FinishCommit(ctx context.Context, commit *pfs.Commit) error {
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

func (d *driver) PutFile(ctx context.Context, file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) error {
	return nil
}
func (d *driver) MakeDirectory(ctx context.Context, file *pfs.File) error {
	return nil
}
func (d *driver) GetFile(ctx context.Context, file *pfs.File, offset int64, size int64) (io.ReadCloser, error) {
	return nil, nil
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
