package server

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"golang.org/x/sync/errgroup"
)

const (
	KB               = 1024
	MB               = 1024 * 1024
	lettersAndSpaces = "abcdefghijklmnopqrstuvwxyz      "
)

type CountWriter struct {
	count int64
}

func (w *CountWriter) Write(p []byte) (int, error) {
	w.count += int64(len(p))
	return len(p), nil
}

func randBytes(size int) io.Reader {
	buf := make([]byte, size)
	rand.Read(buf)
	return bytes.NewReader(buf)
}

func BenchmarkPachyderm(b *testing.B) {
	repo := "BenchmarkPachyderm" + uuid.NewWithoutDashes()[0:12]
	c, err := client.NewFromAddress("127.0.0.1:30650")
	require.NoError(b, err)
	require.NoError(b, c.CreateRepo(repo))
	nFiles := 1000000

	commit, err := c.StartCommit(repo, "")
	require.NoError(b, err)
	if !b.Run(fmt.Sprintf("Put%dFiles", nFiles), func(b *testing.B) {
		var eg errgroup.Group
		for k := 0; k < nFiles; k++ {
			k := k
			eg.Go(func() error {
				_, err := c.PutFile(repo, commit.ID, fmt.Sprintf("file%d", k), randBytes(KB))
				return err
			})
		}
		b.SetBytes(int64(nFiles * KB))
		require.NoError(b, eg.Wait())
	}) {
		return
	}
	require.NoError(b, c.FinishCommit(repo, commit.ID))
	require.NoError(b, c.SetBranch(repo, commit.ID, "master"))

	if !b.Run(fmt.Sprintf("Get%dFiles", nFiles), func(b *testing.B) {
		var eg errgroup.Group
		w := &CountWriter{}
		defer func() { b.SetBytes(w.count) }()
		for k := 0; k < nFiles; k++ {
			k := k
			eg.Go(func() error {
				return c.GetFile(repo, commit.ID, fmt.Sprintf("file%d", k), 0, 0, w)
			})
		}
		require.NoError(b, eg.Wait())
	}) {
		return
	}

	if !b.Run(fmt.Sprintf("PipelineCopy%dFiles", nFiles), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pipeline := "BenchmarkPachydermPipeline" + uuid.NewWithoutDashes()[0:12]
			require.NoError(b, c.CreatePipeline(
				pipeline,
				"",
				[]string{"bash"},
				[]string{fmt.Sprintf("cp -R %s /pfs/out/", path.Join("/pfs", repo, "/*"))},
				&ppsclient.ParallelismSpec{
					Strategy: ppsclient.ParallelismSpec_CONSTANT,
					Constant: 4,
				},
				[]*ppsclient.PipelineInput{{
					Repo: client.NewRepo(repo),
					Glob: "/*",
				}},
				"",
				false,
			))
			_, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(repo, commit.ID)}, nil)
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
