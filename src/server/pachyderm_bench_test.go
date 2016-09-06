package server

import (
	"fmt"
	"math/rand"
	"path"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
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

func BenchmarkPachyderm(b *testing.B) {
	repo := uniqueString("BenchmarkPachyderm")
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
	if !b.Run("Pipeline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pipeline := uniqueString("BenchmarkPachydermPipeline")
			require.NoError(b, c.CreatePipeline(
				pipeline,
				"",
				[]string{"cp", "-R", path.Join("/pfs", repo), "/pfs/out/"},
				nil,
				1,
				[]*ppsclient.PipelineInput{{Repo: client.NewRepo(repo)}},
				false,
			))
			_, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(repo, "master")}, nil)
			require.NoError(b, err)
			b.StopTimer()
			repoInfo, err := c.InspectRepo(repo)
			require.NoError(b, err)
			b.SetBytes(int64(repoInfo.SizeBytes))
			repoInfo, err = c.InspectRepo(pipeline)
			require.NoError(b, err)
			b.SetBytes(int64(repoInfo.SizeBytes))
			b.StartTimer()
		}
	}) {
		return
	}
}
