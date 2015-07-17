package address

import "github.com/pachyderm/pachyderm/src/pfs/discovery"

type Addresser interface {
	GetMasterShards(address string) (map[int]bool, error)
	GetSlaveShards(address string) (map[int]bool, error)
	GetAllAddresses() ([]string, error)
}

func NewSingleAddresser(address string, numShards int) Addresser {
	return newSingleAddresser(address, numShards)
}

func NewDiscoveryAddresser(discoveryClient discovery.Client) Addresser {
	return nil
}
