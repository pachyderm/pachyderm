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
	"testing/quick"

	"github.com/pachyderm/pfs/lib/traffic"
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

func commit(url, commit, branch string, t *testing.T) {
	_url := fmt.Sprintf("%s/commit?branch=%s&commit=%s", url, branch, commit)
	res, err := http.Post(_url, "", nil)
	check(err, t)
	checkResp(res, fmt.Sprintf("%s\n", commit), t)
}

func branch(url, commit, branch string, t *testing.T) {
	_url := fmt.Sprintf("%s/branch?branch=%s&commit=%s", url, branch, commit)
	res, err := http.Post(_url, "", nil)
	check(err, t)
	checkResp(res, fmt.Sprintf("Created branch. (%s) -> %s.\n", commit, branch), t)
}

func runWorkload(url string, w traffic.Workload, t *testing.T) {
	t.Logf("Workload of size %d", len(w))
	for _, o := range w {
		t.Logf("Running %#v.", o)
		switch {
		case o.Object == traffic.File && o.RW == traffic.W:
			writeFile(url, o.Path, o.Branch, o.Data, t)
		case o.Object == traffic.File && o.RW == traffic.R:
			checkFile(url, o.Path, o.Commit, o.Data, t)
		case o.Object == traffic.Commit:
			commit(url, o.Commit, o.Branch, t)
		case o.Object == traffic.Branch:
			commit(url, o.Commit, o.Branch, t)
		default:
			t.Fatal("Unrecognized op.")
		}
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

func TestBasic(t *testing.T) {
	c := 0
	f := func(w traffic.Workload) bool {
		shard := NewShard(fmt.Sprintf("TestBasic%d", c), fmt.Sprintf("TestBasicComp%d", c), 0, 1)
		c++
		check(shard.EnsureRepos(), t)
		s := httptest.NewServer(shard.ShardMux())
		defer s.Close()

		runWorkload(s.URL, w, t)
		facts := w.Facts()
		runWorkload(s.URL, facts, t)
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 10}); err != nil {
		t.Error(err)
	}
}

func TestPull(t *testing.T) {
	//srcShard := NewShard("TestPullSrc", "TestPullSrcComp", 0, 1)
	//check(srcShard.EnsureRepos(), t)
	//src := httptest.NewServer(srcShard.ShardMux())
	//defer src.Close()

	//writeFile(s.URL, "file1", "master", "file1", t)

	//res, err := http.Post(s.URL+"/commit?commit=commit1", "", nil)
	//check(err, t)

	//res, err = http.Post(s.URL+"/branch?branch=branch1&commit=commit1", "", nil)
	//check(err, t)
	//writeFile(s.URL, "file2", "branch1", "file2", t)
	//res, err = http.Post(s.URL+"/commit?commit=commit2&branch=branch1", "", nil)
	//check(err, t)

	//shard := NewShard("TestPullSrc", "TestPullSrcComp", 0, 1)
}
