package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// newTestClient is used internally for testing
func newTestClient(t testing.TB) pfs.APIClient {
	db := dockertestenv.NewTestDB(t)
	etcdEnv := testetcd.NewEnv(t)
	return NewTestClient(t, db, etcdEnv.EtcdClient)
}

func TestExample(t *testing.T) {
	newTestClient(t)
	// write the test here, don't do any cleanup, assume a clean environment.
}
