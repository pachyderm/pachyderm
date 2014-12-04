package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
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

var reader ConstReader

func timeParam(r *http.Request) string {
	if p := r.URL.Query().Get("time"); p != "" {
		return p
	}
	return "30"
}

// insert inserts a single file in to the filesystem
func insert(t *testing.T) {
	url := "http://172.17.42.1/pfs/" + randSeq(10)
	resp, err := http.Post(url, "application/text", io.LimitReader(reader, docSize))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "200" {
		t.Error(resp)
	}
}

func traffic(t *testing.T) {
	workers := 3
	var wg sync.WaitGroup
	wg.Add(workers)
	startTime := time.Now()
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for time.Since(startTime) < (10 * time.Second) {
				insert(t)
			}
		}()
	}
	wg.Wait()
}

func commit(t *testing.T) {
	resp, err := http.Post("http://172.17.42.1/commit", "application/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "200" {
		t.Error(resp)
	}
}

func TestSmoke(t *testing.T) {
	for i := 0; i < 20; i++ {
		traffic(t)
		commit(t)
	}
}
