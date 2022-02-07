package testing

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// TestCheckStorage checks that the CheckStorage rpc is wired up correctly.
// An more extensive test lives in the `chunk` package.
func TestCheckStorage(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	client := newClient(t)
	res, err := client.CheckStorage(ctx, &pfs.CheckStorageRequest{
		ReadChunkData: false,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestReadFileSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	client := newClient(t)
	fileContents := "test data"
	fsID, err := createFileSet(ctx, client, map[string][]byte{
		"file1.txt": []byte(fileContents),
		"file2.txt": []byte(fileContents),
	})
	require.NoError(t, err)

	rfsClient, err := client.ReadFileSet(ctx, &pfs.ReadFileSetRequest{
		FilesetId:       fsID,
		IncludeContents: true,
		Filters: []*pfs.FileSetFilter{
			{
				Value: &pfs.FileSetFilter_PathRegexp{PathRegexp: "/file2.txt"},
			},
		},
	})
	require.NoError(t, err)
	res, err := rfsClient.Recv()
	require.NoError(t, err)
	require.Equal(t, &pfs.ReadFileSetResponse{
		Path: res.Path,
	}, res)
	res, err = rfsClient.Recv()
	require.NoError(t, err)
	require.Equal(t, &pfs.ReadFileSetResponse{
		Path: res.Path,
		Data: []byte(fileContents),
	}, res)
}

func newClient(t testing.TB) pfs.APIClient {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	return env.PachClient.PfsAPIClient
}

func createFileSet(ctx context.Context, client pfs.APIClient, spec map[string][]byte) (string, error) {
	cfs, err := client.CreateFileSet(ctx)
	if err != nil {
		return "", err
	}
	// TODO(brendon): maybe sort?
	for k, v := range spec {
		err := cfs.Send(&pfs.ModifyFileRequest{
			Body: &pfs.ModifyFileRequest_AddFile{
				AddFile: &pfs.AddFile{
					Path: k,
					Source: &pfs.AddFile_Raw{
						Raw: &types.BytesValue{Value: v},
					},
				},
			},
		})
		if err != nil {
			return "", err
		}
	}
	res, err := cfs.CloseAndRecv()
	if err != nil {
		return "", err
	}
	return res.FileSetId, nil
}
