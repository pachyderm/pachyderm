package main

import (
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var once sync.Once
var docSize int64 = 1 << 12

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

var KB int64 = 1 << 10
var MB int64 = 1 << 20
var GB int64 = 1 << 30
var TB int64 = 1 << 40
var PB int64 = 1 << 50

var reader ConstReader

// insert inserts a single file in to the filesystem
func insert(t testing.TB) {
	url := "http://172.17.42.1/pfs/" + randSeq(10)
	resp, err := http.Post(url, "application/text", io.LimitReader(reader, docSize))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}

func traffic(sizeLimit int64, t testing.TB) {
	workers := 3
	var wg sync.WaitGroup
	wg.Add(workers)
	var totalSize int64
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for atomic.AddInt64(&totalSize, docSize) <= sizeLimit {
				insert(t)
			}
		}()
	}
	wg.Wait()
}

func commit(t testing.TB) {
	resp, err := http.Post("http://172.17.42.1/commit", "application/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}

func TestSmoke(t *testing.T) {
	t.Log("TestSmoke")
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			insert(t)
		}
		commit(t)
	}
}

func TestFire(t *testing.T) {
	t.Log("TestFire")
	for i := 0; i < 5; i++ {
		traffic(5*MB, t)
		commit(t)
	}
}

func BenchmarkInsert(b *testing.B) {
	commit(b)
	for i := 0; i < b.N; i++ {
		insert(b)
		commit(b)
	}
	commit(b)
}

func Benchmark100MB(b *testing.B) {
	commit(b)
	for i := 0; i < b.N; i++ {
		traffic(100*MB, b)
		commit(b)
	}
	commit(b)
}
