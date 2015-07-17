package route

import (
	"math/rand"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
)

type router struct {
	addresser    address.Addresser
	dialer       dial.Dialer
	localAddress string
}

func newRouter(
	addresser address.Addresser,
	dialer dial.Dialer,
	localAddress string,
) *router {
	return &router{
		addresser,
		dialer,
		localAddress,
	}
}

func (r *router) IsLocalMasterShard(shard int) (bool, error) {
	address, err := r.addresser.GetMasterAddress(shard)
	if err != nil {
		return false, err
	}
	return address == r.localAddress, nil
}

func (r *router) IsLocalSlaveShard(shard int) (bool, error) {
	addresses, err := r.addresser.GetSlaveAddresses(shard)
	if err != nil {
		return false, err
	}
	for _, address := range addresses {
		if address == r.localAddress {
			return true, nil
		}
	}
	return false, nil
}

func (r *router) GetMasterClientConn(shard int) (*grpc.ClientConn, error) {
	address, err := r.addresser.GetMasterAddress(shard)
	if err != nil {
		return nil, err
	}
	return r.dialer.Dial(address)
}

func (r *router) GetMasterOrSlaveClientConn(shard int) (*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetSlaveAddresses(shard)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return r.GetMasterClientConn(shard)
	}
	address := addresses[int(rand.Uint32())%len(addresses)]
	return r.dialer.Dial(address)
}

func (r *router) GetAllClientConns() ([]*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	clientConns := make([]*grpc.ClientConn, len(addresses)-1)
	for i, address := range addresses {
		// TODO(pedge): huge race, this whole thing is bad
		if address != r.localAddress {
			clientConn, err := r.dialer.Dial(address)
			if err != nil {
				return nil, err
			}
			clientConns[i] = clientConn
		}
	}
	return clientConns, nil
}
