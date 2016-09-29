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
	nFiles = 100
	MB     = 1024 * 1024
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

	commit, err := c.StartCommit(repo, "master")
	require.NoError(b, err)
	if !b.Run(fmt.Sprintf("Put%dFiles", nFiles), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var eg errgroup.Group
			for k := 0; k < nFiles; k++ {
				k := k
				eg.Go(func() error {
					rand := rand.New(rand.NewSource(int64(time.Now().UnixNano())))
					_, err := c.PutFile(repo, "master", fmt.Sprintf("file%d", k), workload.NewReader(rand, MB))
					return err
				})
			}
			b.SetBytes(nFiles * MB)
			require.NoError(b, eg.Wait())
		}
	}) {
		return
	}
	require.NoError(b, c.FinishCommit(repo, "master"))

	if !b.Run(fmt.Sprintf("Get%dFiles", nFiles), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var eg errgroup.Group
			w := &CountWriter{}
			defer func() { b.SetBytes(w.count) }()
			for k := 0; k < nFiles; k++ {
				k := k
				eg.Go(func() error {
					return c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", k), 0, 0, "", false, nil, w)
				})
			}
			require.NoError(b, eg.Wait())
		}
	}) {
		return
	}
	if !b.Run(fmt.Sprintf("PipelineCopy%dFiles", nFiles), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pipeline := uniqueString("BenchmarkPachydermPipeline")
			require.NoError(b, c.CreatePipeline(
				pipeline,
				"",
				[]string{"cp", "-R", path.Join("/pfs", repo), "/pfs/out/"},
				nil,
				&ppsclient.ParallelismSpec{
					Strategy: ppsclient.ParallelismSpec_CONSTANT,
					Constant: 1,
				},
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
		}
	}) {
		return
	}
}
