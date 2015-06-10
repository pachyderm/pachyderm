package main

import (
	"log"
	"net/http/httptest"
	"strconv"
	"testing"
	"testing/quick"

	"github.com/pachyderm/pfs/lib/etcache"

	"github.com/pachyderm/pfs/lib/router"
	"github.com/pachyderm/pfs/lib/shard"
	"github.com/pachyderm/pfs/lib/traffic"
)

func TestTwoShards(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// used to prevent collisions
	counter := 0
	f := func(w traffic.Workload) bool {
		defer func() { counter++ }()
		suf := strconv.Itoa(counter)
		// Setup 2 shards
		shard1 := shard.NewShard("TestTwoShardsData-0-2"+suf,
			"TestTwoShardsComp-0-2"+suf, "TestTwoShardsPipelines-0-2"+suf, 0, 2)
		shard.Check(shard1.EnsureRepos(), t)
		s1 := httptest.NewServer(shard1.ShardMux())
		defer s1.Close()
		shard2 := shard.NewShard("TestTwoShardsData-1-2"+suf,
			"TestTwoShardsComp-1-2"+suf, "TestTwoShardsPipelines-1-2"+suf, 1, 2)
		shard.Check(shard2.EnsureRepos(), t)
		s2 := httptest.NewServer(shard2.ShardMux())
		defer s2.Close()
		// Setup a Router
		r := httptest.NewServer(router.NewRouter(2).RouterMux())
		defer r.Close()
		// Spoof the shards in etcache
		etcache.SpoofMany("/pfs/master", []string{s1.URL, s2.URL}, false)
		etcache.Spoof1("/pfs/master/0-2", s1.URL)
		etcache.Spoof1("/pfs/master/1-2", s2.URL)
		// Run the workload
		shard.RunWorkload(r.URL, w, t)
		// Make sure we see the changes we should
		facts := w.Facts()
		shard.RunWorkload(r.URL, facts, t)
		//increment the counter
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}
