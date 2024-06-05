package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/cdr"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
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
	s, err := storageserver.New(ctx, storageserver.Env{
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

func TestCreateAndRead(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient.FilesetClient
	id, testFiles, err := createFileset(ctx, c, 99, units.KB)
	require.NoError(t, err)
	t.Run("Full", func(t *testing.T) {
		request := &storage.ReadFilesetRequest{FilesetId: id}
		checkFileset(ctx, t, c, request, testFiles)
	})
	t.Run("PathRange", func(t *testing.T) {
		for _, pr := range []*storage.PathRange{
			{
				Lower: "",
				Upper: "/33",
			},
			{
				Lower: "/33",
				Upper: "/66",
			},
			{
				Lower: "/66",
				Upper: "",
			},
		} {
			filters := []*storage.FileFilter{
				{
					Filter: &storage.FileFilter_PathRange{
						PathRange: pr,
					},
				},
			}
			request := &storage.ReadFilesetRequest{
				FilesetId: id,
				Filters:   filters,
			}
			checkFiles := applyFilters(t, testFiles, filters)
			require.Equal(t, 33, len(checkFiles))
			checkFileset(ctx, t, c, request, checkFiles)
		}
	})
	t.Run("PathRangeIntersection", func(t *testing.T) {
		for _, prs := range [][]*storage.PathRange{
			{
				{
					Lower: "",
					Upper: "/66",
				},
				{
					Lower: "/33",
					Upper: "",
				},
			},
			{
				{
					Lower: "/33",
					Upper: "",
				},
				{
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
			checkFiles := applyFilters(t, testFiles, filters)
			require.Equal(t, 33, len(checkFiles))
			checkFileset(ctx, t, c, request, checkFiles)
		}
	})
	t.Run("PathRangeDisjoint", func(t *testing.T) {
		for _, prs := range [][]*storage.PathRange{
			{
				{
					Lower: "",
					Upper: "/50",
				},
				{
					Lower: "/50",
					Upper: "",
				},
			},
			{
				{
					Lower: "/50",
					Upper: "",
				},
				{
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
			checkFiles := applyFilters(t, testFiles, filters)
			require.Equal(t, 0, len(checkFiles))
			checkFileset(ctx, t, c, request, checkFiles)
		}
	})
	t.Run("PathRegex", func(t *testing.T) {
		filters := []*storage.FileFilter{
			{
				Filter: &storage.FileFilter_PathRegex{
					PathRegex: "/.0",
				},
			},
		}
		request := &storage.ReadFilesetRequest{
			FilesetId: id,
			Filters:   filters,
		}
		checkFiles := applyFilters(t, testFiles, filters)
		require.Equal(t, 10, len(checkFiles))
		checkFileset(ctx, t, c, request, checkFiles)
	})
	t.Run("PathRangeAndPathRegex", func(t *testing.T) {
		filters := []*storage.FileFilter{
			{
				Filter: &storage.FileFilter_PathRange{
					PathRange: &storage.PathRange{
						Lower: "/33",
						Upper: "/66",
					},
				},
			},
			{
				Filter: &storage.FileFilter_PathRegex{
					PathRegex: "/.0",
				},
			},
		}
		request := &storage.ReadFilesetRequest{
			FilesetId: id,
			Filters:   filters,
		}
		checkFiles := applyFilters(t, testFiles, filters)
		require.Equal(t, 3, len(checkFiles))
		checkFileset(ctx, t, c, request, checkFiles)
	})
}

type testFile struct {
	path string
	data []byte
}

func createFileset(ctx context.Context, c storage.FilesetClient, num, size int) (string, []*testFile, error) {
	seed := int64(1648577872380609229)
	random := rand.New(rand.NewSource(seed))
	cfc, err := c.CreateFileset(ctx)
	if err != nil {
		return "", nil, err
	}
	var testFiles []*testFile
	padding := len(strconv.Itoa(num))
	for i := 0; i < num; i++ {
		tf := &testFile{
			path: fmt.Sprintf("/%0"+strconv.Itoa(padding)+"v", i),
			data: randutil.Bytes(random, size),
		}
		for _, c := range chunk(tf.data) {
			if err := cfc.Send(&storage.CreateFilesetRequest{
				Modification: &storage.CreateFilesetRequest_AppendFile{
					AppendFile: &storage.AppendFile{
						Path: tf.path,
						Data: wrapperspb.Bytes(c),
					},
				},
			}); err != nil {
				return "", nil, err
			}
		}
		testFiles = append(testFiles, tf)
	}
	response, err := cfc.CloseAndRecv()
	if err != nil {
		return "", nil, err
	}
	return response.FilesetId, testFiles, nil
}

// TODO: Why do messages need to be smaller than grpcutil.MaxMsgSize?
func chunk(data []byte) [][]byte {
	chunkSize := 4 * units.MB
	var result [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}

func applyFilters(t *testing.T, testFiles []*testFile, filters []*storage.FileFilter) []*testFile {
	var result []*testFile
Loop:
	for _, tf := range testFiles {
		for _, f := range filters {
			switch f := f.Filter.(type) {
			case *storage.FileFilter_PathRange:
				if tf.path < f.PathRange.Lower || f.PathRange.Upper != "" && tf.path >= f.PathRange.Upper {
					continue Loop
				}
			case *storage.FileFilter_PathRegex:
				r, err := regexp.Compile(f.PathRegex)
				require.NoError(t, err)
				if !r.MatchString(tf.path) {
					continue Loop
				}
			}
		}
		result = append(result, tf)
	}
	return result
}

func checkFileset(ctx context.Context, t *testing.T, c storage.FilesetClient, request *storage.ReadFilesetRequest, expected []*testFile) {
	rfc, err := c.ReadFileset(ctx, request)
	require.NoError(t, err)
	for {
		msg, err := rfc.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		require.True(t, len(expected) > 0)
		next := expected[0]
		require.Equal(t, next.path, msg.Path)
		require.True(t, bytes.Equal(next.data, msg.Data.Value))
		expected = expected[1:]
	}
	require.Equal(t, 0, len(expected))
}

func TestReadFilesetCDR(t *testing.T) {
	pachClient := pachd.NewTestPachd(t)
	ctx := pachClient.Ctx()
	c := pachClient.FilesetClient
	tests := []struct {
		name string
		num  int
		size int
	}{
		// TODO: Implement signed url caching, then enable these tests.
		//{"0B", 1000000, 0},
		//{"1KB", 100000, units.KB},
		{"100KB", 1000, 100 * units.KB},
		{"10MB", 10, 10 * units.MB},
		{"100MB", 1, 100 * units.MB},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, testFiles, err := createFileset(ctx, c, test.num, test.size)
			require.NoError(t, err)
			request := &storage.ReadFilesetRequest{FilesetId: id}
			checkFilesetCDR(ctx, t, c, request, testFiles)
		})
	}
}

func checkFilesetCDR(ctx context.Context, t *testing.T, c storage.FilesetClient, request *storage.ReadFilesetRequest, expected []*testFile) {
	rfc, err := c.ReadFilesetCDR(ctx, request)
	require.NoError(t, err)
	r := cdr.NewResolver()
	for {
		msg, err := rfc.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		next := expected[0]
		require.Equal(t, next.path, msg.Path)
		rc, err := r.Deref(ctx, msg.Ref)
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.True(t, bytes.Equal(next.data, data))
		expected = expected[1:]
	}
	require.Equal(t, 0, len(expected))
}
