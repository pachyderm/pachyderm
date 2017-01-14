package drive

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"

	etcd "github.com/coreos/etcd/clientv3"
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

	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
		repos := d.repos(stm)

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
		commits := d.commits(stm)(repo.Name)
		refs := d.refs(stm)(repo.Name)
		// Check if this repo is the provenance of some other repos
		if !force {
			iterate, err := repos.List()
			if err != nil {
				return err
			}
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

		if err := repos.Delete(repo.Name); err != nil {
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
	_, err := newSTM(ctx, d.etcdClient, func(stm STM) error {
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
