package shard

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/traffic"
)

func Check(err error, t *testing.T) {
	if err != nil {
		if t == nil {
			log.Print(err)
		} else {
			debug.PrintStack()
			t.Fatal(err)
		}
	}
}

func CheckResp(res *http.Response, expected string, t *testing.T) {
	if res.StatusCode != 200 {
		if t == nil {
			log.Printf("Got error status: %s", res.Status)
		} else {
			debug.PrintStack()
			t.Fatalf("Got error status: %s", res.Status)
		}
	}
	value, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	Check(err, t)
	if string(value) != expected {
		if t == nil {
			log.Printf("Body:\n%s\ndidn't match:\n%s\n", string(value), expected)
		} else {
			t.Fatalf("Body:\n%s\ndidn't match:\n%s\n", string(value), expected)
		}
	}
}

func WriteFile(url, name, branch, data string, t *testing.T) {
	res, err := http.Post(url+path.Join("/file", name)+"?branch="+branch, "application/text", strings.NewReader(data))
	Check(err, t)
	CheckResp(res, fmt.Sprintf("Created %s, size: %d.\n", name, len(data)), t)
}

func Checkfile(url, name, commit, data string, t *testing.T) {
	res, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	Check(err, t)
	CheckResp(res, data, t)
}

func CheckNoFile(url, name, commit string, t *testing.T) {
	res, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	Check(err, t)
	if res.StatusCode != 404 {
		debug.PrintStack()
		t.Fatalf("File: %s at commit: %s should have returned 404 but returned %s.", name, commit, res.Status)
	}
}

func Commit(url, commit, branch string, t *testing.T) {
	_url := fmt.Sprintf("%s/commit?branch=%s&commit=%s", url, branch, commit)
	res, err := http.Post(_url, "", nil)
	Check(err, t)
	CheckResp(res, fmt.Sprintf("%s\n", commit), t)
}

func Branch(url, commit, branch string, t *testing.T) {
	_url := fmt.Sprintf("%s/branch?branch=%s&commit=%s", url, branch, commit)
	res, err := http.Post(_url, "", nil)
	Check(err, t)
	CheckResp(res, fmt.Sprintf("Created branch. (%s) -> %s.\n", commit, branch), t)
}

func RunOp(url string, o traffic.Op, t *testing.T) {
	switch {
	case o.Object == traffic.File && o.RW == traffic.W:
		WriteFile(url, o.Path, o.Branch, o.Data, t)
	case o.Object == traffic.File && o.RW == traffic.R:
		Checkfile(url, o.Path, o.Commit, o.Data, t)
	case o.Object == traffic.Commit:
		Commit(url, o.Commit, o.Branch, t)
	case o.Object == traffic.Branch:
		Branch(url, o.Commit, o.Branch, t)
	default:
		t.Fatal("Unrecognized op.")
	}
}

func RunWorkload(url string, w traffic.Workload, t *testing.T) {
	for _, o := range w {
		RunOp(url, o, t)
	}
}
