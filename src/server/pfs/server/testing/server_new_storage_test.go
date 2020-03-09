package testing

import (
	"archive/tar"
	"bytes"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestCompaction(t *testing.T) {
	config := &serviceenv.PachdFullConfiguration{}
	config.NewStorageLayer = true
	config.StorageMemoryThreshold = 20
	config.StorageShardThreshold = 20
	config.StorageLevelZeroSize = 10
	testutil.WithRealEnv(func(env *testutil.RealEnv) error {
		c := env.PachClient
		repo := "test"
		branch := "master"
		filePrefix := "/file"
		require.NoError(t, c.CreateRepo(repo))
		var commit *pfs.Commit
		t.Run("PutTar", func(t *testing.T) {
			var err error
			for i := 0; i < 10; i++ {
				commit, err = c.StartCommit(repo, branch)
				require.NoError(t, err)
				buf := &bytes.Buffer{}
				tw := tar.NewWriter(buf)
				// Create files.
				for j := 0; j < 10; j++ {
					s := strconv.Itoa(i*10 + j)
					hdr := &tar.Header{
						Name: filePrefix + s,
						Size: int64(len(s)),
					}
					require.NoError(t, tw.WriteHeader(hdr))
					_, err := io.Copy(tw, strings.NewReader(s))
					require.NoError(t, err)
					require.NoError(t, tw.Flush())
				}
				require.NoError(t, tw.Close())
				require.NoError(t, c.PutTar(repo, commit.ID, buf))
				require.NoError(t, c.FinishCommit(repo, commit.ID))
			}
		})
		getTarContent := func(r io.Reader) string {
			tr := tar.NewReader(r)
			_, err := tr.Next()
			require.NoError(t, err)
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, tr)
			require.NoError(t, err)
			return buf.String()
		}
		t.Run("GetTar", func(t *testing.T) {
			r, err := c.GetTar(repo, commit.ID, filePrefix+"0")
			require.NoError(t, err)
			require.Equal(t, "0", getTarContent(r))
			r, err = c.GetTar(repo, commit.ID, filePrefix+"50")
			require.NoError(t, err)
			require.Equal(t, "50", getTarContent(r))
			r, err = c.GetTar(repo, commit.ID, filePrefix+"99")
			require.NoError(t, err)
			require.Equal(t, "99", getTarContent(r))
		})
		t.Run("GetTarConditional", func(t *testing.T) {
			downloadProb := 0.25
			require.NoError(t, c.GetTarConditional(repo, commit.ID, filePrefix, func(fileInfo *pfs.FileInfoNewStorage, r io.Reader) error {
				if rand.Float64() < downloadProb {
					require.Equal(t, strings.TrimPrefix(fileInfo.File.Path, filePrefix), getTarContent(r))
				}
				return nil
			}))
		})
		return nil
	}, config)
}
