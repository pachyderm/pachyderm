package route

import (
	"fmt"
	"math"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
)

var (
	holdTTL      uint64 = 20
	marshaler           = &jsonpb.Marshaler{}
	ErrCancelled        = fmt.Errorf("cancelled by user")
)

type discoveryAddresser struct {
	discoveryClient discovery.Client
	sharder         Sharder
	namespace       string
	addresses       map[int64]*Addresses
}

func newDiscoveryAddresser(discoveryClient discovery.Client, sharder Sharder, namespace string) *discoveryAddresser {
	return &discoveryAddresser{discoveryClient, sharder, namespace, make(map[int64]*Addresses)}
}

func (a *discoveryAddresser) GetMasterAddress(shard uint64, version int64) (string, bool, error) {
	addresses, err := a.getAddresses(version)
	if err != nil {
		return "", false, err
	}
	shardAddresses, ok := addresses.Addresses[shard]
	if !ok {
		return "", false, nil
	}
	return shardAddresses.Master, true, nil
}

func (a *discoveryAddresser) GetReplicaAddresses(shard uint64, version int64) (map[string]bool, error) {
	addresses, err := a.getAddresses(version)
	if err != nil {
		return nil, err
	}
	shardAddresses, ok := addresses.Addresses[shard]
	if !ok {
		return nil, fmt.Errorf("shard %d not found", shard)
	}
	return shardAddresses.Replicas, nil
}

func (a *discoveryAddresser) GetShardToMasterAddress(version int64) (map[uint64]string, error) {
	addresses, err := a.getAddresses(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]string)
	for shard, shardAddresses := range addresses.Addresses {
		result[shard] = shardAddresses.Master
	}
	return result, nil
}

func (a *discoveryAddresser) GetShardToReplicaAddresses(version int64) (map[uint64]map[string]bool, error) {
	addresses, err := a.getAddresses(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]map[string]bool)
	for shard, shardAddresses := range addresses.Addresses {
		result[shard] = shardAddresses.Replicas
	}
	return result, nil
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

func (a *discoveryAddresser) serverStateDir() string {
	return path.Join(a.serverDir(), "state")
}

func (a *discoveryAddresser) serverStateKey(id string) string {
	return path.Join(a.serverStateDir(), id)
}

func (a *discoveryAddresser) serverRoleDir() string {
	return path.Join(a.serverDir(), "role")
}

func (a *discoveryAddresser) serverRoleKey(id string) string {
	return path.Join(a.serverRoleDir(), id)
}

func (a *discoveryAddresser) serverRoleKeyVersion(id string, version int64) string {
	return path.Join(a.serverRoleKey(id), fmt.Sprint(version))
}

func (a *discoveryAddresser) addressesDir() string {
	return fmt.Sprintf("%s/pfs/roles", a.namespace)
}

func (a *discoveryAddresser) addressesKey(version int64) string {
	return path.Join(a.addressesDir(), fmt.Sprint(version))
}

func (a *discoveryAddresser) Register(cancel chan bool, id string, address string, server Server) (retErr error) {
	var once sync.Once
	versionChan := make(chan int64)
	internalCancel := make(chan bool)
	go func() {
		if err := a.announceState(id, address, server, versionChan, internalCancel); err != nil {
			once.Do(func() {
				retErr = err
				close(internalCancel)
			})
		}
	}()
	go func() {
		if err := a.fillRoles(id, server, versionChan, internalCancel); err != nil {
			once.Do(func() {
				retErr = err
				close(internalCancel)
			})
		}
	}()
	<-cancel
	once.Do(func() {
		retErr = ErrCancelled
		close(internalCancel)
	})
	return
}

func (a *discoveryAddresser) AssignRoles(cancel chan bool) error {
	var version int64
	oldServerStates := make(map[string]ServerState)
	oldRoles := make(map[string]ServerRole)
	oldMasters := make(map[uint64]string)
	oldReplicas := make(map[uint64][]string)
	var oldMinVersion int64
	err := a.discoveryClient.WatchAll(a.serverStateDir(), cancel,
		func(encodedServerStates map[string]string) (uint64, error) {
			if len(encodedServerStates) == 0 {
				return 0, nil
			}
			newServerStates := make(map[string]ServerState)
			shardLocations := make(map[uint64][]string)
			newRoles := make(map[string]ServerRole)
			newMasters := make(map[uint64]string)
			newReplicas := make(map[uint64][]string)
			masterRolesPerServer := a.sharder.NumShards() / uint64(len(encodedServerStates))
			masterRolesRemainder := a.sharder.NumShards() % uint64(len(encodedServerStates))
			replicaRolesPerServer := (a.sharder.NumShards() * (a.sharder.NumReplicas())) / uint64(len(encodedServerStates))
			replicaRolesRemainder := (a.sharder.NumShards() * (a.sharder.NumReplicas())) % uint64(len(encodedServerStates))
			for _, encodedServerState := range encodedServerStates {
				var serverState ServerState
				if err := jsonpb.UnmarshalString(encodedServerState, &serverState); err != nil {
					return 0, err
				}
				newServerStates[serverState.Id] = serverState
				newRoles[serverState.Id] = ServerRole{
					Version:  version,
					Masters:  make(map[uint64]bool),
					Replicas: make(map[uint64]bool),
				}
				for shard := range serverState.Shards {
					shardLocations[shard] = append(shardLocations[shard], serverState.Id)
				}
			}
			// See if there's any roles we can delete
			minVersion := int64(math.MaxInt64)
			for _, serverState := range newServerStates {
				if serverState.Version < minVersion {
					minVersion = serverState.Version
				}
			}
			// Delete roles that no servers are using anymore
			if minVersion > oldMinVersion {
				oldMinVersion = minVersion
				serverRoles, err := a.discoveryClient.GetAll(a.serverRoleDir())
				if err != nil {
					return 0, err
				}
				for key, encodedServerRole := range serverRoles {
					var serverRole ServerRole
					if err := jsonpb.UnmarshalString(encodedServerRole, &serverRole); err != nil {
						return 0, err
					}
					if serverRole.Version < minVersion {
						if _, err := a.discoveryClient.Delete(key); err != nil {
							return 0, err
						}
					}
				}
			}
			// if the servers are identical to last time then we know we'll
			// assign shards the same way
			if sameServers(oldServerStates, newServerStates) {
				return 0, nil
			}
		Master:
			for shard := uint64(0); shard < a.sharder.NumShards(); shard++ {
				if id, ok := oldMasters[shard]; ok {
					if assignMaster(newRoles, newMasters, id, shard, masterRolesPerServer, &masterRolesRemainder) {
						continue Master
					}
				}
				for _, id := range oldReplicas[shard] {
					if assignMaster(newRoles, newMasters, id, shard, masterRolesPerServer, &masterRolesRemainder) {
						continue Master
					}
				}
				for _, id := range shardLocations[shard] {
					if assignMaster(newRoles, newMasters, id, shard, masterRolesPerServer, &masterRolesRemainder) {
						continue Master
					}
				}
				for id := range newServerStates {
					if assignMaster(newRoles, newMasters, id, shard, masterRolesPerServer, &masterRolesRemainder) {
						continue Master
					}
				}
				return 0, nil
			}
			for replica := uint64(0); replica < a.sharder.NumReplicas(); replica++ {
			Replica:
				for shard := uint64(0); shard < a.sharder.NumShards(); shard++ {
					if id, ok := oldMasters[shard]; ok {
						if assignReplica(newRoles, newMasters, newReplicas, id, shard, replicaRolesPerServer, &replicaRolesRemainder) {
							continue Replica
						}
					}
					for _, id := range oldReplicas[shard] {
						if assignReplica(newRoles, newMasters, newReplicas, id, shard, replicaRolesPerServer, &replicaRolesRemainder) {
							continue Replica
						}
					}
					for _, id := range shardLocations[shard] {
						if assignReplica(newRoles, newMasters, newReplicas, id, shard, replicaRolesPerServer, &replicaRolesRemainder) {
							continue Replica
						}
					}
					for id := range newServerStates {
						if assignReplica(newRoles, newMasters, newReplicas, id, shard, replicaRolesPerServer, &replicaRolesRemainder) {
							continue Replica
						}
					}
					for id := range newServerStates {
						if swapReplica(newRoles, newMasters, newReplicas, id, shard, replicaRolesPerServer) {
							continue Replica
						}
					}
					return 0, nil
				}
			}
			addresses := Addresses{
				Version:   version,
				Addresses: make(map[uint64]*ShardAddresses),
			}
			for shard := uint64(0); shard < a.sharder.NumShards(); shard++ {
				addresses.Addresses[shard] = &ShardAddresses{Replicas: make(map[string]bool)}
			}
			for id, serverRole := range newRoles {
				encodedServerRole, err := marshaler.MarshalToString(&serverRole)
				if err != nil {
					return 0, err
				}
				if _, err := a.discoveryClient.Set(a.serverRoleKeyVersion(id, version), encodedServerRole, 0); err != nil {
					return 0, err
				}
				address := newServerStates[id].Address
				for shard := range serverRole.Masters {
					shardAddresses := addresses.Addresses[shard]
					shardAddresses.Master = address
					addresses.Addresses[shard] = shardAddresses
				}
				for shard := range serverRole.Replicas {
					shardAddresses := addresses.Addresses[shard]
					shardAddresses.Replicas[address] = true
					addresses.Addresses[shard] = shardAddresses
				}
			}
			encodedAddresses, err := marshaler.MarshalToString(&addresses)
			if err != nil {
				return 0, err
			}
			if _, err := a.discoveryClient.Set(a.addressesKey(version), encodedAddresses, 0); err != nil {
				return 0, err
			}
			version++
			oldServerStates = newServerStates
			oldRoles = newRoles
			oldMasters = newMasters
			oldReplicas = newReplicas
			return 0, nil
		})
	if err == discovery.ErrCancelled {
		return ErrCancelled
	}
	return err
}

func (a *discoveryAddresser) Version() (int64, error) {
	minVersion := int64(math.MaxInt64)
	encodedServerStates, err := a.discoveryClient.GetAll(a.serverStateDir())
	if err != nil {
		return 0, err
	}
	for _, encodedServerState := range encodedServerStates {
		var serverState ServerState
		if err := jsonpb.UnmarshalString(encodedServerState, &serverState); err != nil {
			return 0, err
		}
		if serverState.Version < minVersion {
			minVersion = serverState.Version
		}
	}
	return minVersion, nil
}

func (a *discoveryAddresser) WaitOneVersion() error {
	errComplete := fmt.Errorf("COMPLETE")
	err := a.discoveryClient.WatchAll(a.serverDir(), nil,
		func(encodedServerStatesAndRoles map[string]string) (uint64, error) {
			var serverStates []ServerState
			var serverRoles []ServerRole
			for key, encodedServerStateOrRole := range encodedServerStatesAndRoles {
				if strings.HasPrefix(key, a.serverStateDir()) {
					var serverState ServerState
					if err := jsonpb.UnmarshalString(encodedServerStateOrRole, &serverState); err != nil {
						return 0, err
					}
					serverStates = append(serverStates, serverState)
				}
				if strings.HasPrefix(key, a.serverRoleDir()) {
					var serverRole ServerRole
					if err := jsonpb.UnmarshalString(encodedServerStateOrRole, &serverRole); err != nil {
						return 0, err
					}
					serverRoles = append(serverRoles, serverRole)
				}
			}
			versions := make(map[int64]bool)
			for _, serverState := range serverStates {
				if serverState.Version == -1 {
					return 0, nil
				}
				versions[serverState.Version] = true
			}
			if len(versions) != 1 {
				return 0, nil
			}
			for _, serverRole := range serverRoles {
				if !versions[serverRole.Version] {
					return 0, nil
				}
			}
			return 0, errComplete
		})
	if err != errComplete {
		return err
	}
	return nil
}

func (a *discoveryAddresser) getAddresses(version int64) (*Addresses, error) {
	if addresses, ok := a.addresses[version]; ok {
		return addresses, nil
	}
	encodedAddresses, ok, err := a.discoveryClient.Get(a.addressesKey(version))
	if err != nil {
		return nil, err
	}
	if !ok {
		fmt.Errorf("version %d not found", version)
	}
	var addresses Addresses
	if err := jsonpb.UnmarshalString(encodedAddresses, &addresses); err != nil {
		return nil, err
	}
	return &addresses, nil
}

func hasShard(serverRole ServerRole, shard uint64) bool {
	return serverRole.Masters[shard] || serverRole.Replicas[shard]
}

func removeReplica(replicas map[uint64][]string, shard uint64, id string) {
	var ids []string
	for _, replicaID := range replicas[shard] {
		if id != replicaID {
			ids = append(ids, replicaID)
		}
	}
	replicas[shard] = ids
}

func assignMaster(
	serverRoles map[string]ServerRole,
	masters map[uint64]string,
	id string,
	shard uint64,
	masterRolesPerServer uint64,
	masterRolesRemainder *uint64,
) bool {
	serverRole, ok := serverRoles[id]
	if !ok {
		return false
	}
	if uint64(len(serverRole.Masters)) > masterRolesPerServer {
		return false
	}
	if uint64(len(serverRole.Masters)) == masterRolesPerServer && *masterRolesRemainder == 0 {
		return false
	}
	if hasShard(serverRole, shard) {
		return false
	}
	if uint64(len(serverRole.Masters)) == masterRolesPerServer && *masterRolesRemainder > 0 {
		*masterRolesRemainder--
	}
	serverRole.Masters[shard] = true
	serverRoles[id] = serverRole
	masters[shard] = id
	return true
}

func assignReplica(
	serverRoles map[string]ServerRole,
	masters map[uint64]string,
	replicas map[uint64][]string,
	id string,
	shard uint64,
	replicaRolesPerServer uint64,
	replicaRolesRemainder *uint64,
) bool {
	serverRole, ok := serverRoles[id]
	if !ok {
		return false
	}
	if uint64(len(serverRole.Replicas)) > replicaRolesPerServer {
		return false
	}
	if uint64(len(serverRole.Replicas)) == replicaRolesPerServer && *replicaRolesRemainder == 0 {
		return false
	}
	if hasShard(serverRole, shard) {
		return false
	}
	if uint64(len(serverRole.Replicas)) == replicaRolesPerServer && *replicaRolesRemainder > 0 {
		*replicaRolesRemainder--
	}
	serverRole.Replicas[shard] = true
	serverRoles[id] = serverRole
	replicas[shard] = append(replicas[shard], id)
	return true
}

func swapReplica(
	serverRoles map[string]ServerRole,
	masters map[uint64]string,
	replicas map[uint64][]string,
	id string,
	shard uint64,
	replicaRolesPerServer uint64,
) bool {
	serverRole, ok := serverRoles[id]
	if !ok {
		return false
	}
	if uint64(len(serverRole.Replicas)) >= replicaRolesPerServer {
		return false
	}
	for swapID, swapServerRole := range serverRoles {
		if swapID == id {
			continue
		}
		for swapShard := range swapServerRole.Replicas {
			if hasShard(serverRole, swapShard) {
				continue
			}
			if hasShard(swapServerRole, shard) {
				continue
			}
			delete(swapServerRole.Replicas, swapShard)
			serverRoles[swapID] = swapServerRole
			removeReplica(replicas, swapShard, swapID)
			// We do some weird things with the limits here, both servers
			// receive a 0 replicaRolesRemainder, swapID doesn't need a
			// remainder because we're replacing a shard we stole so it also
			// has MaxInt64 for replicaRolesPerServer. We already know id
			// doesn't need the remainder since we check that it has fewer than
			// replicaRolesPerServer replicas.
			var noReplicaRemainder uint64
			assignReplica(serverRoles, masters, replicas, swapID, shard, math.MaxUint64, &noReplicaRemainder)
			assignReplica(serverRoles, masters, replicas, id, swapShard, replicaRolesPerServer, &noReplicaRemainder)
			return true
		}
	}
	return false
}

func (a *discoveryAddresser) announceState(
	id string,
	address string,
	server Server,
	versionChan chan int64,
	cancel chan bool,
) error {
	serverState := &ServerState{
		Id:      id,
		Address: address,
		Version: -1,
	}
	for {
		shards, err := server.LocalShards()
		if err != nil {
			return err
		}
		serverState.Shards = shards
		encodedServerState, err := marshaler.MarshalToString(serverState)
		if err != nil {
			return err
		}
		if _, err := a.discoveryClient.Set(a.serverStateKey(id), encodedServerState, holdTTL); err != nil {
			return err
		}
		select {
		case <-cancel:
			return nil
		case version := <-versionChan:
			serverState.Version = version
		case <-time.After(time.Second * time.Duration(holdTTL/2)):
		}
	}
}

type int64Slice []int64

func (s int64Slice) Len() int           { return len(s) }
func (s int64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s int64Slice) Less(i, j int) bool { return s[i] < s[j] }

func (a *discoveryAddresser) fillRoles(
	id string,
	server Server,
	versionChan chan int64,
	cancel chan bool,
) error {
	oldRoles := make(map[int64]ServerRole)
	return a.discoveryClient.WatchAll(
		a.serverRoleKey(id),
		cancel,
		func(encodedServerRoles map[string]string) (uint64, error) {
			roles := make(map[int64]ServerRole, len(encodedServerRoles))
			var versions int64Slice
			// Decode the roles
			for _, encodedServerRole := range encodedServerRoles {
				var serverRole ServerRole
				if err := jsonpb.UnmarshalString(encodedServerRole, &serverRole); err != nil {
					return 0, err
				}
				roles[serverRole.Version] = serverRole
				versions = append(versions, serverRole.Version)
			}
			sort.Sort(versions)
			// For each new version bring the server up to date
			for _, version := range versions {
				if _, ok := oldRoles[version]; ok {
					// we've already seen these roles, so nothing to do here
					continue
				}
				serverRole := roles[version]
				var wg sync.WaitGroup
				var addShardErr error
				var addShardOnce sync.Once
				for _, shard := range shards(serverRole) {
					if !containsShard(oldRoles, shard) {
						wg.Add(1)
						go func(shard uint64) {
							defer wg.Done()
							if err := server.AddShard(shard); err != nil {
								addShardOnce.Do(func() {
									addShardErr = err
								})
							}
						}(shard)
					}
				}
				wg.Wait()
				if addShardErr != nil {
					return 0, addShardErr
				}
				oldRoles[version] = serverRole
				versionChan <- version
			}
			// See if there are any old roles that aren't needed
			var wg sync.WaitGroup
			var removeShardErr error
			var removeShardOnce sync.Once
			for version, serverRole := range oldRoles {
				if _, ok := roles[version]; ok {
					// these roles haven't expired yet, so nothing to do
					continue
				}
				for _, shard := range shards(serverRole) {
					if !containsShard(roles, shard) {
						wg.Add(1)
						go func(shard uint64) {
							defer wg.Done()
							if err := server.RemoveShard(shard); err != nil {
								removeShardOnce.Do(func() {
									removeShardErr = err
								})
							}
						}(shard)
					}
				}
			}
			wg.Wait()
			oldRoles = make(map[int64]ServerRole, len(roles))
			for version, serverRole := range roles {
				oldRoles[version] = serverRole
			}
			return 0, removeShardErr
		},
	)
}

func shards(serverRole ServerRole) []uint64 {
	var result []uint64
	for shard := range serverRole.Masters {
		result = append(result, shard)
	}
	for shard := range serverRole.Replicas {
		result = append(result, shard)
	}
	return result
}

func containsShard(roles map[int64]ServerRole, shard uint64) bool {
	for _, serverRole := range roles {
		if serverRole.Masters[shard] || serverRole.Replicas[shard] {
			return true
		}
	}
	return false
}

func sameServers(oldServerStates map[string]ServerState, newServerStates map[string]ServerState) bool {
	if len(oldServerStates) != len(newServerStates) {
		return false
	}
	for id := range oldServerStates {
		if _, ok := newServerStates[id]; !ok {
			return false
		}
	}
	return true
}
