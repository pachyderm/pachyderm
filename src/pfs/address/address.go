package address

import "github.com/pachyderm/pachyderm/src/pfs/discovery"

type Addresser interface {
	GetMasterAddress(shard int) (string, error)
	GetSlaveAddresses(shard int) ([]string, error)
	GetAllAddresses() ([]string, error)
}

func NewSingleAddresser(address string) Addresser {
	return newSingleAddresser(address)
}

func NewDiscoveryAddresser(discoveryClient discovery.Client) Addresser {
	return nil
}
