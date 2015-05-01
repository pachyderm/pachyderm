package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func checkBody(res *http.Response, expected string, t *testing.T) {
	value, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	check(err, t)
	if string(value) != expected {
		t.Fatalf("Got: %s\nExpected: %s\n", value, expected)
	}
}

func writeFile(url, name, branch, data string, t *testing.T) {
	res, err := http.Post(url+path.Join("/file", name)+"?branch="+branch, "application/text", strings.NewReader(data))
	check(err, t)
	checkBody(res, fmt.Sprintf("Created %s, size: %d.\n", name, len(data)), t)
}

func checkFile(url, name, commit, data string, t *testing.T) {
	res, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	check(err, t)
	checkBody(res, data, t)
}

func TestPing(t *testing.T) {
	shard := NewShard("TestPingData", "TestPingComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Get(s.URL + "/ping")
	check(err, t)
	checkBody(res, "pong\n", t)
	res.Body.Close()
}

func TestCommit(t *testing.T) {
	shard := NewShard("TestCommit", "TestPingComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	writeFile(s.URL, "foo", "master", "foo", t)
	checkFile(s.URL, "foo", "master", "foo", t)

	res, err := http.Post(s.URL+"/commit?commit=commit1", "", nil)
	check(err, t)
	checkBody(res, "commit1\n", t)
}
