package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"path"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pachyderm-io/pfs/lib/mapreduce"
)

var KB int64 = 1 << 10
var MB int64 = 1 << 20
var GB int64 = 1 << 30
var TB int64 = 1 << 40
var PB int64 = 1 << 50

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var once sync.Once

func randSeq(n int) string {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type ConstReader struct{}

func (r ConstReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'a'
	}
	return len(p), nil
}

var reader ConstReader

// insert inserts a single file in to the filesystem
func insert(dir string, fileSize int64, t testing.TB) {
	url := "http://172.17.42.1/file/" + path.Join(dir, randSeq(10))
	resp, err := http.Post(url, "application/text", io.LimitReader(reader, fileSize))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}

func traffic(dir string, fileSize, sizeLimit int64, t testing.TB) {
	workers := 8
	var wg sync.WaitGroup
	wg.Add(workers)
	var totalSize int64
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for atomic.AddInt64(&totalSize, fileSize) <= sizeLimit {
				insert(dir, fileSize, t)
			}
		}()
	}
	wg.Wait()
}

func newJob(job mapreduce.Job, t testing.TB) {
	jobJson, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	url := "http://172.17.42.1/job/" + randSeq(4)
	resp, err := http.Post(url, "application/text", bytes.NewReader(jobJson))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}

func commit(t testing.TB) {
	resp, err := http.Post("http://172.17.42.1/commit", "application/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}

func TestSmoke(t *testing.T) {
	commit(t)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			insert("", 4*KB, t)
		}
		commit(t)
	}
}

func TestFire(t *testing.T) {
	commit(t)
	for i := 0; i < 5; i++ {
		traffic("", 4*KB, 5*MB, t)
		commit(t)
	}
}

func TestMRInsert(t *testing.T) {
	commit(t)
	newJob(mapreduce.Job{Input: "TestMRInsert", Container: "jdoliner/hello-world", Command: []string{"/go/bin/hello-world-mr"}}, t)
	insert("TestMRInsert", 4*KB, t)
	commit(t)
}

func TestMRTraffic(t *testing.T) {
	commit(t)
	newJob(mapreduce.Job{Input: "TestMRTraffic", Container: "jdoliner/hello-world", Command: []string{"/go/bin/hello-world-mr"}}, t)
	traffic("TestMRTraffic", 4*KB, 128*KB, t)
	commit(t)
}

func _BenchmarkInsert(fileSize int64, b *testing.B) {
	commit(b)
	for i := 0; i < b.N; i++ {
		insert("", fileSize, b)
		commit(b)
	}
	commit(b)
}

func BenchmarkInsert1B(b *testing.B) {
	_BenchmarkInsert(1, b)
}

func BenchmarkInsert1KB(b *testing.B) {
	_BenchmarkInsert(KB, b)
}
func BenchmarkInsert1MB(b *testing.B) {
	_BenchmarkInsert(MB, b)
}
func BenchmarkInsert1GB(b *testing.B) {
	_BenchmarkInsert(GB, b)
}

func _BenchmarkTraffic(fileSize, totalSize int64, b *testing.B) {
	commit(b)
	for i := 0; i < b.N; i++ {
		traffic("", fileSize, totalSize, b)
		commit(b)
	}
	commit(b)
}

func Benchmark_1_GB_x_1_MB(b *testing.B) {
	_BenchmarkTraffic(MB, GB, b)
}

func Benchmark_1_GB_x_10_MB(b *testing.B) {
	_BenchmarkTraffic(10*MB, GB, b)
}

func Benchmark_1_GB_x_100_MB(b *testing.B) {
	_BenchmarkTraffic(100*MB, GB, b)
}

func Benchmark_10_GB_x_500_MB(b *testing.B) {
	_BenchmarkTraffic(500*MB, 10*GB, b)
}
