package fuse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func put(path string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost%s/%s", FuseServerPort, path), body)
	if err != nil {
		panic(err)
	}
	x, err := client.Do(req)
	return x, errors.EnsureStack(err)
}

func get(path string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost%s/%s", FuseServerPort, path), nil)
	if err != nil {
		panic(err)
	}
	x, err := client.Do(req)
	return x, errors.EnsureStack(err)
}

// TODO: pass reference to the MountManager object to the test func, so that the
// test can call MountBranch, UnmountBranch etc directly for convenience
func withServerMount(ctx context.Context, tb testing.TB, c *client.APIClient, sopts *ServerOptions, f func(mountPoint string)) {
	dir := tb.TempDir()
	if sopts == nil {
		sopts = &ServerOptions{
			MountDir: dir,
		}
	}
	if sopts.Unmount == nil {
		sopts.Unmount = make(chan struct{})
	}
	unmounted := make(chan struct{})
	var mountErr error
	defer func() {
		close(sopts.Unmount)
		<-unmounted
		require.ErrorIs(tb, mountErr, http.ErrServerClosed)
	}()
	defer func() {
		// panic() leaves the mount in a weird state that makes it hard to rerun the
		// tests (mostly relevent when you're iterating on these tests, or the code
		// they test).
		//
		// N.B. (msteffen): AFAICT, recover() returns nil unless called in the
		// _same_ goroutine as panic(). Go's HTTP server handles each request in a
		// separate goroutine, so if panic() is called inside a handler, this call
		// to recover() will return nil, and tb.Fatal() will not be called. In
		// general, I don't know of a way to determine, at this location, whether
		// the program is exiting due to a panic in another goro or exiting
		// normally.
		if r := recover(); r != nil {
			tb.Fatal(r)
		}
	}()
	go func() {
		mountErr = Serve(ctx, sopts, c)
		close(unmounted)
	}()
	// Gotta give the fuse mount time to come up.
	time.Sleep(2 * time.Second)
	f(dir)
}
