package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/traffic"
	"github.com/stretchr/testify/require"
)

var (
	fakeTestingTInstance = &fakeTestingT{}
)

func RunWorkload(t require.TestingT, url string, w traffic.Workload) {
	if t == nil {
		t = fakeTestingTInstance
	}
	for _, o := range w {
		runOp(t, url, o)
	}
}

func checkAndCloseHTTPResponseBody(t require.TestingT, response *http.Response, expected string) {
	data, err := ioutil.ReadAll(response.Body)
	require.NoError(t, response.Body.Close())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode, string(data))
	require.Equal(t, expected, string(data))
}

func checkWriteFile(t require.TestingT, url, name, branch, data string) {
	response, err := http.Post(url+path.Join("/file", name)+"?branch="+branch, "application/text", strings.NewReader(data))
	require.NoError(t, err)
	checkAndCloseHTTPResponseBody(t, response, "")
}

func checkFile(t require.TestingT, url, name, commit, data string) {
	response, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	require.NoError(t, err)
	checkAndCloseHTTPResponseBody(t, response, data)
}

func checkNoFile(t require.TestingT, url, name, commit string) {
	response, err := http.Get(url + path.Join("/file", name) + "?commit=" + commit)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, response.Status)
}

func commit(t require.TestingT, url, commit, branch string) {
	_url := fmt.Sprintf("%s/commit?branch=%s&commit=%s", url, branch, commit)
	response, err := http.Post(_url, "", nil)
	require.NoError(t, err)
	checkAndCloseHTTPResponseBody(t, response, fmt.Sprintf("%s\n", commit))
}

func branch(t require.TestingT, url, commit, branch string) {
	_url := fmt.Sprintf("%s/branch?branch=%s&commit=%s", url, branch, commit)
	response, err := http.Post(_url, "", nil)
	require.NoError(t, err)
	checkAndCloseHTTPResponseBody(t, response, fmt.Sprintf("%s\n", branch))
}

func runOp(t require.TestingT, url string, o traffic.Op) {
	switch {
	case o.Object == traffic.File && o.RW == traffic.W:
		checkWriteFile(t, url, o.Path, o.Branch, o.Data)
	case o.Object == traffic.File && o.RW == traffic.R:
		checkFile(t, url, o.Path, o.Commit, o.Data)
	case o.Object == traffic.Commit:
		commit(t, url, o.Commit, o.Branch)
	case o.Object == traffic.Branch:
		branch(t, url, o.Commit, o.Branch)
	default:
		require.FailNow(t, "Unrecognized op.")
	}
}

// TODO(pedge): remove when testing.T is removed from these methods, we need to keep
// this around because of the dependency in router which passes a nil value
type fakeTestingT struct{}

func (f *fakeTestingT) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (f *fakeTestingT) FailNow() {
}
