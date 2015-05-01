package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"runtime/debug"
	"strings"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func checkResp(res *http.Response, expected string, t *testing.T) {
	if res.StatusCode != 200 {
		t.Fatalf("Got error status: %s", res.Status)
	}
	value, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	check(err, t)
	match, err := regexp.Match(expected, value)
	check(err, t)
	if match != true {
		t.Fatalf("Body:\n%s\ndidn't match:\n%s\n", string(value), expected)
	}
}

func writeFile(url, name, branch, data string, t *testing.T) {
	res, err := http.Post(url+path.Join("/file", name)+"?branch="+branch, "application/text", strings.NewReader(data))
	check(err, t)
	checkResp(res, fmt.Sprintf("Created %s, size: %d.\n", name, len(data)), t)
}

func checkFile(url, name, commit, data string, t *testing.T) {
	res, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	check(err, t)
	checkResp(res, data, t)
}

func checkNoFile(url, name, commit string, t *testing.T) {
	res, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	check(err, t)
	if res.StatusCode != 404 {
		debug.PrintStack()
		t.Fatalf("File: %s at commit: %s should have returned 404 but returned %s.", name, commit, res.Status)
	}
}

func TestPing(t *testing.T) {
	shard := NewShard("TestPingData", "TestPingComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	res, err := http.Get(s.URL + "/ping")
	check(err, t)
	checkResp(res, "pong\n", t)
	res.Body.Close()
}

func TestCommit(t *testing.T) {
	shard := NewShard("TestCommit", "TestPingComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	checkNoFile(s.URL, "file1", "master", t)
	writeFile(s.URL, "file1", "master", "file1", t)
	checkFile(s.URL, "file1", "master", "file1", t)

	res, err := http.Post(s.URL+"/commit?commit=commit1", "", nil)
	check(err, t)
	checkResp(res, "commit1\n", t)

	checkNoFile(s.URL, "file2", "master", t)
	writeFile(s.URL, "file2", "master", "file2", t)
	checkFile(s.URL, "file1", "master", "file1", t)
	checkFile(s.URL, "file2", "master", "file2", t)
	checkFile(s.URL, "file1", "commit1", "file1", t)
	checkNoFile(s.URL, "file2", "commit1", t)

	res, err = http.Get(s.URL + "/commit")
	check(err, t)
	checkResp(res, "commit1 - .*\n", t)
}

func TestBranch(t *testing.T) {
	shard := NewShard("TestBranch", "TestBranchComp", 0, 1)
	check(shard.EnsureRepos(), t)
	s := httptest.NewServer(shard.ShardMux())
	defer s.Close()

	checkNoFile(s.URL, "file1", "master", t)
	writeFile(s.URL, "file1", "master", "file1", t)
	checkFile(s.URL, "file1", "master", "file1", t)

	res, err := http.Post(s.URL+"/commit?commit=commit1", "", nil)
	check(err, t)
	checkResp(res, "commit1\n", t)
	checkFile(s.URL, "file1", "commit1", "file1", t)

	res, err = http.Post(s.URL+"/branch?branch=branch1&commit=commit1", "", nil)
	check(err, t)
	checkFile(s.URL, "file1", "branch1", "file1", t)

	writeFile(s.URL, "file2", "branch1", "file2", t)
	checkFile(s.URL, "file2", "branch1", "file2", t)
	res, err = http.Post(s.URL+"/commit?commit=commit2&branch=branch1", "", nil)
	check(err, t)
	checkResp(res, "commit2\n", t)
	checkFile(s.URL, "file2", "commit2", "file2", t)
	checkFile(s.URL, "file1", "commit2", "file1", t)

	res, err = http.Get(s.URL + "/commit")
	check(err, t)
	checkResp(res, "commit2 - .*\ncommit1 - .*\n", t)

	res, err = http.Get(s.URL + "/branch")
	check(err, t)
	checkResp(res, "branch1 - .*\nmaster - .*\n", t)
}
