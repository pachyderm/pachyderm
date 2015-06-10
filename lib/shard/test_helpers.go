package shard

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pachyderm/pfs/lib/traffic"
)

func check(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func checkResp(res *http.Response, expected string, t *testing.T) {
	if res.StatusCode != 200 {
		debug.PrintStack()
		t.Fatalf("Got error status: %s", res.Status)
	}
	value, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	check(err, t)
	if string(value) != expected {
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

func runOp(url string, o traffic.Op, t *testing.T) {
	switch {
	case o.Object == traffic.File && o.RW == traffic.W:
		writeFile(url, o.Path, o.Branch, o.Data, t)
	case o.Object == traffic.File && o.RW == traffic.R:
		checkFile(url, o.Path, o.Commit, o.Data, t)
	case o.Object == traffic.Commit:
		commit(url, o.Commit, o.Branch, t)
	case o.Object == traffic.Branch:
		branch(url, o.Commit, o.Branch, t)
	default:
		t.Fatal("Unrecognized op.")
	}
}

func runWorkload(url string, w traffic.Workload, t *testing.T) {
	for _, o := range w {
		runOp(url, o, t)
	}
}
