package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/version"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestInspectCluster(t *testing.T) {
	var a apiServer
	_, err := a.InspectCluster(pctx.TestContext(t), &admin.InspectClusterRequest{
		ClientVersion:  version.Version,
		CurrentProject: &pfs.Project{Name: "#<does-not-exist>"},
	})
	require.NoError(t, err, "InspectCluster must not err")
}
