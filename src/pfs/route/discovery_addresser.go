package route

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
)

var (
	holdTTL uint64 = 20
)

type discoveryAddresser struct {
	discoveryClient discovery.Client
	namespace       string
}

func newDiscoveryAddresser(discoveryClient discovery.Client, namespace string) *discoveryAddresser {
	return &discoveryAddresser{discoveryClient, namespace}
}

func (a *discoveryAddresser) GetMasterAddress(shard int) (string, bool, error) {
	return a.discoveryClient.Get(a.masterKey(shard))
}

func (a *discoveryAddresser) GetReplicaAddresses(shard int) (map[string]bool, error) {
	base := a.replicaKey(shard)
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, 0)
	for _, address := range addresses {
		m[address] = true
	}
	return m, nil
}

func (a *discoveryAddresser) GetShardToMasterAddress() (map[int]string, error) {
	addresses, err := a.discoveryClient.GetAll(a.masterDir())
	if err != nil {
		return nil, err
	}
	return a.makeMasterMap(addresses)
}

func (a *discoveryAddresser) WatchShardToMasterAddress(cancel chan bool, callBack func(map[int]string) (uint64, error)) error {
	return a.discoveryClient.WatchAll(
		a.masterDir(),
		cancel,
		func(addresses map[string]string) (uint64, error) {
			shardToMasterAddress, err := a.makeMasterMap(addresses)
			if err != nil {
				return 0, err
			}
			return callBack(shardToMasterAddress)
		},
	)
}

func (a *discoveryAddresser) GetShardToReplicaAddresses() (map[int]map[string]bool, error) {
	base := a.replicaDir()
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[int]map[string]bool, 0)
	for shardString, address := range addresses {
		shardString = strings.TrimPrefix(shardString, fmt.Sprintf("%s/", base))
		shardString = strings.Split(shardString, "/")[0]
		shard, err := strconv.ParseInt(shardString, 10, 64)
		if err != nil {
			return nil, err
		}
		if _, ok := m[int(shard)]; !ok {
			m[int(shard)] = make(map[string]bool, 0)
		}
		m[int(shard)][address] = true
	}
	return m, nil
}

func (a *discoveryAddresser) SetMasterAddress(shard int, address string) (uint64, error) {
	return a.discoveryClient.Set(a.masterKey(shard), address, 0)
}

func (a *discoveryAddresser) ClaimMasterAddress(shard int, address string, prevAddress string) (uint64, error) {
	return a.discoveryClient.CheckAndSet(a.masterKey(shard), address, holdTTL, prevAddress)
}

func (a *discoveryAddresser) HoldMasterAddress(shard int, address string, cancel chan bool) error {
	return a.discoveryClient.Hold(a.masterKey(shard), address, holdTTL, cancel)
}

func (a *discoveryAddresser) SetReplicaAddress(shard int, address string) (uint64, error) {
	return a.discoveryClient.CreateInDir(a.replicaKey(shard), address, holdTTL)
}

func (a *discoveryAddresser) ClaimReplicaAddress(shard int, address string, prevAddress string) (uint64, error) {
	return a.discoveryClient.CheckAndSet(a.replicaKey(shard), address, holdTTL, prevAddress)
}

func (a *discoveryAddresser) HoldReplicaAddress(shard int, address string, cancel chan bool) error {
	return a.discoveryClient.Hold(a.replicaKey(shard), address, holdTTL, cancel)
}

func (a *discoveryAddresser) DeleteMasterAddress(shard int) (uint64, error) {
	return a.discoveryClient.Delete(a.masterKey(shard))
}

func (a *discoveryAddresser) DeleteReplicaAddress(shard int, address string) (uint64, error) {
	return a.discoveryClient.Delete(a.replicaKey(shard))
}

func (a *discoveryAddresser) masterDir() string {
	return fmt.Sprintf("%s/pfs/shard/master", a.namespace)
}

func (a *discoveryAddresser) masterKey(shard int) string {
	return path.Join(a.masterDir(), fmt.Sprint(shard))
}

func (a *discoveryAddresser) replicaDir() string {
	return fmt.Sprintf("%s/pfs/shard/replica", a.namespace)
}

func (a *discoveryAddresser) replicaKey(shard int) string {
	return path.Join(a.replicaDir(), fmt.Sprint(shard))
}

func (a *discoveryAddresser) makeMasterMap(addresses map[string]string) (map[int]string, error) {
	result := make(map[int]string, 0)
	for shardString, address := range addresses {
		shard, err := strconv.ParseInt(strings.TrimPrefix(shardString, fmt.Sprintf("%s/", a.masterDir())), 10, 64)
		if err != nil {
			return nil, err
		}
		result[int(shard)] = address
	}
	return result, nil
}
