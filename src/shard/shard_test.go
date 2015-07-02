package shard

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"testing/quick"

	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/pachyderm/pachyderm/src/traffic"
	"github.com/stretchr/testify/require"
)

const (
	defaultMaxCount = 5
	shortMaxCount   = 1
)

func TestPing(t *testing.T) {
	shard := NewShard("TestPingData", "TestPingComp", "TestPingPipelines", 0, 1)
	require.NoError(t, shard.EnsureRepos())
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Get(s.URL + "/ping")
	require.NoError(t, err)
	checkAndCloseHTTPResponseBody(t, res, "pong\n")
}

func TestBasic(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		shard := NewShard(fmt.Sprintf("TestBasic%d", c), fmt.Sprintf("TestBasicComp%d", c), fmt.Sprintf("TestBasicPipelines%d", c), 0, 1)
		c++
		require.NoError(t, shard.EnsureRepos())
		s := httptest.NewServer(shard.ShardMux())
		defer s.Close()

		RunWorkload(t, s.URL, w)
		facts := w.Facts()
		RunWorkload(t, s.URL, facts)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: getMaxCount()}); err != nil {
		t.Error(err)
	}
}

func TestPull(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestPullSrc%d", c), fmt.Sprintf("TestPullSrcComp%d", c), fmt.Sprintf("TestPullSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestPullDst%d", c), fmt.Sprintf("TestPullDstComp%d", c), fmt.Sprintf("TestPullDstPipelines%d", c), 0, 1)
		c++
		require.NoError(t, _src.EnsureRepos())
		require.NoError(t, _dst.EnsureRepos())
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		RunWorkload(t, src.URL, w)

		// Replicate the data
		srcReplica := NewShardReplica(src.URL)
		dstReplica := NewShardReplica(dst.URL)
		err := srcReplica.Pull("", dstReplica)
		require.NoError(t, err)
		facts := w.Facts()
		RunWorkload(t, dst.URL, facts)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: getMaxCount()}); err != nil {
		t.Error(err)
	}
}

// TestSync is similar to TestPull but it does it syncs after every commit.
func TestSyncTo(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestSyncToSrc%d", c), fmt.Sprintf("TestSyncToSrcComp%d", c), fmt.Sprintf("TestSyncToSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestSyncToDst%d", c), fmt.Sprintf("TestSyncToDstComp%d", c), fmt.Sprintf("TestSyncToDstPipelines%d", c), 0, 1)
		require.NoError(t, _src.EnsureRepos())
		require.NoError(t, _dst.EnsureRepos())
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		for _, o := range w {
			runOp(t, src.URL, o)
			if o.Object == traffic.Commit {
				// Replicate the data
				err := SyncTo(fmt.Sprintf("TestSyncToSrc%d", c), []string{dst.URL})
				require.NoError(t, err)
			}
		}

		facts := w.Facts()
		RunWorkload(t, dst.URL, facts)

		c++
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: getMaxCount()}); err != nil {
		t.Error(err)
	}
}

// TestSyncFrom
func TestSyncFrom(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestSyncFromSrc%d", c), fmt.Sprintf("TestSyncFromSrcComp%d", c), fmt.Sprintf("TestSyncFromSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestSyncFromDst%d", c), fmt.Sprintf("TestSyncFromDstComp%d", c), fmt.Sprintf("TestSyncFromDstPipelines%d", c), 0, 1)
		require.NoError(t, _src.EnsureRepos())
		require.NoError(t, _dst.EnsureRepos())
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		for _, o := range w {
			runOp(t, src.URL, o)
			if o.Object == traffic.Commit {
				// Replicate the data
				err := SyncFrom(fmt.Sprintf("TestSyncFromDst%d", c), []string{src.URL})
				require.NoError(t, err)
			}
		}

		facts := w.Facts()
		RunWorkload(t, dst.URL, facts)

		c++
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: getMaxCount()}); err != nil {
		t.Error(err)
	}
}

// TestPipeline creates a basic pipeline on a shard.
func TestPipeline(t *testing.T) {
	shard := NewShard("TestPipelineData", "TestPipelineComp", "TestPipelinePipelines", 0, 1)
	require.NoError(t, shard.EnsureRepos())
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/touch_foo", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
`))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)
	checkFile(t, s.URL+"/pipeline/touch_foo", "foo", "commit1", "")
}

// TestShardFilter creates a basic pipeline on a shard and then requests files
// from it using shard filtering.

func TestShardFilter(t *testing.T) {
	shard := NewShard("TestShardFilterData", "TestShardFilterComp", "TestShardFilterPipelines", 0, 1)
	require.NoError(t, shard.EnsureRepos())
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/files", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
run touch /out/bar
run touch /out/buzz
run touch /out/bizz
`))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)

	// Map to store files we receive
	files := make(map[string]struct{})
	res, err = http.Get(s.URL + path.Join("/pipeline", "files", "file", "*") + "?commit=commit1&shard=0-2")
	require.NoError(t, err)
	if res.StatusCode != 200 {
		t.Fatal(res.Status)
	}
	reader := multipart.NewReader(res.Body, res.Header.Get("Boundary"))

	for p, err := reader.NextPart(); err != io.EOF; p, err = reader.NextPart() {
		match, err := route.Match(p.FileName(), "0-2")
		require.NoError(t, err)
		if !match {
			t.Fatalf("Filename: %s should match.", p.FileName())
		}
		if _, ok := files[p.FileName()]; ok == true {
			t.Fatalf("File: %s received twice.", p.FileName())
		}
		files[p.FileName()] = struct{}{}
	}

	res, err = http.Get(s.URL + path.Join("/pipeline", "files", "file", "*") + "?commit=commit1&shard=1-2")
	require.NoError(t, err)
	if res.StatusCode != 200 {
		t.Fatal(res.Status)
	}
	reader = multipart.NewReader(res.Body, res.Header.Get("Boundary"))

	for p, err := reader.NextPart(); err != io.EOF; p, err = reader.NextPart() {
		match, err := route.Match(p.FileName(), "1-2")
		require.NoError(t, err)
		if !match {
			t.Fatalf("Filename: %s should match.", p.FileName())
		}
		if _, ok := files[p.FileName()]; ok == true {
			t.Fatalf("File: %s received twice.", p.FileName())
		}
		files[p.FileName()] = struct{}{}
	}
}

func TestShuffle(t *testing.T) {
	// Setup 2 shards
	shard1 := NewShard("TestShuffleData-0-2", "TestShuffleComp-0-2", "TestShufflePipelines-0-2", 0, 2)
	require.NoError(t, shard1.EnsureRepos())
	s1 := httptest.NewServer(shard1.ShardMux())
	defer s1.Close()
	shard2 := NewShard("TestShuffleData-1-2", "TestShuffleComp-1-2", "TestShufflePipelines-1-2", 1, 2)
	require.NoError(t, shard2.EnsureRepos())
	s2 := httptest.NewServer(shard2.ShardMux())
	defer s2.Close()

	files := []string{"foo", "bar", "fizz", "buzz"}

	for _, file := range files {
		checkWriteFile(t, s1.URL, path.Join("data", file), "master", file)
		checkWriteFile(t, s2.URL, path.Join("data", file), "master", file)
	}

	// Spoof the shards in etcache
	etcache.SpoofMany("/pfs/master", []string{s1.URL, s2.URL}, false)

	pipeline := `
image ubuntu

input data

run cp -r /in/data /out

shuffle data
`
	res, err := http.Post(s1.URL+"/pipeline/shuffle", "application/text", strings.NewReader(pipeline))
	require.NoError(t, err)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/pipeline/shuffle", "application/text", strings.NewReader(pipeline))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s1.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)
	res, err = http.Post(s2.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)

	for _, file := range files {
		match, err := route.Match(path.Join("data", file), "0-2")
		require.NoError(t, err)
		if match {
			log.Print("shard: s1 file: ", file)
			checkFile(t, s1.URL+"/pipeline/shuffle", path.Join("data", file), "commit1", file+file)
		} else {
			log.Print("shard: s2 file: ", file)
			checkFile(t, s2.URL+"/pipeline/shuffle", path.Join("data", file), "commit1", file+file)
		}
	}
}

func TestWordCount(t *testing.T) {
	// Setup 2 shards
	shard1 := NewShard("TestWordCountData-0-2", "TestWordCountComp-0-2", "TestWordCountPipelines-0-2", 0, 2)
	require.NoError(t, shard1.EnsureRepos())
	s1 := httptest.NewServer(shard1.ShardMux())
	defer s1.Close()
	shard2 := NewShard("TestWordCountData-1-2", "TestWordCountComp-1-2", "TestWordCountPipelines-1-2", 1, 2)
	require.NoError(t, shard2.EnsureRepos())
	s2 := httptest.NewServer(shard2.ShardMux())
	defer s2.Close()

	checkWriteFile(t, s1.URL, path.Join("data", "1"), "master",
		`Mr and Mrs Dursley, of number four, Privet Drive, were proud to say that they
were perfectly normal, thank you very much. They were the last people you'd
expect to be involved in anything strange or mysterious, because they just
didn't hold with such nonsense.`)
	checkWriteFile(t, s2.URL, path.Join("data", "2"), "master",
		`Mr Dursley was the director of a firm called Grunnings, which made drills.
He was a big, beefy man with hardly any neck, although he did have a very
large moustache. Mrs Dursley was thin and blonde and had nearly twice the
usual amount of neck, which came in very useful as she spent so much of her
time craning over garden fences, spying on the neighbours. The Dursleys had
a small son called Dudley and in their opinion there was no finer boy
anywhere.`)

	// Spoof the shards in etcache
	etcache.SpoofMany("/pfs/master", []string{s1.URL, s2.URL}, false)

	pipeline := `
image ubuntu

input data

run mkdir /out/counts
run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
`
	res, err := http.Post(s1.URL+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
	require.NoError(t, err)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s1.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)
	res.Body.Close()

	// There should be 3 occurances of Dursley
	checkFile(t, s1.URL+"/pipeline/wc", path.Join("counts", "Dursley"), "commit1", "3\n")
}

func TestFail(t *testing.T) {
	shard := NewShard("TestFailData", "TestFailComp", "TestFailPipelines", 0, 1)
	require.NoError(t, shard.EnsureRepos())
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/fail", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
run exit 1
`))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)

	res, err = http.Get(s.URL + "/pipeline/fail/file/foo?commit=commit1")
	require.NoError(t, err)
	if res.StatusCode != 500 {
		t.Fatal("Request should return failure.")
	}
}

// TestChess uses our chess data set to test s3 integration.
func TestChess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// Notice this shard is behaving like 1 node of a 5000 node cluster to downsample to data.
	shard := NewShard("TestChessData", "TestChessComp", "TestChessPipelines", 0, 5000)
	require.NoError(t, shard.EnsureRepos())
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/count", "application/text", strings.NewReader(`
image ubuntu

input s3://pachyderm-data/chess

run cat /in/pachyderm-data/chess/* | wc -l > /out/count
`))
	require.NoError(t, err)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	require.NoError(t, err)

	res, err = http.Get(s.URL + "/pipeline/count/file/count?commit=commit1")
	require.NoError(t, err)
	if res.StatusCode != 200 {
		t.Fatal("Bad status code.")
	}
}

func getMaxCount() int {
	if testing.Short() {
		return defaultMaxCount
	}
	return shortMaxCount
}
