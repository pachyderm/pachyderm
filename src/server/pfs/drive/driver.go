package drive

import (
	"context"
	"fmt"
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
	resp, err := d.etcdClient.Txn(ctx).If(absent(d.repos(repo.Name))).Then(etcd.OpPut(repo.Name, &pfs.RepoInfo{
		Repo:       repo,
		Created:    now(),
		Provenance: provenance,
	}.String())).Commit()
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
		return fmt.Errorf("repo %s not found", repo.Name)
	}
	repoInfo := &pfs.RepoInfo{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, repoInfo); err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(ctx context.Context, provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {
}

func (d *driver) DeleteRepo(ctx context.Context, repo *pfs.Repo) error {
}
