package testpachd

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

// MockEnv contains the basic setup for running end-to-end pachyderm tests
// entirely locally within the test process. It provides a temporary directory
// for storing data, an embedded etcd server with a connected client, as well as
// a local mock pachd instance which allows a test to hook into any pachd calls.
type MockEnv struct {
	testetcd.Env
	MockPachd  *MockPachd
	PachClient *client.APIClient
}

// NewMockEnv constructs a MockEnv for testing, which will be destroyed at the
// end of the test.
func NewMockEnv(t testing.TB, options ...InterceptorOption) *MockEnv {
	etcdEnv := testetcd.NewEnv(t)

	// Use an error group with a cancelable context to supervise every component
	// and cancel everything if one fails
	ctx, cancel := context.WithCancel(etcdEnv.Context)
	eg, ctx := errgroup.WithContext(ctx)
	t.Cleanup(func() {
		require.NoError(t, eg.Wait())
	})
	t.Cleanup(cancel)

	mockEnv := &MockEnv{Env: *etcdEnv}
	mockEnv.Context = ctx

	var err error
	mockEnv.MockPachd, err = NewMockPachd(mockEnv.Context, 0, options...)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mockEnv.MockPachd.Close())
	})

	eg.Go(func() error {
		return errorWait(ctx, mockEnv.MockPachd.Err())
	})

	mockEnv.PachClient, err = client.NewFromURI(mockEnv.MockPachd.Addr.String())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mockEnv.PachClient.Close())
	})

	// TODO: supervise the PachClient connection and error the errgroup if they
	// go down

	return mockEnv
}

func errorWait(ctx context.Context, errChan <-chan error) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
