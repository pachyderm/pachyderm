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

func (a *discoveryAddresser) WatchShardToAddress(cancel chan bool, callBack func(map[int]string, map[int]map[string]bool) (uint64, error)) error {
	return a.discoveryClient.WatchAll(
		a.shardDir(),
		cancel,
		func(addresses map[string]string) (uint64, error) {
			shardToMasterAddress, shardToReplicaAddress, err := a.makeShardMaps(addresses)
			if err != nil {
				return 0, err
			}
			return callBack(shardToMasterAddress, shardToReplicaAddress)
		},
	)
}

func (a *discoveryAddresser) GetShardToReplicaAddresses() (map[int]map[string]bool, error) {
	base := a.replicaDir()
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	return a.makeReplicaMap(addresses)
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
	if prevAddress != "" {
		// NOTE ignoring the index CheckAndDelete returns is ok because the
		// index returned by SetReplicaAddress below is guaranteed to be higher
		// since it occurs after.
		if _, err := a.discoveryClient.CheckAndDelete(a.replicaKey(shard), prevAddress); err != nil {
			return 0, err
		}
	}
	return a.SetReplicaAddress(shard, address)
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

func (a *discoveryAddresser) shardDir() string {
	return fmt.Sprintf("%s/pfs/shard", a.namespace)
}

func (a *discoveryAddresser) masterDir() string {
	return path.Join(a.shardDir(), "master")
}

func (a *discoveryAddresser) masterKey(shard int) string {
	return path.Join(a.masterDir(), fmt.Sprint(shard))
}

func (a *discoveryAddresser) replicaDir() string {
	return path.Join(a.shardDir(), "replica")
}

func (a *discoveryAddresser) replicaKey(shard int) string {
	return path.Join(a.replicaDir(), fmt.Sprint(shard))
}

func (a *discoveryAddresser) makeShardMaps(addresses map[string]string) (map[int]string, map[int]map[string]bool, error) {
	masterMap := make(map[int]string)
	replicaMap := make(map[int]map[string]bool)
	masterPrefix := fmt.Sprintf("%s/", a.masterDir())
	replicaPrefix := fmt.Sprintf("%s/", a.replicaDir())
	for shardString, address := range addresses {
		if strings.HasPrefix(shardString, masterPrefix) {
			shard, err := strconv.ParseInt(strings.TrimPrefix(shardString, masterPrefix), 10, 64)
			if err != nil {
				return nil, nil, err
			}
			masterMap[int(shard)] = address
		}
		if strings.HasPrefix(shardString, replicaPrefix) {
			shardString = strings.TrimPrefix(shardString, replicaPrefix)
			shardString = strings.Split(shardString, "/")[0]
			shard, err := strconv.ParseInt(shardString, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			if _, ok := replicaMap[int(shard)]; !ok {
				replicaMap[int(shard)] = make(map[string]bool, 0)
			}
			replicaMap[int(shard)][address] = true
		}
	}
	return masterMap, replicaMap, nil
}

func (a *discoveryAddresser) makeMasterMap(addresses map[string]string) (map[int]string, error) {
	result, _, err := a.makeShardMaps(addresses)
	return result, err
}

func (a *discoveryAddresser) makeReplicaMap(addresses map[string]string) (map[int]map[string]bool, error) {
	_, result, err := a.makeShardMaps(addresses)
	return result, err
}
