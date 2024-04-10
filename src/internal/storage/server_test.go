package storage_test

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	storageserver "github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestServer(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	b, buckURL := dockertestenv.NewTestBucket(ctx, t)
	t.Log("bucket", buckURL)

	tracker := track.NewTestTracker(t, db)
	fileset.NewTestStorage(ctx, t, db, tracker)

	var config pachconfig.StorageConfiguration
	s, err := storageserver.New(storageserver.Env{
		DB:     db,
		Bucket: b,
		Config: config,
	})
	require.NoError(t, err)

	w := s.Filesets.NewWriter(ctx)
	require.NoError(t, w.Add("test.txt", "", strings.NewReader("hello world")))
	id, err := w.Close()
	require.NoError(t, err)
	t.Log(id)

	require.NoError(t, stream.ForEach(ctx, obj.NewKeyIterator(b, ""), func(k string) error {
		if !strings.HasPrefix(k, storageserver.ChunkPrefix) {
			t.Errorf("object key %q does not have prefix %q", k, storageserver.ChunkPrefix)
		}
		return nil
	}))
}

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

func readFileset(ctx context.Context, c storage.FilesetClient, request *storage.ReadFilesetRequest) ([]string, error) {
	rfc, err := c.ReadFileset(ctx, request)
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
		request := &storage.ReadFilesetRequest{FilesetId: id}
		files, err := readFileset(ctx, c, request)
		require.NoError(t, err)
		require.Equal(t, 100, len(files))
	})
	t.Run("PathRange", func(t *testing.T) {
		var files []string
		for _, pr := range []*storage.PathRange{
			&storage.PathRange{
				Lower: "",
				Upper: "/33",
			},
			&storage.PathRange{
				Lower: "/33",
				Upper: "/66",
			},
			&storage.PathRange{
				Lower: "/66",
				Upper: "",
			},
		} {
			filters := []*storage.FileFilter{
				&storage.FileFilter{
					Filter: &storage.FileFilter_PathRange{
						PathRange: pr,
					},
				},
			}
			request := &storage.ReadFilesetRequest{
				FilesetId: id,
				Filters:   filters,
			}
			nextFiles, err := readFileset(ctx, c, request)
			require.NoError(t, err)
			for _, f := range nextFiles {
				validateFilters(t, f, filters)
			}
			files = append(files, nextFiles...)
		}
		require.Equal(t, 100, len(files))
	})
	t.Run("PathRangeIntersection", func(t *testing.T) {
		for _, prs := range [][]*storage.PathRange{
			[]*storage.PathRange{
				&storage.PathRange{
					Lower: "",
					Upper: "/66",
				},
				&storage.PathRange{
					Lower: "/33",
					Upper: "",
				},
			},
			[]*storage.PathRange{
				&storage.PathRange{
					Lower: "/33",
					Upper: "",
				},
				&storage.PathRange{
					Lower: "",
					Upper: "/66",
				},
			},
		} {
			var filters []*storage.FileFilter
			for _, pr := range prs {
				filters = append(filters, &storage.FileFilter{
					Filter: &storage.FileFilter_PathRange{
						PathRange: pr,
					},
				})
			}
			request := &storage.ReadFilesetRequest{
				FilesetId: id,
				Filters:   filters,
			}
			files, err := readFileset(ctx, c, request)
			require.NoError(t, err)
			for _, f := range files {
				validateFilters(t, f, filters)
			}
			require.Equal(t, 33, len(files))
		}
	})
	t.Run("PathRangeDisjoint", func(t *testing.T) {
		for _, prs := range [][]*storage.PathRange{
			[]*storage.PathRange{
				&storage.PathRange{
					Lower: "",
					Upper: "/50",
				},
				&storage.PathRange{
					Lower: "/50",
					Upper: "",
				},
			},
			[]*storage.PathRange{
				&storage.PathRange{
					Lower: "/50",
					Upper: "",
				},
				&storage.PathRange{
					Lower: "",
					Upper: "/50",
				},
			},
		} {
			var filters []*storage.FileFilter
			for _, pr := range prs {
				filters = append(filters, &storage.FileFilter{
					Filter: &storage.FileFilter_PathRange{
						PathRange: pr,
					},
				})
			}
			request := &storage.ReadFilesetRequest{
				FilesetId: id,
				Filters:   filters,
			}
			files, err := readFileset(ctx, c, request)
			require.NoError(t, err)
			for _, f := range files {
				validateFilters(t, f, filters)
			}
			require.Equal(t, 0, len(files))
		}
	})
	t.Run("PathRegex", func(t *testing.T) {
		filters := []*storage.FileFilter{
			&storage.FileFilter{
				Filter: &storage.FileFilter_PathRegex{
					PathRegex: "/.0",
				},
			},
		}
		request := &storage.ReadFilesetRequest{
			FilesetId: id,
			Filters:   filters,
		}
		files, err := readFileset(ctx, c, request)
		require.NoError(t, err)
		for _, f := range files {
			validateFilters(t, f, filters)
		}
		require.Equal(t, 10, len(files))
	})
	t.Run("PathRangeAndPathRegex", func(t *testing.T) {
		filters := []*storage.FileFilter{
			&storage.FileFilter{
				Filter: &storage.FileFilter_PathRange{
					PathRange: &storage.PathRange{
						Lower: "/33",
						Upper: "/66",
					},
				},
			},
			&storage.FileFilter{
				Filter: &storage.FileFilter_PathRegex{
					PathRegex: "/.0",
				},
			},
		}
		request := &storage.ReadFilesetRequest{
			FilesetId: id,
			Filters:   filters,
		}
		files, err := readFileset(ctx, c, request)
		require.NoError(t, err)
		for _, f := range files {
			validateFilters(t, f, filters)
		}
		require.Equal(t, 3, len(files))
	})
}

func validateFilters(t *testing.T, path string, filters []*storage.FileFilter) {
	for _, f := range filters {
		switch f := f.Filter.(type) {
		case *storage.FileFilter_PathRange:
			require.True(t, path >= f.PathRange.Lower && f.PathRange.Upper == "" || path < f.PathRange.Upper)
		case *storage.FileFilter_PathRegex:
			r, err := regexp.Compile(f.PathRegex)
			require.NoError(t, err)
			require.True(t, r.MatchString(path))
		}
	}
}
