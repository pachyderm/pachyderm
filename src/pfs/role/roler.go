package role

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/pachyderm/pachyderm/src/pfs/route"
)

type roler struct {
	addresser    route.Addresser
	sharder      route.Sharder
	server       Server
	localAddress string
	cancel       chan bool
	numReplicas  int
}

func newRoler(addresser route.Addresser, sharder route.Sharder, server Server, localAddress string, numReplicas int) *roler {
	return &roler{addresser, sharder, server, localAddress, make(chan bool), numReplicas}
}

func (r *roler) Run() error {
	return r.addresser.WatchShardToAddress(r.cancel, r.findRole)
}

func (r *roler) Cancel() {
	close(r.cancel)
}

func (r *roler) hasRoleForShard(shard int, shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[int]string) bool {
	address, ok := shardToMasterAddress[shard]
	if ok && address == r.localAddress {
		return true
	}

	addresses, ok := shardToReplicaAddress[shard]
	if ok {
		for _, address := range addresses {
			if address == r.localAddress {
				return true
			}
		}
	}
	return false
}

func (r *roler) openMasterRole(shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[int]string) (int, bool) {
	for _, i := range rand.Perm(r.sharder.NumShards()) {
		_, ok := shardToMasterAddress[i]
		if !ok && !r.hasRoleForShard(i, shardToMasterAddress, shardToReplicaAddress) {
			return i, true
		}
	}
	return 0, false
}

func (r *roler) openReplicaRole(shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[int]string) (int, int, bool) {
	for _, shard := range rand.Perm(r.sharder.NumShards()) {
		addresses := shardToReplicaAddress[shard]
		if len(addresses) < r.numReplicas && !r.hasRoleForShard(shard, shardToMasterAddress, shardToReplicaAddress) {
			for _, index := range rand.Perm(r.numReplicas) {
				if _, ok := addresses[index]; !ok {
					return shard, index, true
				}
			}
		}
	}
	return 0, 0, false
}

func (r *roler) randomMasterRole(
	maxAddress string,
	shardToMasterAddress map[int]string,
	shardToReplicaAddress map[int]map[int]string,
) (int, bool) {
	// we want this function to return a random shard which belongs to address
	// so that not everyone tries to steal the same shard since Go 1 the
	// runtime randomizes iteration of maps to prevent people from depending on
	// a stable ordering. We're doing the opposite here which is depending on
	// the randomness, this seems ok to me but maybe we should change it?
	// Note we only depend on the randomness for performance reason, this code
	// is all still correct if the order isn't random.
	for shard, address := range shardToMasterAddress {
		if address == maxAddress && !r.hasRoleForShard(shard, shardToMasterAddress, shardToReplicaAddress) {
			return shard, true
		}
	}
	return 0, false
}

func (r *roler) randomReplicaRole(
	address string,
	shardToMasterAddress map[int]string,
	shardToReplicaAddress map[int]map[int]string,
) (int, int, bool) {
	for shard, addresses := range shardToReplicaAddress {
		for index, address := range addresses {
			if address == address && !r.hasRoleForShard(shard, shardToMasterAddress, shardToReplicaAddress) {
				return shard, index, true
			}
		}
	}
	return 0, 0, false
}

func (r *roler) masterCounts(shardToMasterAddress map[int]string, out map[string]int) {
	for _, address := range shardToMasterAddress {
		out[address]++
	}
}

func (r *roler) replicaCounts(shardToReplicaAddress map[int]map[int]string, out map[string]int) {
	for _, addresses := range shardToReplicaAddress {
		for _, address := range addresses {
			out[address]++
		}
	}
}

func (r *roler) minCount(counts map[string]int) (string, int) {
	address := ""
	result := math.MaxInt64
	for iAddress, count := range counts {
		if count < result {
			address = iAddress
			result = count
		}
	}
	return address, result
}

func (r *roler) maxCount(counts map[string]int) (string, int) {
	address := ""
	result := 0
	for iAddress, count := range counts {
		if count > result {
			address = iAddress
			result = count
		}
	}
	return address, result
}

func (r *roler) beMaster(shard int, prevAddress string) (uint64, error) {
	modifiedIndex, err := r.addresser.ClaimMasterAddress(shard, r.localAddress, prevAddress)
	if err != nil {
		return 0, nil
	}
	if err := r.server.Master(shard); err != nil {
		return 0, err
	}
	go func() {
		r.addresser.HoldMasterAddress(shard, r.localAddress, r.cancel)
		r.server.Clear(shard)
	}()
	return modifiedIndex, nil
}

func (r *roler) beReplica(shard int, index int, prevAddress string) (uint64, error) {
	modifiedIndex, err := r.addresser.ClaimReplicaAddress(shard, index, r.localAddress, prevAddress)
	if err != nil {
		return 0, nil
	}
	if err := r.server.Replica(shard); err != nil {
		return 0, err
	}
	go func() {
		r.addresser.HoldReplicaAddress(shard, index, r.localAddress, r.cancel)
		r.server.Clear(shard)
	}()
	return modifiedIndex, nil
}

func (r *roler) findRole(shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[int]string) (uint64, error) {
	counts := make(map[string]int)
	r.masterCounts(shardToMasterAddress, counts)
	r.replicaCounts(shardToReplicaAddress, counts)
	_, min := r.minCount(counts)
	if counts[r.localAddress] > min {
		// someone else has fewer roles than us let them claim them
		return 0, nil
	}
	// No server has fewer roles than us, that means it's our turn to claim a role.
	// We start with the most important thing for the cluster, unclaimed master
	// slots.
	shard, ok := r.openMasterRole(shardToMasterAddress, shardToReplicaAddress)
	if ok {
		return r.beMaster(shard, "")
	}

	// No open masters found. Next we look for unclaimed replica roles.
	shard, index, ok := r.openReplicaRole(shardToMasterAddress, shardToReplicaAddress)
	if ok {
		return r.beReplica(shard, index, "")
	}

	// No unclaimed roles were found. Next we check if there's someone in the
	// cluster we could steal from.
	// First we find out who the most role rich server is.
	maxAddress, max := r.maxCount(counts)
	if counts[r.localAddress]+1 <= max-1 {
		// When stealing we consider replica shards first because migrating
		// replicas is less disruptive to the cluster.
		shard, index, ok = r.randomReplicaRole(maxAddress, shardToMasterAddress, shardToReplicaAddress)
		if ok {
			return r.beReplica(shard, index, maxAddress)
		}

		// Lastly we steal a master role
		shard, ok = r.randomMasterRole(maxAddress, shardToMasterAddress, shardToReplicaAddress)
		if ok {
			return r.beMaster(shard, maxAddress)
		}
		return 0, fmt.Errorf("Error we need to be able to find a role here.")
	}
	return 0, nil
}
