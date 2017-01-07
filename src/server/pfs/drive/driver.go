package drive

import (
	"client/pfs"
	"server/pfs/drive"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
)

type driver struct {
	blockClient pfs.BlockAPIClient
	etcdClient  *clientv3.Client
	prefix      string
}

// NewDriver is used to create a new Driver instance
func NewDriver(blockAddress string, etcdAddresses []string, etcdPrefix string) (drive.Driver, error) {
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
		etcdClient:  client,
		prefix:      etcdPrefix,
	}, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo, provenance []*pfs.Repo) error {
}

func (d *driver) InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error) {

}

func (d *driver) ListRepo(provenance []*pfs.Repo) ([]*pfs.RepoInfo, error) {

}

func (d *driver) DeleteRepo(repo *pfs.Repo, force bool) error {

}
