package dockertestenv

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// TestDockerService check if the docker service is running and can be pinged.
//
// You can run it alone with `go test -run=TestDockerService` to debug your connection to Docker.
func TestDockerService(t *testing.T) {
	ctx := pctx.TestContext(t)
	dclient, err := newDockerClient(ctx)
	require.NoError(t, err)
	defer dclient.Close()
}
