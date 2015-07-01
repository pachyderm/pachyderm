package shard

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"testing/quick"

	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/pachyderm/pachyderm/src/traffic"
)

func TestPing(t *testing.T) {
	shard := NewShard("TestPingData", "TestPingComp", "TestPingPipelines", 0, 1)
	Check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Get(s.URL + "/ping")
	Check(err, t)
	CheckResp(res, "pong\n", t)
	res.Body.Close()
}

func TestBasic(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		shard := NewShard(fmt.Sprintf("TestBasic%d", c), fmt.Sprintf("TestBasicComp%d", c), fmt.Sprintf("TestBasicPipelines%d", c), 0, 1)
		c++
		Check(shard.EnsureRepos(), t)
		s := httptest.NewServer(shard.ShardMux())
		defer s.Close()

		RunWorkload(s.URL, w, t)
		facts := w.Facts()
		RunWorkload(s.URL, facts, t)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}

func TestPull(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestPullSrc%d", c), fmt.Sprintf("TestPullSrcComp%d", c), fmt.Sprintf("TestPullSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestPullDst%d", c), fmt.Sprintf("TestPullDstComp%d", c), fmt.Sprintf("TestPullDstPipelines%d", c), 0, 1)
		c++
		Check(_src.EnsureRepos(), t)
		Check(_dst.EnsureRepos(), t)
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		RunWorkload(src.URL, w, t)

		// Replicate the data
		srcReplica := NewShardReplica(src.URL)
		dstReplica := NewShardReplica(dst.URL)
		err := srcReplica.Pull("", dstReplica)
		Check(err, t)
		facts := w.Facts()
		RunWorkload(dst.URL, facts, t)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}

// TestSync is similar to TestPull but it does it syncs after every commit.
func TestSyncTo(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestSyncToSrc%d", c), fmt.Sprintf("TestSyncToSrcComp%d", c), fmt.Sprintf("TestSyncToSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestSyncToDst%d", c), fmt.Sprintf("TestSyncToDstComp%d", c), fmt.Sprintf("TestSyncToDstPipelines%d", c), 0, 1)
		Check(_src.EnsureRepos(), t)
		Check(_dst.EnsureRepos(), t)
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		for _, o := range w {
			RunOp(src.URL, o, t)
			if o.Object == traffic.Commit {
				// Replicate the data
				err := SyncTo(fmt.Sprintf("TestSyncToSrc%d", c), []string{dst.URL})
				Check(err, t)
			}
		}

		facts := w.Facts()
		RunWorkload(dst.URL, facts, t)

		c++
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}

// TestSyncFrom
func TestSyncFrom(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	c := 0
	f := func(w traffic.Workload) bool {
		_src := NewShard(fmt.Sprintf("TestSyncFromSrc%d", c), fmt.Sprintf("TestSyncFromSrcComp%d", c), fmt.Sprintf("TestSyncFromSrcPipelines%d", c), 0, 1)
		_dst := NewShard(fmt.Sprintf("TestSyncFromDst%d", c), fmt.Sprintf("TestSyncFromDstComp%d", c), fmt.Sprintf("TestSyncFromDstPipelines%d", c), 0, 1)
		Check(_src.EnsureRepos(), t)
		Check(_dst.EnsureRepos(), t)
		src := httptest.NewServer(_src.ShardMux())
		dst := httptest.NewServer(_dst.ShardMux())
		defer src.Close()
		defer dst.Close()

		for _, o := range w {
			RunOp(src.URL, o, t)
			if o.Object == traffic.Commit {
				// Replicate the data
				err := SyncFrom(fmt.Sprintf("TestSyncFromDst%d", c), []string{src.URL})
				Check(err, t)
			}
		}

		facts := w.Facts()
		RunWorkload(dst.URL, facts, t)

		c++
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error(err)
	}
}

// TestPipeline creates a basic pipeline on a shard.
func TestPipeline(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	shard := NewShard("TestPipelineData", "TestPipelineComp", "TestPipelinePipelines", 0, 1)
	Check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/touch_foo", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
`))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)
	Checkfile(s.URL+"/pipeline/touch_foo", "foo", "commit1", "", t)
}

// TestShardFilter creates a basic pipeline on a shard and then requests files
// from it using shard filtering.

func TestShardFilter(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	shard := NewShard("TestShardFilterData", "TestShardFilterComp", "TestShardFilterPipelines", 0, 1)
	Check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/files", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
run touch /out/bar
run touch /out/buzz
run touch /out/bizz
`))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)

	// Map to store files we receive
	files := make(map[string]struct{})
	res, err = http.Get(s.URL + path.Join("/pipeline", "files", "file", "*") + "?commit=commit1&shard=0-2")
	Check(err, t)
	if res.StatusCode != 200 {
		t.Fatal(res.Status)
	}
	reader := multipart.NewReader(res.Body, res.Header.Get("Boundary"))

	for p, err := reader.NextPart(); err != io.EOF; p, err = reader.NextPart() {
		match, err := route.Match(p.FileName(), "0-2")
		Check(err, t)
		if !match {
			t.Fatalf("Filename: %s should match.", p.FileName())
		}
		if _, ok := files[p.FileName()]; ok == true {
			t.Fatalf("File: %s received twice.")
		}
		files[p.FileName()] = struct{}{}
	}

	res, err = http.Get(s.URL + path.Join("/pipeline", "files", "file", "*") + "?commit=commit1&shard=1-2")
	Check(err, t)
	if res.StatusCode != 200 {
		t.Fatal(res.Status)
	}
	reader = multipart.NewReader(res.Body, res.Header.Get("Boundary"))

	for p, err := reader.NextPart(); err != io.EOF; p, err = reader.NextPart() {
		match, err := route.Match(p.FileName(), "1-2")
		Check(err, t)
		if !match {
			t.Fatalf("Filename: %s should match.", p.FileName())
		}
		if _, ok := files[p.FileName()]; ok == true {
			t.Fatalf("File: %s received twice.")
		}
		files[p.FileName()] = struct{}{}
	}
}

func TestShuffle(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	// Setup 2 shards
	shard1 := NewShard("TestShuffleData-0-2", "TestShuffleComp-0-2", "TestShufflePipelines-0-2", 0, 2)
	Check(shard1.EnsureRepos(), t)
	s1 := httptest.NewServer(shard1.ShardMux())
	defer s1.Close()
	shard2 := NewShard("TestShuffleData-1-2", "TestShuffleComp-1-2", "TestShufflePipelines-1-2", 1, 2)
	Check(shard2.EnsureRepos(), t)
	s2 := httptest.NewServer(shard2.ShardMux())
	defer s2.Close()

	files := []string{"foo", "bar", "fizz", "buzz"}

	for _, file := range files {
		WriteFile(s1.URL, path.Join("data", file), "master", file, t)
		WriteFile(s2.URL, path.Join("data", file), "master", file, t)
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
	Check(err, t)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/pipeline/shuffle", "application/text", strings.NewReader(pipeline))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s1.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)
	res, err = http.Post(s2.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)

	for _, file := range files {
		match, err := route.Match(path.Join("data", file), "0-2")
		Check(err, t)
		if match {
			log.Print("shard: s1 file: ", file)
			Checkfile(s1.URL+"/pipeline/shuffle", path.Join("data", file), "commit1", file+file, t)
		} else {
			log.Print("shard: s2 file: ", file)
			Checkfile(s2.URL+"/pipeline/shuffle", path.Join("data", file), "commit1", file+file, t)
		}
	}
}

func TestWordCount(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	// Setup 2 shards
	shard1 := NewShard("TestWordCountData-0-2", "TestWordCountComp-0-2", "TestWordCountPipelines-0-2", 0, 2)
	Check(shard1.EnsureRepos(), t)
	s1 := httptest.NewServer(shard1.ShardMux())
	defer s1.Close()
	shard2 := NewShard("TestWordCountData-1-2", "TestWordCountComp-1-2", "TestWordCountPipelines-1-2", 1, 2)
	Check(shard2.EnsureRepos(), t)
	s2 := httptest.NewServer(shard2.ShardMux())
	defer s2.Close()

	WriteFile(s1.URL, path.Join("data", "1"), "master",
		`Mr and Mrs Dursley, of number four, Privet Drive, were proud to say that they
were perfectly normal, thank you very much. They were the last people you'd
expect to be involved in anything strange or mysterious, because they just
didn't hold with such nonsense.`, t)
	WriteFile(s2.URL, path.Join("data", "2"), "master",
		`Mr Dursley was the director of a firm called Grunnings, which made drills.
He was a big, beefy man with hardly any neck, although he did have a very
large moustache. Mrs Dursley was thin and blonde and had nearly twice the
usual amount of neck, which came in very useful as she spent so much of her
time craning over garden fences, spying on the neighbours. The Dursleys had
a small son called Dudley and in their opinion there was no finer boy
anywhere.`, t)

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
	Check(err, t)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s1.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)
	res.Body.Close()
	res, err = http.Post(s2.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)
	res.Body.Close()

	// There should be 3 occurances of Dursley
	Checkfile(s1.URL+"/pipeline/wc", path.Join("counts", "Dursley"), "commit1", "3\n", t)
}

func TestFail(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	shard := NewShard("TestFailData", "TestFailComp", "TestFailPipelines", 0, 1)
	Check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/fail", "application/text", strings.NewReader(`
image ubuntu

run touch /out/foo
run exit 1
`))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)

	res, err = http.Get(s.URL + "/pipeline/fail/file/foo?commit=commit1")
	Check(err, t)
	if res.StatusCode != 500 {
		t.Fatal("Request should return failure.")
	}
}

// TestChess uses our chess data set to test s3 integration.
func TestChess(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// Notice this shard is behaving like 1 node of a 5000 node cluster to downsample to data.
	shard := NewShard("TestChessData", "TestChessComp", "TestChessPipelines", 0, 5000)
	Check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Post(s.URL+"/pipeline/count", "application/text", strings.NewReader(`
image ubuntu

input s3://pachyderm-data/chess

run cat /in/pachyderm-data/chess/* | wc -l > /out/count
`))
	Check(err, t)
	res.Body.Close()

	res, err = http.Post(s.URL+"/commit?commit=commit1", "", nil)
	Check(err, t)

	res, err = http.Get(s.URL + "/pipeline/count/file/count?commit=commit1")
	Check(err, t)
	if res.StatusCode != 200 {
		t.Fatal("Bad status code.")
	}
}
