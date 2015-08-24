package role

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/pachyderm/pachyderm/src/pfs/route"
)

var (
	numReplicas = 2
)

type roler struct {
	addresser    route.Addresser
	sharder      route.Sharder
	server       Server
	localAddress string
	cancel       chan bool
}

func newRoler(addresser route.Addresser, sharder route.Sharder, server Server, localAddress string) *roler {
	return &roler{addresser, sharder, server, localAddress, make(chan bool)}
}

func (r *roler) Run() error {
	return r.addresser.WatchShardToAddress(
		r.cancel,
		func(shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[string]bool) (uint64, error) {
			modifiedIndex, ok, err := r.findMasterRole(shardToMasterAddress)
			if ok {
				return modifiedIndex, err
			}
			return 0, nil
		},
	)
}

func (r *roler) Cancel() {
	close(r.cancel)
}

type counts map[string]int

func (r *roler) openMasterRole(shardToMasterAddress map[int]string) (int, bool) {
	for _, i := range rand.Perm(r.sharder.NumShards()) {
		if _, ok := shardToMasterAddress[i]; !ok {
			return i, true
		}
	}
	return 0, false
}

func (r *roler) openReplicaRole(shardToReplicaAddress map[int]map[string]bool) (int, bool) {
	for _, i := range rand.Perm(r.sharder.NumShards()) {
		if addresses := shardToReplicaAddress[i]; len(addresses) < numReplicas {
			return i, true
		}
	}
	return 0, false
}

func (r *roler) randomMasterRole(address string, shardToMasterAddress map[int]string) (int, bool) {
	// we want this function to return a random shard which belongs to address
	// so that not everyone tries to steal the same shard since Go 1 the
	// runtime randomizes iteration of maps to prevent people from depending on
	// a stable ordering. We're doing the opposite here which is depending on
	// the randomness, this seems ok to me but maybe we should change it?
	// Note we only depend on the randomness for performance reason, this code
	// is all still correct if the order isn't random.
	for shard, iAddress := range shardToMasterAddress {
		if address == iAddress {
			return shard, true
		}
	}
	return 0, false
}

func (r *roler) randomReplicaRole(address string, shardToReplicaAddress map[int]map[string]bool) (int, bool) {
	for shard, addresses := range shardToReplicaAddress {
		if _, ok := addresses[address]; ok {
			return shard, true
		}
	}
	return 0, false
}

func (r *roler) masterCounts(shardToMasterAddress map[int]string) counts {
	result := make(map[string]int)
	for _, address := range shardToMasterAddress {
		result[address]++
	}
	return result
}

func (r *roler) replicaCounts(shardToReplicaAddress map[int]map[string]bool) counts {
	result := make(map[string]int)
	for _, addresses := range shardToReplicaAddress {
		for address := range addresses {
			result[address]++
		}
	}
	return result
}

func (r *roler) minCount(counts counts) (string, int) {
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

func (r *roler) maxCount(counts counts) (string, int) {
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

func (r *roler) findMasterRole(shardToMasterAddress map[int]string) (uint64, bool, error) {
	counts := r.masterCounts(shardToMasterAddress)
	_, min := r.minCount(counts)
	if counts[r.localAddress] > min {
		// someone else has fewer roles than us let them claim them
		return 0, false, nil
	}
	shard, ok := r.openMasterRole(shardToMasterAddress)
	if ok {
		modifiedIndex, err := r.addresser.ClaimMasterAddress(shard, r.localAddress, "")
		if err != nil {
			// error from ClaimMasterAddress means our change raced with someone else's,
			// we want to try again so we return nil
			return 0, false, nil
		}
		if err := r.server.Master(shard); err != nil {
			return 0, false, err
		}
		go func() {
			r.addresser.HoldMasterAddress(shard, r.localAddress, r.cancel)
			r.server.Clear(shard)
		}()
		return modifiedIndex, true, nil
	}

	maxAddress, max := r.maxCount(counts)
	if counts[r.localAddress]+1 <= max-1 {
		shard, ok = r.randomMasterRole(maxAddress, shardToMasterAddress)
		if !ok {
			return 0, false, fmt.Errorf("pachyderm: unreachable, randomMasterRole should always return ok")
		}
		modifiedIndex, err := r.addresser.ClaimMasterAddress(shard, r.localAddress, maxAddress)
		if err != nil {
			// error from ClaimMasterAddress means our change raced with someone else's,
			// we want to try again so we return nil
			return 0, false, nil
		}
		if err := r.server.Master(shard); err != nil {
			return 0, false, err
		}
		go func() {
			r.addresser.HoldMasterAddress(shard, r.localAddress, r.cancel)
			r.server.Clear(shard)
		}()
		return modifiedIndex, true, nil
	}
	return 0, false, nil
}

func (r *roler) findReplicaRole(shardToReplicaAddress map[int]map[string]bool) (uint64, bool, error) {
	counts := r.replicaCounts(shardToReplicaAddress)
	_, min := r.minCount(counts)
	if counts[r.localAddress] > min {
		// someone else has fewer roles than us let them claim them
		return 0, false, nil
	}
	shard, ok := r.openReplicaRole(shardToReplicaAddress)
	if ok {
		modifiedIndex, err := r.addresser.ClaimReplicaAddress(shard, r.localAddress, "")
		if err != nil {
			// error from ClaimReplicaAddress means our change raced with someone else's,
			// we want to try again so we return nil
			return 0, false, nil
		}
		if err := r.server.Replica(shard); err != nil {
			return 0, false, err
		}
		go func() {
			r.addresser.HoldReplicaAddress(shard, r.localAddress, r.cancel)
			r.server.Clear(shard)
		}()
		return modifiedIndex, true, nil
	}

	maxAddress, max := r.maxCount(counts)
	if counts[r.localAddress]+1 <= max-1 {
		shard, ok = r.randomReplicaRole(maxAddress, shardToReplicaAddress)
		if !ok {
			return 0, false, fmt.Errorf("pachyderm: unreachable, randomReplicaRole should always return ok")
		}
		modifiedIndex, err := r.addresser.ClaimReplicaAddress(shard, r.localAddress, maxAddress)
		if err != nil {
			// error from ClaimReplicaAddress means our change raced with someone else's,
			// we want to try again so we return nil
			return 0, false, nil
		}
		if err := r.server.Replica(shard); err != nil {
			return 0, false, err
		}
		go func() {
			r.addresser.HoldReplicaAddress(shard, r.localAddress, r.cancel)
			r.server.Clear(shard)
		}()
		return modifiedIndex, true, nil
	}
	return 0, false, nil
}
