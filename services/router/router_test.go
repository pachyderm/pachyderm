package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/pachyderm/pfs/lib/etcache"

	"github.com/pachyderm/pfs/lib/router"
	"github.com/pachyderm/pfs/lib/shard"
	"github.com/pachyderm/pfs/lib/traffic"
)

type Cluster struct {
	router *httptest.Server
	shards []*httptest.Server
}

func (c Cluster) Close() {
	c.router.Close()
	for _, shard := range c.shards {
		shard.Close()
	}
}

func NewCluster(prefix string, shards int, t *testing.T) Cluster {
	var res Cluster
	for i := 0; i < shards; i++ {
		repoStr := fmt.Sprintf("%s-%d-%d", prefix, i, shards)
		s := shard.NewShard(repoStr+"-data", repoStr+"-comp",
			repoStr+"-pipeline", uint64(i), uint64(shards))
		shard.Check(s.EnsureRepos(), t)
		server := httptest.NewServer(s.ShardMux())
		res.shards = append(res.shards, server)
		etcache.Spoof1(fmt.Sprintf("/pfs/master/%d-%d", i, shards), server.URL)
	}
	var urls []string
	for _, server := range res.shards {
		urls = append(urls, server.URL)
	}
	etcache.SpoofMany("/pfs/master", urls, false)
	res.router = httptest.NewServer(router.NewRouter(uint64(shards)).RouterMux())
	return res
}

func TestTwoShards(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// used to prevent collisions
	counter := 0
	f := func(w traffic.Workload) bool {
		defer func() { counter++ }()
		cluster := NewCluster(fmt.Sprintf("TestTwoShards-%d", counter), 2, t)
		defer cluster.Close()
		// Run the workload
		shard.RunWorkload(cluster.router.URL, w, t)
		// Make sure we see the changes we should
		facts := w.Facts()
		shard.RunWorkload(cluster.router.URL, facts, t)
		//increment the counter
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}

func TestWordCount(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// First setup the WordCount pipeline

	pipeline := `
image ubuntu

input .

run mkdir -p /out/counts
run cat /in/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
`
	// used to prevent collisions
	counter := 0
	f := func(w traffic.Workload) bool {
		defer func() { counter++ }()
		cluster := NewCluster(fmt.Sprintf("TestWordCount-%d", counter), 2, t)
		defer cluster.Close()
		// Install the pipeline
		res, err := http.Post(cluster.router.URL+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
		shard.Check(err, t)
		res.Body.Close()
		// Run the workload
		shard.RunWorkload(cluster.router.URL, w, t)
		// Make sure we see the changes we should
		facts := w.Facts()
		shard.RunWorkload(cluster.router.URL, facts, t)
		time.Sleep(time.Second * 5)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 1}); err != nil {
		t.Error(err)
	}

}
