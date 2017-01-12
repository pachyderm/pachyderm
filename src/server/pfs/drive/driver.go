package drive

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	reposPrefix = "/repos"
)

type driver struct {
	blockClient pfs.BlockAPIClient
	etcdClient  *etcd.Client
	prefix      string
}

type collectionInterface interface {
	Get(key string, val proto.Message) error
	Put(key string, val proto.Message) error
	Create(key string, val proto.Message) error
	List(key *string, val proto.Message, iterate func() error) error
	Delete(key string) error
}

type collection struct {
	ctx        context.Context
	etcdClient *etcd.Client
	prefix     string
}

func (c *collection) path(key string) string {
	return path.Join(c.prefix, key)
}

func (c *collection) Get(key string, val proto.Message) error {
	resp, err := c.etcdClient.Get(c.ctx, c.path(key))
	if err != nil {
		return err
	}
	if resp.Count == 0 {
		return fmt.Errorf("%s %s not found", c.prefix, key)
	}
	return proto.UnmarshalText(string(resp.Kvs[0].Value), val)
}

func (c *collection) Put(key string, val proto.Message) error {
	valBytes, err := proto.Marshal(val)
	if err != nil {
		return err
	}
	_, err = c.etcdClient.Put(c.ctx, c.path(key), string(valBytes))
	return err
}

func (c *collection) Create(key string, val proto.Message) error {
	valBytes, err := proto.Marshal(val)
	if err != nil {
		return err
	}
	resp, err := c.etcdClient.Txn(c.ctx).If(present(key)).Then(etcd.OpPut(c.path(key), string(valBytes))).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("%s %s already exists", c.prefix, key)
	}
	return nil
}

func (c *collection) List(key *string, val proto.Message, iterate func() error) error {
	resp, err := c.etcdClient.Get(c.ctx, c.path(""), etcd.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		*key = string(kv.Key)
		if err := proto.Unmarshal(kv.Value, val); err != nil {
			return err
		}
		if err := iterate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) Delete(key string) error {
	_, err := c.etcdClient.Delete(c.ctx, key)
	return err
}

func (d *driver) repos(ctx context.Context) *collection {
	return &collection{
		ctx:        ctx,
		prefix:     path.Join(d.prefix, reposPrefix),
		etcdClient: d.etcdClient,
	}
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
	_, err := concurrency.NewSTMRepeatable(ctx, d.etcdClient, func(stm concurrency.STM) error {
		for _, prov := range provenance {
			repoVal := stm.Get(d.repos(ctx).path(prov.Name))
			if repoVal == "" {
				return pfsserver.ErrRepoNotFound{prov}
			}
		}
		repoInfo := &pfs.RepoInfo{
			Repo:       repo,
			Created:    now(),
			Provenance: provenance,
		}
		repoKey := d.repos(ctx).path(repo.Name)
		repoVal := stm.Get(repoKey)
		if repoVal != "" {
			return pfsserver.ErrRepoExists{repo}
		}
		stm.Put(repoKey, repoInfo.String())
		return nil
	})
	return err
}

func (d *driver) InspectRepo(ctx context.Context, repo *pfs.Repo) (*pfs.RepoInfo, error) {
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos(ctx).Get(repo.Name, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	var key string
	repoInfo := &pfs.RepoInfo{}
	var result []*pfs.RepoInfo
	if err := d.repos(ctx).List(&key, repoInfo, func() error {
		for _, reqProv := range provenance {
			matched := false
			for _, prov := range repoInfo.Provenance {
				if prov.Name == reqProv.Name {
					matched = true
				}
			}
			if !matched {
				return nil
			}
		}
		copy := &pfs.RepoInfo{}
		*copy = *repoInfo
		result = append(result, copy)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
	//resp, err := d.etcdClient.Txn(ctx).If(present(repoKey)).Then(
	//etcd.OpDelete(repoKey, repoInfo.String()),
	//etcd.OpDelete(d.commits(repo.Name), etcd.WithPrefix()),
	//etcd.OpDelete(d.refs(repo.Name), etcd.WithPrefix())).Commit()
	//if err != nil {
	//return err
	//}
	//if !resp.Succeeded {
	//return fmt.Errorf("repo %s doesn't exist", repo.Name)
	//}
	return nil
}

func (d *driver) StartCommit(ctx context.Context, parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) FinishCommit(ctx context.Context, commit *pfs.Commit, cancel bool) error {
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
