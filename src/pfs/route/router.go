package route

import (
	"math/rand"

	"github.com/pachyderm/pachyderm/src/pfs"
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

func (r *router) GetMasterAPIClient(shard int) (pfs.ApiClient, error) {
	address, err := r.addresser.GetMasterAddress(shard)
	if err != nil {
		return nil, err
	}
	clientConn, err := r.dialer.Dial(address)
	if err != nil {
		return nil, err
	}
	return pfs.NewApiClient(clientConn), nil
}

func (r *router) GetMasterOrSlaveAPIClient(shard int) (pfs.ApiClient, error) {
	addresses, err := r.addresser.GetSlaveAddresses(shard)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return r.GetMasterAPIClient(shard)
	}
	address := addresses[int(rand.Uint32())%len(addresses)]
	clientConn, err := r.dialer.Dial(address)
	if err != nil {
		return nil, err
	}
	return pfs.NewApiClient(clientConn), nil
}
