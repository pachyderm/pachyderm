package route

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
)

const (
	holdTTL uint64 = 60
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
	base := a.masterDir()
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

func (a *discoveryAddresser) SetMasterAddress(shard int, address string, ttl uint64) error {
	return a.discoveryClient.Set(a.masterKey(shard), address, ttl)
}

func (a *discoveryAddresser) HoldMasterAddress(shard int, address string, prevAddress string) error {
	if prevAddress == "" {
		if err := a.discoveryClient.Create(a.masterKey(shard), address, holdTTL); err != nil {
			return err
		}
	} else {
		if err := a.discoveryClient.CheckAndSet(a.masterKey(shard), address, holdTTL, prevAddress); err != nil {
			return err
		}
	}
	for {
		// TODO we could make this function more responsive by watching for updates
		time.Sleep(time.Second * time.Duration(holdTTL/2))
		if err := a.discoveryClient.CheckAndSet(a.masterKey(shard), address, holdTTL, address); err != nil {
			return err
		}
	}
}

func (a *discoveryAddresser) SetReplicaAddress(shard int, address string, ttl uint64) error {
	return a.discoveryClient.CreateInDir(a.replicaKey(shard), address, ttl)
}

func (a *discoveryAddresser) HoldReplicaAddress(shard int, address string) error {
	if err := a.SetReplicaAddress(shard, address, holdTTL); err != nil {
		return err
	}
	for {
		// TODO we could make this function more responsive by watching for updates
		time.Sleep(time.Second * time.Duration(holdTTL/2))
		if err := a.discoveryClient.CheckAndSet(a.replicaKey(shard), address, holdTTL, address); err != nil {
			return err
		}
	}
}

func (a *discoveryAddresser) DeleteMasterAddress(shard int) error {
	return a.discoveryClient.Delete(a.masterKey(shard))
}

func (a *discoveryAddresser) DeleteReplicaAddress(shard int, address string) error {
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
