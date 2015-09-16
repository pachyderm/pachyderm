package route

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/libkv/store"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
)

var (
	holdTTL      uint64              = 20
	writeOptions *store.WriteOptions = &store.WriteOptions{TTL: time.Second * 20}
	marshaler                        = &jsonpb.Marshaler{}
)

type discoveryAddresser struct {
	discoveryClient discovery.Client
	namespace       string
	store           store.Store
}

func newDiscoveryAddresser(discoveryClient discovery.Client, store store.Store, namespace string) *discoveryAddresser {
	return &discoveryAddresser{discoveryClient, namespace, store}
}

func (a *discoveryAddresser) GetMasterAddress(shard int) (Address, bool, error) {
	address, ok, err := a.discoveryClient.Get(a.masterKey(shard))
	return decodeAddress(address), ok, err
}

func (a *discoveryAddresser) GetReplicaAddresses(shard int) (map[Address]bool, error) {
	base := a.replicaShardDir(shard)
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	m := make(map[Address]bool, 0)
	for _, address := range addresses {
		m[decodeAddress(address)] = true
	}
	return m, nil
}

func (a *discoveryAddresser) GetShardToMasterAddress() (map[int]Address, error) {
	addresses, err := a.discoveryClient.GetAll(a.masterDir())
	if err != nil {
		return nil, err
	}
	return a.makeMasterMap(addresses)
}

func (a *discoveryAddresser) WatchShardToAddress(cancel chan bool, callBack func(map[int]Address, map[int]map[int]Address) (uint64, error)) error {
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

func (a *discoveryAddresser) GetShardToReplicaAddresses() (map[int]map[int]Address, error) {
	base := a.replicaDir()
	addresses, err := a.discoveryClient.GetAll(base)
	if err != nil {
		return nil, err
	}
	return a.makeReplicaMap(addresses)
}

func (a *discoveryAddresser) SetMasterAddress(shard int, address Address) (uint64, error) {
	return a.discoveryClient.Set(a.masterKey(shard), encodeAddress(address), 0)
}

func (a *discoveryAddresser) ClaimMasterAddress(shard int, address Address, prevAddress Address) (uint64, error) {
	return a.discoveryClient.CheckAndSet(a.masterKey(shard), encodeAddress(address), holdTTL, encodeAddress(prevAddress))
}

func (a *discoveryAddresser) HoldMasterAddress(shard int, address Address, cancel chan bool) error {
	return a.discoveryClient.Hold(a.masterKey(shard), encodeAddress(address), holdTTL, cancel)
}

func (a *discoveryAddresser) SetReplicaAddress(shard int, index int, address Address) (uint64, error) {
	return a.discoveryClient.CreateInDir(a.replicaKey(shard, index), encodeAddress(address), holdTTL)
}

func (a *discoveryAddresser) ClaimReplicaAddress(shard int, index int, address Address, prevAddress Address) (uint64, error) {
	return a.discoveryClient.CheckAndSet(a.replicaKey(shard, index), encodeAddress(address), holdTTL, encodeAddress(prevAddress))
}

func (a *discoveryAddresser) HoldReplicaAddress(shard int, index int, address Address, cancel chan bool) error {
	return a.discoveryClient.Hold(a.replicaKey(shard, index), encodeAddress(address), holdTTL, cancel)
}

func (a *discoveryAddresser) DeleteMasterAddress(shard int) (uint64, error) {
	return a.discoveryClient.Delete(a.masterKey(shard))
}

func (a *discoveryAddresser) DeleteReplicaAddress(shard int, index int, address Address) (uint64, error) {
	return a.discoveryClient.Delete(a.replicaKey(shard, index))
}

func decodeAddress(rawAddress string) Address {
	return Address{
		strings.TrimPrefix(rawAddress, "-"),
		strings.HasPrefix(rawAddress, "-"),
	}
}

func encodeAddress(address Address) string {
	if address.Backfilling {
		return "-" + address.Address
	}
	return address.Address
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

func (a *discoveryAddresser) replicaShardDir(shard int) string {
	return path.Join(a.replicaDir(), fmt.Sprint(shard))
}

func (a *discoveryAddresser) replicaKey(shard int, index int) string {
	return path.Join(a.replicaShardDir(shard), fmt.Sprint(index))
}

func (a *discoveryAddresser) serverDir() string {
	return fmt.Sprintf("%s/pfs/server", a.namespace)
}

func (a *discoveryAddresser) serverKey(id string) string {
	return path.Join(a.serverDir(), id)
}

func (a *discoveryAddresser) serverAliveKey(id string) string {
	return path.Join(a.serverDir(), "alive", id)
}

func (a *discoveryAddresser) serverStateKey(id string) string {
	return path.Join(a.serverDir(), "state", id)
}

func (a *discoveryAddresser) serverRoleKey(id string) string {
	return path.Join(a.serverDir(), "role", id)
}

func (a *discoveryAddresser) addressDir() string {
	return fmt.Sprintf("%s/pfs/address", a.namespace)
}

func (a *discoveryAddresser) makeShardMaps(addresses map[string]string) (map[int]Address, map[int]map[int]Address, error) {
	masterMap := make(map[int]Address)
	replicaMap := make(map[int]map[int]Address)
	masterPrefix := fmt.Sprintf("%s/", a.masterDir())
	replicaPrefix := fmt.Sprintf("%s/", a.replicaDir())
	for shardString, address := range addresses {
		if strings.HasPrefix(shardString, masterPrefix) {
			shard, err := strconv.ParseInt(strings.TrimPrefix(shardString, masterPrefix), 10, 64)
			if err != nil {
				return nil, nil, err
			}
			masterMap[int(shard)] = decodeAddress(address)
		}
		if strings.HasPrefix(shardString, replicaPrefix) {
			shardString = strings.TrimPrefix(shardString, replicaPrefix)
			shardAndIndex := strings.Split(shardString, "/")
			shard, err := strconv.ParseInt(shardAndIndex[0], 10, 64)
			if err != nil {
				return nil, nil, err
			}
			index, err := strconv.ParseInt(shardAndIndex[1], 10, 64)
			if err != nil {
				return nil, nil, err
			}
			if _, ok := replicaMap[int(shard)]; !ok {
				replicaMap[int(shard)] = make(map[int]Address, 0)
			}
			replicaMap[int(shard)][int(index)] = decodeAddress(address)
		}
	}
	return masterMap, replicaMap, nil
}

func (a *discoveryAddresser) makeMasterMap(addresses map[string]string) (map[int]Address, error) {
	result, _, err := a.makeShardMaps(addresses)
	return result, err
}

func (a *discoveryAddresser) makeReplicaMap(addresses map[string]string) (map[int]map[int]Address, error) {
	_, result, err := a.makeShardMaps(addresses)
	return result, err
}

func (a *discoveryAddresser) Announce(cancel chan bool, id string, address string, server Server) (retErr error) {
	var once sync.Once
	internalCancel := make(chan struct{})
	versionChan := make(chan int64)
	go func() {
		serverState := &ServerState{
			Id:      id,
			Address: address,
			Version: -1,
		}
		for {
			var buffer bytes.Buffer
			if err := marshaler.Marshal(&buffer, serverState); err != nil {
				once.Do(func() {
					retErr = err
					close(internalCancel)
				})
				return
			}
			if err := a.store.Put(a.serverStateKey(id), buffer.Bytes(), writeOptions); err != nil {
				once.Do(func() {
					retErr = err
					close(internalCancel)
				})
				return
			}
			select {
			case <-internalCancel:
				return
			case version := <-versionChan:
				serverState.Version = version
			case <-time.After(writeOptions.TTL / 2):
			}
		}
	}()
	go func() {
		kvPairs, err := a.store.Watch(a.serverRoleKey(id), internalCancel)
		if err != nil {
			once.Do(func() {
				retErr = err
				close(internalCancel)
			})
			return
		}
		for kvPair := range kvPairs {
			var serverRole ServerRole
			if err := jsonpb.Unmarshal(bytes.NewBuffer(kvPair.Value), &serverRole); err != nil {
				once.Do(func() {
					retErr = err
					close(internalCancel)
				})
				return
			}
			if err := a.runRole(serverRole, server); err != nil {
				once.Do(func() {
					retErr = err
					close(internalCancel)
				})
				return
			}
		}
	}()
	<-cancel
	once.Do(func() {
		retErr = fmt.Errorf("announce cancelled")
		close(internalCancel)
	})
	return
}

func (a *discoveryAddresser) Version() (string, error) {
}
