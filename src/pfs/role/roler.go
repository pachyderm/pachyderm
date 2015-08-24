package role

import (
	"fmt"
	"log"
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
}

func newRoler(addresser route.Addresser, sharder route.Sharder, server Server, localAddress string) *roler {
	return &roler{addresser, sharder, server, localAddress, make(chan bool)}
}

func (r *roler) Run() error {
	return r.addresser.WatchShardToMasterAddress(
		r.cancel,
		func(shardToMasterAddress map[int]string) (uint64, error) {
			counts := r.masterCounts(shardToMasterAddress)
			minAddress, min := r.minCount(counts)
			if counts[r.localAddress] > min {
				// someone else has fewer roles than us let them claim them
				log.Printf("r.localAddress %s, shardToMasterAddress: %+v\n%s has few roles (%d)", r.localAddress, shardToMasterAddress, minAddress, min)
				return 0, nil
			}
			shard, ok := r.openShard(shardToMasterAddress)
			if ok {
				log.Printf("%s.Master(%d)", r.localAddress, shard)
				//TODO constant
				modifiedIndex, err := r.addresser.ClaimMasterAddress(shard, r.localAddress, 20, "")
				if err != nil {
					// error from ClaimMasterAddress means our change raced with someone else's,
					// we want to try again so we return nil
					return 0, nil
				}
				if err := r.server.Master(shard); err != nil {
					return 0, err
				}
				go func() {
					r.addresser.HoldMasterAddress(shard, r.localAddress, r.cancel)
					log.Printf("%s.Clear(%d)", r.localAddress, shard)
					r.server.Clear(shard)
				}()
				log.Printf("r.localAddress %s, shardToMasterAddress: %+v\nopen: %d", r.localAddress, shardToMasterAddress, shard)
				return modifiedIndex, nil
			}

			maxAddress, max := r.maxCount(counts)
			if maxAddress == r.localAddress || counts[r.localAddress]+1 > max-1 {
				// either we're the maxAddress or stealing a role from
				// maxAddress would make us the new maxAddress that'd cause
				// flappying which is bad
				log.Printf("r.localAddress %s, shardToMasterAddress: %+v\ntoo many shards (%d)", r.localAddress, shardToMasterAddress, counts[r.localAddress])
				return 0, nil
			}
			shard, ok = r.randomShard(maxAddress, shardToMasterAddress)
			if ok {
				log.Printf("r.localAddress %s, shardToMasterAddress: %+v\nstealing %d from %s", r.localAddress, shardToMasterAddress, shard, maxAddress)
				log.Printf("%s.Master(%d)", r.localAddress, shard)
				modifiedIndex, err := r.addresser.ClaimMasterAddress(shard, r.localAddress, 20, maxAddress)
				if err != nil {
					// error from ClaimMasterAddress means our change raced with someone else's,
					// we want to try again so we return nil
					return 0, nil
				}
				if err := r.server.Master(shard); err != nil {
					return 0, err
				}
				go func() {
					r.addresser.HoldMasterAddress(shard, r.localAddress, r.cancel)
					log.Printf("%s.Clear(%d)", r.localAddress, shard)
					r.server.Clear(shard)
				}()
				return modifiedIndex, nil
			}
			return 0, fmt.Errorf("pachyderm: unreachable, randomShard should always return ok")
		},
	)
}

func (r *roler) Cancel() {
	close(r.cancel)
}

type counts map[string]int

func (r *roler) openShard(shardToMasterAddress map[int]string) (int, bool) {
	for _, i := range rand.Perm(r.sharder.NumShards()) {
		if _, ok := shardToMasterAddress[i]; !ok {
			return i, true
		}
	}
	return 0, false
}

func (r *roler) randomShard(address string, shardToMasterAddress map[int]string) (int, bool) {
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

func (r *roler) masterCounts(shardToMasterAddress map[int]string) counts {
	result := make(map[string]int)
	for _, address := range shardToMasterAddress {
		result[address]++
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
