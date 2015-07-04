package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/quick"

	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/storage"
	"github.com/pachyderm/pachyderm/src/traffic"
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

func NewCluster(prefix string, shards int, testCache etcache.TestCache, t *testing.T) Cluster {
	var res Cluster
	for i := 0; i < shards; i++ {
		repoStr := fmt.Sprintf("%s-%d-%d", prefix, i, shards)
		s := storage.NewShard("", repoStr+"-data", repoStr+"-comp",
			repoStr+"-pipeline", uint64(i), uint64(shards), testCache)
		if err := s.EnsureRepos(); err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(storage.NewShardHTTPHandler(s))
		res.shards = append(res.shards, server)
		testCache.SpoofOne(fmt.Sprintf("/pfs/master/%d-%d", i, shards), server.URL)
	}
	var urls []string
	for _, server := range res.shards {
		urls = append(urls, server.URL)
	}
	testCache.SpoofMany("/pfs/master", urls, false)
	res.router = httptest.NewServer(NewRouter(uint64(shards), testCache).RouterMux())
	return res
}

func TestTwoShards(t *testing.T) {
	t.Parallel()
	maxCount := 5
	if testing.Short() {
		maxCount = 1
	}
	// used to prevent collisions
	counter := 0
	f := func(w traffic.Workload) bool {
		defer func() { counter++ }()
		cluster := NewCluster(fmt.Sprintf("TestTwoShards-%d", counter), 2, etcache.NewTestCache(), t)
		defer cluster.Close()
		// Run the workload
		storage.RunWorkload(t, cluster.router.URL, w)
		// Make sure we see the changes we should
		facts := w.Facts()
		storage.RunWorkload(t, cluster.router.URL, facts)
		//increment the counter
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: maxCount}); err != nil {
		t.Error(err)
	}
}

func TestWordCount(t *testing.T) {
	t.Parallel()
	maxCount := 2
	if testing.Short() {
		maxCount = 1
	}
	// First setup the WordCount pipeline
	pipeline := `
image ubuntu

input data

run mkdir -p /out/counts
run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
`
	// used to prevent collisions
	counter := 0
	f := func(w traffic.Workload) bool {
		defer func() { counter++ }()
		cluster := NewCluster(fmt.Sprintf("TestWordCount-%d", counter), 4, etcache.NewTestCache(), t)
		defer cluster.Close()
		// Run the workload
		storage.RunWorkload(t, cluster.router.URL, w)
		// Install the pipeline
		response, err := http.Post(cluster.router.URL+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
		defer response.Body.Close()
		if err != nil {
			t.Error(err)
		}
		// Make a commit
		response, err = http.Post(cluster.router.URL+"/commit?commit=commit1", "", nil)
		defer response.Body.Close()
		if err != nil {
			t.Error(err)
		}
		// TODO(jd) make this check for correctness, not just that the request
		// completes. It's a bit hard because the input is random. Probably the
		// right idea is to modify the traffic package so that it keeps track of
		// this.
		response, err = http.Get(cluster.router.URL + "/pipeline/wc/file/counts/*?commit=commit1")
		defer response.Body.Close()
		if err != nil {
			t.Error(err)
		}
		if response.StatusCode != 200 {
			t.Fatal("Bad status code.")
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: maxCount}); err != nil {
		t.Error(err)
	}
}
