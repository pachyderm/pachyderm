package pjs

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	storagesrv "github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"testing"
)

func TestCreateJob(t *testing.T) {
	t.Run("valid/parent/nil", func(t *testing.T) {
		c, fc := setupTest(t)
		resp := createJob(t, c, fc)
		t.Log(resp)
	})
}

func setupTest(t testing.TB) (pjs.APIClient, storage.FilesetClient) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return NewTestClient(t, db), storagesrv.NewTestFilesetClient(t, db)
}

func createFileSet(t *testing.T, fc storage.FilesetClient, files map[string][]byte) string {
	ctx := pctx.TestContext(t)
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	createFileClient, err := fc.CreateFileset(ctx)
	require.NoError(t, err)
	for path, data := range files {
		require.NoError(t, createFileClient.Send(&storage.CreateFilesetRequest{
			Modification: &storage.CreateFilesetRequest_AppendFile{
				AppendFile: &storage.AppendFile{
					Path: path,
					Data: &wrapperspb.BytesValue{
						Value: data,
					},
				},
			},
		}))
	}
	resp, err := createFileClient.CloseAndRecv()
	require.NoError(t, err)
	return resp.FilesetId
}

func createJob(t *testing.T, c pjs.APIClient, fc storage.FilesetClient) *pjs.CreateJobResponse {
	ctx := pctx.TestContext(t)
	programFileset := createFileSet(t, fc, map[string][]byte{
		"file": []byte(`!#/bin/bash; ls /input/;`),
	})
	inputFileset := createFileSet(t, fc, map[string][]byte{
		"a.txt": []byte(`dummy`),
	})
	resp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
		Program: programFileset,
		Input:   []string{inputFileset},
	})
	require.NoError(t, err)
	return resp
}
