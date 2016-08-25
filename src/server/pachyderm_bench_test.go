package server

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
	"golang.org/x/sync/errgroup"
)

const (
	nCommits = 10
	nFiles   = 10
	MB       = 1024 * 1024
)

type CountWriter struct {
	count int64
}

func (w *CountWriter) Write(p []byte) (int, error) {
	w.count += int64(len(p))
	return len(p), nil
}

func BenchmarkPutGetFile(b *testing.B) {
	repo := uniqueString("BenchmarkPutGetFile")
	c, err := client.NewInCluster()
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(repo))
	var commits []*pfsclient.Commit
	if !b.Run("PutFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < nCommits; j++ {
				commit, err := c.StartCommit(repo, "", "master")
				commits = append(commits, commit)
				require.NoError(b, err)
				var eg errgroup.Group
				for k := 0; k < nFiles; k++ {
					k := k
					eg.Go(func() error {
						rand := rand.New(rand.NewSource(int64(time.Now().UnixNano())))
						_, err := c.PutFile(repo, "master", fmt.Sprintf("file%d", k), workload.NewReader(rand, MB))
						b.SetBytes(MB)
						return err
					})
				}
				require.NoError(b, eg.Wait())
				require.NoError(b, c.FinishCommit(repo, "master"))
			}
		}
	}) {
		return
	}
	if !b.Run("GetFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var eg errgroup.Group
			for _, commit := range commits {
				commit := commit
				for k := 0; k < nFiles; k++ {
					k := k
					eg.Go(func() error {
						w := &CountWriter{}
						defer func() { b.SetBytes(w.count) }()
						return c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", k), 0, 0, "", false, nil, w)
					})
				}
			}
			require.NoError(b, eg.Wait())
		}
	}) {
		return
	}
}
