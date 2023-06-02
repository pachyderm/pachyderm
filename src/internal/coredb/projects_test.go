package coredb

import (
	"github.com/gogo/protobuf/types"
	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"testing"
)

func TestCreateProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	require.NoError(t, v2_7_0.SetupCore(ctx, tx), "should be able to set up core tables")
	info := &pfs.ProjectInfo{Project: &pfs.Project{Name: "testproj"}, Description: "this is a test project", CreatedAt: types.TimestampNow()}
	require.NoError(t, CreateProject(ctx, tx, info))
	info2, err := GetProject(ctx, tx, "testproj")
	require.NoError(t, err, "should be able to create a tx")
}
