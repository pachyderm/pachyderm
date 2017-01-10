package drive

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

type driver struct {
	blockClient pfs.BlockAPIClient
	etcdClient  *etcd.Client
	prefix      string
}

func (d *driver) repos(repo string) string {
	return path.Join(d.prefix, reposPrefix, repo)
}

const (
	reposPrefix = "/repos"
)

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
func NewLocalDriver(blockAddress string) (Driver, error) {
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
		prefix:      "",
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
	repoInfo := &pfs.RepoInfo{
		Repo:       repo,
		Created:    now(),
		Provenance: provenance,
	}
	resp, err := d.etcdClient.Txn(ctx).If(absent(d.repos(repo.Name))).Then(etcd.OpPut(repo.Name, repoInfo.String())).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return fmt.Errorf("repo %s already exists", repo.Name)
	}
	return nil
}

func (d *driver) InspectRepo(ctx context.Context, repo *pfs.Repo) (*pfs.RepoInfo, error) {
	resp, err := d.etcdClient.Get(ctx, d.repos(repo.Name))
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	repoInfo := &pfs.RepoInfo{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
	return nil, nil
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo, force bool) error {
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
