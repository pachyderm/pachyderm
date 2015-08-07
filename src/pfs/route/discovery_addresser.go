package route

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
)

type discoveryAddresser struct {
	discoveryClient discovery.Client
	namespace       string
}

func newDiscoveryAddresser(discoveryClient discovery.Client, namespace string) *discoveryAddresser {
	return &discoveryAddresser{discoveryClient, namespace}
}

func (a *discoveryAddresser) GetMasterAddress(shard int) (string, bool, error) {
	return a.discoveryClient.Get(fmt.Sprintf("%s/pfs/shard/master/%d", a.namespace, shard))
}

func (a *discoveryAddresser) GetSlaveAddresses(shard int) (map[string]bool, error) {
	base := fmt.Sprintf("%s/pfs/shard/slave/%d", a.namespace, shard)
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, 0)
	for address := range addresses {
		m[strings.TrimPrefix(address, fmt.Sprintf("%s/", base))] = true
	}
	return m, nil
}

func (a *discoveryAddresser) GetShardToMasterAddress() (map[int]string, error) {
	base := fmt.Sprintf("%s/pfs/shard/master", a.namespace)
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, 0)
	for shardString, address := range addresses {
		shard, err := strconv.ParseInt(strings.TrimPrefix(shardString, fmt.Sprintf("%s/", base)), 10, 64)
		if err != nil {
			return nil, err
		}
		m[int(shard)] = address
	}
	return m, nil
}

func (a *discoveryAddresser) GetShardToSlaveAddresses() (map[int]map[string]bool, error) {
	base := fmt.Sprintf("%s/pfs/shard/slave", a.namespace)
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[int]map[string]bool, 0)
	for shardAddress := range addresses {
		shardAddress := strings.TrimPrefix(shardAddress, fmt.Sprintf("%s/", base))
		split := strings.Split(shardAddress, "/")
		if len(split) != 2 {
			return nil, fmt.Errorf("cannot split %s into shard and address", shardAddress)
		}
		shard, err := strconv.ParseInt(strings.TrimPrefix(split[0], fmt.Sprintf("%s/", base)), 10, 64)
		if err != nil {
			return nil, err
		}
		if _, ok := m[int(shard)]; !ok {
			m[int(shard)] = make(map[string]bool, 0)
		}
		m[int(shard)][split[1]] = true
	}
	return m, nil
}

func (a *discoveryAddresser) SetMasterAddress(shard int, address string, ttl uint64) error {
	return a.discoveryClient.Set(fmt.Sprintf("%s/pfs/shard/master/%d", a.namespace, shard), address, ttl)
}

func (a *discoveryAddresser) SetSlaveAddress(shard int, address string, ttl uint64) error {
	return a.discoveryClient.Set(fmt.Sprintf("%s/pfs/shard/slave/%d/%s", a.namespace, shard, address), "1", ttl)
}

func (a *discoveryAddresser) DeleteMasterAddress(shard int) error {
	return a.discoveryClient.Delete(fmt.Sprintf("%s/pfs/shard/master/%d", a.namespace, shard))
}

func (a *discoveryAddresser) DeleteSlaveAddress(shard int, address string) error {
	return a.discoveryClient.Delete(fmt.Sprintf("%s/pfs/shard/slave/%d/%s", a.namespace, shard, address))
}
