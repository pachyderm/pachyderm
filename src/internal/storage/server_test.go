package storage_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TODO: Update
//func TestServer(t *testing.T) {
//	ctx := pctx.TestContext(t)
//	db := dockertestenv.NewTestDB(t)
//	b, buckURL := dockertestenv.NewTestBucket(ctx, t)
//	t.Log("bucket", buckURL)
//
//	tracker := track.NewTestTracker(t, db)
//	fileset.NewTestStorage(ctx, t, db, tracker)
//
//	var config pachconfig.StorageConfiguration
//	s, err := New(Env{
//		DB:     db,
//		Bucket: b,
//		Config: config,
//	})
//	require.NoError(t, err)
//
//	w := s.Filesets.NewWriter(ctx)
//	require.NoError(t, w.Add("test.txt", "", strings.NewReader("hello world")))
//	id, err := w.Close()
//	require.NoError(t, err)
//	t.Log(id)
//
//	require.NoError(t, stream.ForEach(ctx, obj.NewKeyIterator(b, ""), func(k string) error {
//		if !strings.HasPrefix(k, chunkPrefix) {
//			t.Errorf("object key %q does not have prefix %q", k, chunkPrefix)
//		}
//		return nil
//	}))
//}

func createFileset(ctx context.Context, c storage.FilesetClient) (string, error) {
	cfc, err := c.CreateFileset(ctx)
	if err != nil {
		return "", err
	}
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("%02v", i)
		if err := cfc.Send(&storage.CreateFilesetRequest{
			Modification: &storage.CreateFilesetRequest_AppendFile{
				AppendFile: &storage.AppendFile{
					Path: fmt.Sprintf("/%02v", i),
					Data: wrapperspb.Bytes([]byte(path)),
				},
			},
		}); err != nil {
			return "", err
		}
	}
	response, err := cfc.CloseAndRecv()
	if err != nil {
		return "", err
	}
	return response.FilesetId, nil
}

func readFileset(ctx context.Context, c storage.FilesetClient, id string) ([]string, error) {
	rfc, err := c.ReadFileset(ctx, &storage.ReadFilesetRequest{
		FilesetId: id,
	})
	if err != nil {
		return nil, err
	}
	var files []string
	for {
		msg, err := rfc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		files = append(files, msg.Path)
	}
	return files, nil
}

func TestCreateAndRead(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient.FilesetClient
	id, err := createFileset(ctx, c)
	require.NoError(t, err)
	t.Run("Full", func(t *testing.T) {
		files, err := readFileset(ctx, c, id)
		require.NoError(t, err)
		require.Equal(t, 100, len(files))
	})
}
