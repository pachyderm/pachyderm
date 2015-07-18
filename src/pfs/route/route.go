package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/discovery"
	"google.golang.org/grpc"
)

type Sharder interface {
	NumShards() int
	GetShard(path *pfs.Path) (int, error)
}

func NewSharder(numShards int) Sharder {
	return newSharder(numShards)
}

type Addresser interface {
	GetMasterShards(address string) (map[int]bool, error)
	GetSlaveShards(address string) (map[int]bool, error)
	GetAllAddresses() ([]string, error)
}

func NewSingleAddresser(address string, numShards int) Addresser {
	return newSingleAddresser(address, numShards)
}

func NewDiscoveryAddresser(discoveryClient discovery.Client) Addresser {
	return newDiscoveryAddresser(discoveryClient)
}

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}

type Router interface {
	GetMasterShards() (map[int]bool, error)
	GetSlaveShards() (map[int]bool, error)
	GetMasterClientConn(shard int) (*grpc.ClientConn, error)
	GetMasterOrSlaveClientConn(shard int) (*grpc.ClientConn, error)
	GetAllClientConns() ([]*grpc.ClientConn, error)
}

func NewRouter(
	addresser Addresser,
	dialer Dialer,
	localAddress string,
) Router {
	return newRouter(
		addresser,
		dialer,
		localAddress,
	)
}
