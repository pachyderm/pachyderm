//go:build unit_test

package testing

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

func CommitToID(commit interface{}) interface{} {
	return pfsdb.CommitKey(commit.(*pfs.Commit))
}

func CommitInfoToID(commit interface{}) interface{} {
	return pfsdb.CommitKey(commit.(*pfs.CommitInfo).Commit)
}

func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func FileInfoToPath(fileInfo interface{}) interface{} {
	return fileInfo.(*pfs.FileInfo).File.Path
}

func finishCommit(pachClient *client.APIClient, repo, branch, id string) error {
	return finishProjectCommit(pachClient, pfs.DefaultProjectName, repo, branch, id)
}

func finishProjectCommit(pachClient *client.APIClient, project, repo, branch, id string) error {
	if err := pachClient.FinishCommit(project, repo, branch, id); err != nil {
		if !pfsserver.IsCommitFinishedErr(err) {
			return err
		}
	}
	_, err := pachClient.WaitCommit(project, repo, branch, id)
	return err
}

func assertMasterHeads(t *testing.T, c *client.APIClient, repoToCommitIDs map[string]string) {
	for repo, id := range repoToCommitIDs {
		info, err := c.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		if id != info.Commit.Id {
			fmt.Printf("for repo %q, expected: %q; actual: %q\n", repo, id, info.Commit.Id)
		}
		require.Equal(t, id, info.Commit.Id)
	}
}

// helper function for building DAGs by creating a repo with specified provenant repos
// this function sets up a master branch for each repo
func buildDAG(t *testing.T, c *client.APIClient, repo string, provenance ...string) {
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
	var provs []*pfs.Branch
	for _, p := range provenance {
		provs = append(provs, client.NewBranch(pfs.DefaultProjectName, p, "master"))
	}
	require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", provs))
	require.NoError(t, finishCommit(c, repo, "master", ""))
}

type fileSetSpec map[string]tarutil.File

func (fs fileSetSpec) makeTarStream() io.Reader {
	buf := &bytes.Buffer{}
	if err := tarutil.WithWriter(buf, func(tw *tar.Writer) error {
		for _, file := range fs {
			if err := tarutil.WriteFile(tw, file); err != nil {
				panic(err)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return buf
}

func finfosToPaths(finfos []*pfs.FileInfo) (paths []string) {
	for _, finfo := range finfos {
		paths = append(paths, finfo.File.Path)
	}
	return paths
}

func TestPFS(suite *testing.T) {
	suite.Parallel()

	suite.Run("WalkFileTest", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir1/file1.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir1/file1.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir2/file2.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir2/file2.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir3/file3.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir/dir3/file3.2", &bytes.Buffer{}))

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		// should list all files including directories and root
		var fis []*pfs.FileInfo
		require.NoError(t, env.PachClient.WalkFile(commit1, "/", func(fi *pfs.FileInfo) error {
			fis = append(fis, fi)
			return nil
		}))
		require.Equal(t, 11, len(fis))

		request := &pfs.WalkFileRequest{File: commit1.NewFile("/dir"), Number: 4}
		walkFileClient, err := env.PachClient.PfsAPIClient.WalkFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](walkFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 4, len(fis))
		require.ElementsEqual(t, []string{"/dir/", "/dir/dir1/", "/dir/dir1/file1.1", "/dir/dir1/file1.2"}, finfosToPaths(fis))

		lastFilePath := fis[3].File.Path
		request = &pfs.WalkFileRequest{File: commit1.NewFile("/dir"), PaginationMarker: commit1.NewFile(lastFilePath)}
		walkFileClient, err = env.PachClient.PfsAPIClient.WalkFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](walkFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 6, len(fis))

		request = &pfs.WalkFileRequest{File: commit1.NewFile("/dir"), Number: 2, Reverse: true}
		walkFileClient, err = env.PachClient.PfsAPIClient.WalkFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](walkFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 2, len(fis))
		require.ElementsEqual(t, []string{"/dir/dir3/file3.1", "/dir/dir3/file3.2"}, finfosToPaths(fis))
		require.Equal(t, true, fis[0].File.Path > fis[1].File.Path)

		request = &pfs.WalkFileRequest{File: commit1.NewFile("/dir"), Reverse: true}
		walkFileClient, err = env.PachClient.PfsAPIClient.WalkFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](walkFileClient, 1000)
		require.YesError(t, err)

		request = &pfs.WalkFileRequest{File: commit1.NewFile("/dir"), PaginationMarker: commit1.NewFile("/dir/dir1/file1.2"), Number: 3, Reverse: true}
		walkFileClient, err = env.PachClient.PfsAPIClient.WalkFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](walkFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 3, len(fis))
		require.ElementsEqual(t, []string{"/dir/dir1/file1.1", "/dir/dir1/", "/dir/"}, finfosToPaths(fis))

	})

	suite.Run("ListFileTest", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.5", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.2", &bytes.Buffer{}))

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		// should list a directory but not siblings
		var fis []*pfs.FileInfo
		require.NoError(t, env.PachClient.ListFile(commit1, "/dir1", func(fi *pfs.FileInfo) error {
			fis = append(fis, fi)
			return nil
		}))
		require.ElementsEqual(t, []string{"/dir1/file1.1", "/dir1/file1.2", "/dir1/file1.5"}, finfosToPaths(fis))
		// should list the root
		fis = nil
		require.NoError(t, env.PachClient.ListFile(commit1, "/", func(fi *pfs.FileInfo) error {
			fis = append(fis, fi)
			return nil
		}))
		require.ElementsEqual(t, []string{"/dir1/", "/dir2/"}, finfosToPaths(fis))

		request := &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), PaginationMarker: commit1.NewFile("/dir1/file1.2")}
		listFileClient, err := env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 1, len(fis))
		require.ElementsEqual(t, []string{"/dir1/file1.5"}, finfosToPaths(fis))

		request = &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), PaginationMarker: commit1.NewFile("/dir1/file1.1"), Number: 2}
		listFileClient, err = env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 2, len(fis))
		require.ElementsEqual(t, []string{"/dir1/file1.2", "/dir1/file1.5"}, finfosToPaths(fis))

		request = &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), Number: 1, Reverse: true}
		listFileClient, err = env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 1, len(fis))
		require.ElementsEqual(t, []string{"/dir1/file1.5"}, finfosToPaths(fis))

		request = &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), Reverse: true}
		listFileClient, err = env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.YesError(t, err)

		request = &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), PaginationMarker: commit1.NewFile("/dir1/file1.1"), Number: 2, Reverse: true}
		listFileClient, err = env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 0, len(fis))

		request = &pfs.ListFileRequest{File: commit1.NewFile("/dir1"), PaginationMarker: commit1.NewFile("/dir1/file1.5"), Number: 2, Reverse: true}
		listFileClient, err = env.PachClient.PfsAPIClient.ListFile(env.PachClient.Ctx(), request)
		require.NoError(t, err)
		fis, err = grpcutil.Collect[*pfs.FileInfo](listFileClient, 1000)
		require.NoError(t, err)
		require.Equal(t, 2, len(fis))
		require.ElementsEqual(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, finfosToPaths(fis))
		require.Equal(t, true, fis[0].File.Path > fis[1].File.Path)
	})

	suite.Run("ListCommitStartedTime", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		err := env.PachClient.CreateRepo(pfs.DefaultProjectName, "foo")
		require.NoError(t, err)
		// create three sequential commits
		commits := make([]*pfs.Commit, 3)
		for i := 0; i < 3; i++ {
			newCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "foo", "master")
			require.NoError(t, err)
			err = env.PachClient.FinishCommit(pfs.DefaultProjectName, "foo", "master", newCommit.Id)
			require.NoError(t, err)
			commits[i] = newCommit
		}
		listCommitsAndCheck := func(request *pfs.ListCommitRequest, expectedIDs []string) []*pfs.CommitInfo {
			listCommitClient, err := env.PachClient.PfsAPIClient.ListCommit(env.PachClient.Ctx(), request)
			require.NoError(t, err)
			cis, err := grpcutil.Collect[*pfs.CommitInfo](listCommitClient, 1000)
			require.NoError(t, err)
			require.Equal(t, len(expectedIDs), len(cis))
			for i, ci := range cis {
				require.Equal(t, expectedIDs[i], ci.Commit.Id)
			}
			return cis
		}
		// we should get the latest commit first
		cis := listCommitsAndCheck(&pfs.ListCommitRequest{
			Repo:   client.NewRepo(pfs.DefaultProjectName, "foo"),
			Number: 1,
		}, []string{commits[2].Id})
		cis = listCommitsAndCheck(&pfs.ListCommitRequest{
			Repo:        client.NewRepo(pfs.DefaultProjectName, "foo"),
			Number:      2,
			StartedTime: cis[0].Started,
		}, []string{commits[1].Id, commits[0].Id})
		// no commits should be returned if we set the started time to be the time of the oldest commit
		_ = listCommitsAndCheck(&pfs.ListCommitRequest{
			Repo:        client.NewRepo(pfs.DefaultProjectName, "foo"),
			Number:      1,
			StartedTime: cis[1].Started,
		}, []string{})
		// we should get the oldest commit first if reverse is set to true
		_ = listCommitsAndCheck(&pfs.ListCommitRequest{
			Repo:    client.NewRepo(pfs.DefaultProjectName, "foo"),
			Number:  1,
			Reverse: true,
		}, []string{commits[0].Id})
	})

	suite.Run("InvalidProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		c := env.PachClient
		badFormatErr := "only alphanumeric characters"
		testCases := []struct {
			projectName string
			errMatch    string // "" means no error
		}{
			{tu.UniqueString("my-PROJECT_0123456789"), ""},
			{"lenny", ""},
			{tu.UniqueString("lenny123"), ""},
			{tu.UniqueString("lenny_123"), ""},
			{tu.UniqueString("lenny-123"), ""},
			// {tu.UniqueString("_project"), badFormatErr}, // Require CORE-1343
			// {tu.UniqueString("project-"), badFormatErr},
			{pfs.DefaultProjectName, "already exists"},
			{tu.UniqueString("/repo"), badFormatErr},
			{tu.UniqueString("lenny.123"), badFormatErr},
			{tu.UniqueString("lenny:"), badFormatErr},
			{tu.UniqueString("lenny,"), badFormatErr},
			{tu.UniqueString("lenny#"), badFormatErr},
			{tu.UniqueString("_lenny"), "must start with an alphanumeric character"},
			{tu.UniqueString("-lenny"), "must start with an alphanumeric character"},
			{tu.UniqueString("!project"), badFormatErr},
			{tu.UniqueString("\""), badFormatErr},
			{tu.UniqueString("\\"), badFormatErr},
			{tu.UniqueString("'"), badFormatErr},
			{tu.UniqueString("[]{}"), badFormatErr},
			{tu.UniqueString("|"), badFormatErr},
			{tu.UniqueString("new->project"), badFormatErr},
			{tu.UniqueString("project?"), badFormatErr},
			{tu.UniqueString("project:1"), badFormatErr},
			{tu.UniqueString("project;"), badFormatErr},
			{tu.UniqueString("project."), badFormatErr},
		}
		for _, tc := range testCases {
			t.Run(tc.projectName, func(t *testing.T) {
				err := c.CreateProject(tc.projectName)
				if tc.errMatch != "" {
					require.YesError(t, err)
					require.ErrorContains(t, err, tc.errMatch)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	suite.Run("DefaultProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		// the default project should already exist
		pp, err := env.PachClient.ListProject()
		require.NoError(t, err)
		var sawDefault bool
		for _, p := range pp {
			if p.Project.Name == pfs.DefaultProjectName {
				sawDefault = true
			}
		}
		require.True(t, sawDefault)
		// this should fail because the default project already exists
		require.YesError(t, env.PachClient.CreateProject(pfs.DefaultProjectName))
		// but this should succeed
		var desc = "A different description."
		require.NoError(t, env.PachClient.UpdateProject(pfs.DefaultProjectName, desc))
		// and it should have taken effect
		pp, err = env.PachClient.ListProject()
		require.NoError(t, err)
		for _, p := range pp {
			if p.Project.Name == pfs.DefaultProjectName {
				require.Equal(t, desc, p.Description)
			}
		}
		// deleting the default project is allowed, too
		require.NoError(t, env.PachClient.DeleteProject(pfs.DefaultProjectName, false))
		// and now there should be no default project
		sawDefault = false
		pp, err = env.PachClient.ListProject()
		require.NoError(t, err)
		for _, p := range pp {
			if p.Project.Name == pfs.DefaultProjectName {
				sawDefault = true
			}
		}
		require.False(t, sawDefault)
		// redeleting should be an error
		require.YesError(t, env.PachClient.DeleteProject(pfs.DefaultProjectName, false))
		// force-deleting should error too (FIXME: confirm)
		require.YesError(t, env.PachClient.DeleteProject(pfs.DefaultProjectName, true))
	})

	suite.Run("InvalidRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "/repo"))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny123"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny_123"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny-123"))

		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny.123"))
		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny:"))
		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny,"))
		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "lenny#"))
	})

	suite.Run("CreateSameRepoInParallel", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numGoros := 1000
		errCh := make(chan error)
		for i := 0; i < numGoros; i++ {
			go func() {
				errCh <- env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo")
			}()
		}
		successCount := 0
		totalCount := 0
		for err := range errCh {
			totalCount++
			if err == nil {
				successCount++
			}
			if totalCount == numGoros {
				break
			}
		}
		// When creating the same repo, precisiely one attempt should succeed
		require.Equal(t, 1, successCount)
	})

	suite.Run("CreateDifferentRepoInParallel", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numGoros := 1000
		errCh := make(chan error)
		for i := 0; i < numGoros; i++ {
			i := i
			go func() {
				errCh <- env.PachClient.CreateRepo(pfs.DefaultProjectName, fmt.Sprintf("repo%d", i))
			}()
		}
		successCount := 0
		totalCount := 0
		for err := range errCh {
			totalCount++
			if err == nil {
				successCount++
			}
			if totalCount == numGoros {
				break
			}
		}
		require.Equal(t, numGoros, successCount)
	})

	suite.Run("CreateProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		// 51-character project names are allowed
		require.NoError(t, env.PachClient.CreateProject("123456789A123456789B123456789C123456789D123456789E1"))
		// 52-character project names are not allowed
		require.YesError(t, env.PachClient.CreateProject("123456789A123456789B123456789C123456789D123456789E12"))
	})

	suite.Run("CreateRepoNonExistentProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.YesError(t, env.PachClient.CreateRepo("foo", "bar"))
		require.NoError(t, env.PachClient.CreateProject("foo"))
		require.NoError(t, env.PachClient.CreateRepo("foo", "bar"))
	})

	suite.Run("CreateRepoDeleteRepoRace", func(t *testing.T) {
		t.Skip()
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		for i := 0; i < 100; i++ {
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "foo"))
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "bar"))
			errCh := make(chan error)
			go func() {
				errCh <- env.PachClient.DeleteRepo(pfs.DefaultProjectName, "foo", false)
			}()
			go func() {
				errCh <- env.PachClient.CreateBranch(pfs.DefaultProjectName, "bar", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "foo", "master")})
			}()
			err1 := <-errCh
			err2 := <-errCh
			// these two operations should never race in such a way that they
			// both succeed, leaving us with a repo bar that has a nonexistent
			// provenance foo
			require.True(t, err1 != nil || err2 != nil)
			require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "bar", true))
			require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "foo", true))
		}
	})

	suite.Run("CreateRepoWithSameNameAndAuthInDifferentProjects", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		// activate auth
		peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
		tu.ActivateLicense(t, env.PachClient, peerPort)
		_, err := env.PachClient.Enterprise.Activate(env.PachClient.Ctx(),
			&enterprise.ActivateRequest{
				LicenseServer: "grpc://localhost:" + peerPort,
				Id:            "localhost",
				Secret:        "localhost",
			})
		require.NoError(t, err, "activate client should work")
		_, err = env.AuthServer.Activate(env.PachClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
		require.NoError(t, err, "activate server should work")
		env.PachClient.SetAuthToken(tu.RootToken)
		require.NoError(t, config.WritePachTokenToConfig(tu.RootToken, false))
		client := env.PachClient.WithCtx(context.Background())
		_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
		require.NoError(t, err)

		// create two projects
		project1 := tu.UniqueString("project")
		project2 := tu.UniqueString("project")
		require.NoError(t, client.CreateProject(project1))
		require.NoError(t, client.CreateProject(project2))

		// create repo with same name across both projects
		repo := tu.UniqueString("repo")
		require.NoError(t, client.CreateRepo(project1, repo))
		require.NoError(t, client.CreateRepo(project2, repo))

		// inspect whether the repos are actually from different projects
		repoInfo1, err := client.InspectRepo(project1, repo)
		require.NoError(t, err)
		require.Equal(t, project1, repoInfo1.Repo.Project.Name)
		repoInfo2, err := client.InspectRepo(project2, repo)
		require.NoError(t, err)
		require.Equal(t, project2, repoInfo2.Repo.Project.Name)

		// delete both repos
		require.NoError(t, client.DeleteRepo(project1, repo, true))
		require.NoError(t, client.DeleteRepo(project2, repo, true))
	})

	suite.Run("Branch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		require.Nil(t, commitInfo.ParentCommit)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		require.NotNil(t, commitInfo.ParentCommit)
	})

	suite.Run("ToggleBranchProvenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "in"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "out"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "out", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "in", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "out", "master", ""))
		outRepo := client.NewRepo(pfs.DefaultProjectName, "out")
		cis, err := env.PachClient.ListCommit(outRepo, outRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(cis))

		// Create initial input commit, and make sure we get an output commit
		require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, "in", "master", ""), "1", strings.NewReader("1")))
		cis, err = env.PachClient.ListCommit(outRepo, outRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))
		require.NoError(t, finishCommit(env.PachClient, "out", "master", ""))
		// make sure the output commit and input commit have the same ID
		inCommitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "in", "master", "")
		require.NoError(t, err)
		outCommitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "out", "master", "")
		require.NoError(t, err)
		require.Equal(t, inCommitInfo.Commit.Id, outCommitInfo.Commit.Id)

		// Toggle out@master provenance off
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "out", "master", "master", "", nil))
		inRepo := client.NewRepo(pfs.DefaultProjectName, "in")
		cis, err = env.PachClient.ListCommit(inRepo, inRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))

		// Create new input commit & make sure no new output commit is created
		require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, "in", "master", ""), "2", strings.NewReader("2")))
		cis, err = env.PachClient.ListCommit(outRepo, outRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))
		// make sure output commit still matches the old input commit
		inCommitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "in", "", "master~1") // old input commit
		require.NoError(t, err)
		outCommitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "out", "master", "")
		require.NoError(t, err)
		require.Equal(t, inCommitInfo.Commit.Id, outCommitInfo.Commit.Id)

		// Toggle out@master provenance back on, creating a new output commit
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "out", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "in", "master"),
		}))
		cis, err = env.PachClient.ListCommit(outRepo, outRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 3, len(cis))

		inRepo = client.NewRepo(pfs.DefaultProjectName, "in")
		cis, err = env.PachClient.ListCommit(inRepo, inRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 3, len(cis))

		// make sure new output commit has the same ID as the new input commit
		outCommitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "out", "master", "")
		require.NoError(t, err)
		inCommitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "in", "", outCommitInfo.Commit.Id)
		require.NoError(t, err)
		require.NotEqual(t, inCommitInfo.Commit.Id, outCommitInfo.Commit.Id)
		inMasterCommitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "in", "master", "")
		require.NoError(t, err)
		require.Equal(t, inCommitInfo.Commit.Id, inMasterCommitInfo.Commit.Id)
	})

	suite.Run("RecreateBranchProvenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "in"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "out"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "out", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "in", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "out", "master", ""))
		require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, "in", "master", ""), "foo", strings.NewReader("foo")))
		outRepo := client.NewRepo(pfs.DefaultProjectName, "out")
		cis, err := env.PachClient.ListCommit(outRepo, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))
		commit1 := cis[0].Commit
		require.NoError(t, env.PachClient.DeleteBranch(pfs.DefaultProjectName, "out", "master", false))
		require.NoError(t, finishCommit(env.PachClient, "out", "", commit1.Id))
		cis, err = env.PachClient.ListCommit(outRepo, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "out", "master", "", commit1.Id, []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "in", "master")}))
		cis, err = env.PachClient.ListCommit(outRepo, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 3, len(cis))
		require.Equal(t, commit1, cis[1].Commit)
	})

	// TODO(acohen4): should we dis-allow moving a branch with provenance? Probably since this would break Branch/Head invariant.
	//
	// When a branch head is moved to an older commit (commonly referred to as a rewind), the expected behavior is for
	// that branch ID to match the assigned ID, and for commit with a new Global ID to propagate to all downstream branches.
	suite.Run("RewindBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		// for provenance DAG: { a.master <- b.master <- c.master }
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "a"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "b"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "c"))
		provB := []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "a", "master")}
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "b", "master", "", "", provB))
		provC := []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "b", "master")}
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "c", "master", "", "", provC))
		// create and finish 3 commits on a.master, each with "/file" containing data "1", "2", & "3" respectively
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "a", "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "file", strings.NewReader("1")))
		require.NoError(t, finishCommit(env.PachClient, commit1.Branch.Repo.Name, commit1.Branch.Name, commit1.Id))
		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "a", "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "file", strings.NewReader("2")))
		require.NoError(t, finishCommit(env.PachClient, commit2.Branch.Repo.Name, commit1.Branch.Name, commit2.Id))
		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "a", "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "file", strings.NewReader("3")))
		require.NoError(t, finishCommit(env.PachClient, commit3.Branch.Repo.Name, commit1.Branch.Name, commit3.Id))
		checkRepoCommits := func(repos []string, commits []*pfs.Commit) {
			ids := []string{}
			for _, c := range commits {
				ids = append(ids, c.Id)
			}
			for _, repo := range repos {
				listCommitClient, err := env.PachClient.PfsAPIClient.ListCommit(env.PachClient.Ctx(), &pfs.ListCommitRequest{
					Repo: client.NewRepo(pfs.DefaultProjectName, repo),
					All:  true,
				})
				require.NoError(t, err)
				cis, err := grpcutil.Collect[*pfs.CommitInfo](listCommitClient, 1000)
				require.NoError(t, err)
				// There will be some empty commits on each branch from creation, ignore
				// those and just check that the latest commits match.
				require.True(t, len(ids) <= len(cis))
				for i, id := range ids {
					require.Equal(t, id, cis[i].Commit.Id)
				}
			}
		}
		checkRepoCommits([]string{"a", "b", "c"}, []*pfs.Commit{commit3, commit2, commit1})
		// Rewinding branch b.master to commit2 can reuses that commit, and propagates a new commit to C
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "b", "master", "master", commit2.Id, provB))
		ci, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "b", "master", "")
		require.NoError(t, err)
		require.Equal(t, ci.Commit.Id, commit2.Id)
		checkRepoCommits([]string{"a", "b"}, []*pfs.Commit{commit3, commit2, commit1})
		// Check that a@<new-global-id> resolves to commit2
		ci, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "c", "master", "")
		require.NoError(t, err)
		checkRepoCommits([]string{"c"}, []*pfs.Commit{ci.Commit, commit3, commit2, commit1})
		// The commit4 data in "a" should be the same as what we wrote into commit2 (as that's the source data in "b")
		ci, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "c", "master", "")
		require.NoError(t, err)
		latestCommit := ci.Commit.Id
		aCommit := client.NewCommit(pfs.DefaultProjectName, "a", "", latestCommit)
		var b bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(aCommit, "file", &b))
		require.Equal(t, "2", b.String())
		// Rewinding branch b.master to commit1 can reuses that commit, and propagates a new commit to C
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "b", "master", "master", commit1.Id, provB))
		require.NoError(t, err)
		ci, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "b", "master", "")
		require.NoError(t, err)
		commit4 := ci.Commit
		require.Equal(t, commit4.Id, commit1.Id)
		// The commit4 data in "a" should be from commit1
		ci, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "c", "master", "")
		require.NoError(t, err)
		latestCommit = ci.Commit.Id
		aCommit = client.NewCommit(pfs.DefaultProjectName, "a", "", latestCommit)
		b.Reset()
		require.NoError(t, env.PachClient.GetFile(aCommit, "file", &b))
		require.Equal(t, "1", b.String())
	})

	suite.Run("RewindInput", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "Z"))
		// A ─▶ B ─▶ Z
		//           ▲
		//      C ───╯
		txnInfo, err := c.ExecuteInTransaction(func(tx *client.APIClient) error {
			if err := tx.CreateBranch(pfs.DefaultProjectName, "A", "master", "", "", nil); err != nil {
				return err
			}
			if err := tx.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "",
				[]*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}); err != nil {
				return err
			}
			if err := tx.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", nil); err != nil {
				return err
			}
			return tx.CreateBranch(pfs.DefaultProjectName, "Z", "master", "", "", []*pfs.Branch{
				client.NewBranch(pfs.DefaultProjectName, "B", "master"),
				client.NewBranch(pfs.DefaultProjectName, "C", "master"),
			})
		})
		require.NoError(t, err)
		firstID := txnInfo.Transaction.Id
		expectedMasterCommits := map[string]string{
			"A": firstID,
			"B": firstID,
			"C": firstID,
			"Z": firstID,
		}
		assertMasterHeads(t, c, expectedMasterCommits)
		// make two commits by putting files in A
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "A", "master", ""), "one", strings.NewReader("foo")))
		info, err := c.InspectCommit(pfs.DefaultProjectName, "A", "master", "")
		require.NoError(t, err)
		secondID := info.Commit.Id
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "A", "master", ""), "two", strings.NewReader("bar")))
		// rewind once, everything should be back to firstCommit
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "A", "master", "master", secondID, nil))
		info, err = c.InspectCommit(pfs.DefaultProjectName, "B", "master", "")
		require.NoError(t, err)
		fourthID := info.Commit.Id
		expectedMasterCommits["A"] = secondID
		expectedMasterCommits["B"] = fourthID
		expectedMasterCommits["Z"] = fourthID
		assertMasterHeads(t, c, expectedMasterCommits)
		// add a file to C
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "C", "master", ""), "file", strings.NewReader("baz")))
		info, err = c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		expectedMasterCommits["C"] = info.Commit.Id
		expectedMasterCommits["Z"] = info.Commit.Id
		assertMasterHeads(t, c, expectedMasterCommits)
		// rewind A back to the start
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "A", "master", "master", firstID, nil))
		info, err = c.InspectCommit(pfs.DefaultProjectName, "B", "master", "")
		require.NoError(t, err)
		expectedMasterCommits["A"] = firstID
		expectedMasterCommits["B"] = info.Commit.Id
		expectedMasterCommits["Z"] = info.Commit.Id
		assertMasterHeads(t, c, expectedMasterCommits)
	})

	suite.Run("RewindProvenanceChange", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "A", "master", "", "", nil))
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", nil))
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		_, err := c.InspectCommit(pfs.DefaultProjectName, "A", "master", "")
		require.NoError(t, err)
		_, err = c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "A", "master", ""), "foo", strings.NewReader("bar")))
		oldHead, err := c.InspectCommit(pfs.DefaultProjectName, "A", "master", "")
		require.NoError(t, err)
		_, err = c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		expectedMasterCommits := map[string]string{
			"A": oldHead.Commit.Id,
			"C": oldHead.Commit.Id,
		}
		assertMasterHeads(t, c, expectedMasterCommits)
		// add B to C's provenance
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "A", "master"),
			client.NewBranch(pfs.DefaultProjectName, "B", "master"),
		}))
		cHead, err := c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		expectedMasterCommits["C"] = cHead.Commit.Id
		assertMasterHeads(t, c, expectedMasterCommits)
		// add a file to B and record C's new head
		require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "B", "master", ""), "foo", strings.NewReader("bar")))
		cHead, err = c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		expectedMasterCommits["B"] = cHead.Commit.Id
		expectedMasterCommits["C"] = cHead.Commit.Id
		assertMasterHeads(t, c, expectedMasterCommits)
		// rewind A to before the provenance change
		require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "A", "master", "master", oldHead.ParentCommit.Id, nil))
		// this must create a new commit set, since the old one isn't consistent with current provenance
		aHead, err := c.InspectBranch(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.Equal(t, aHead.Head.Id, oldHead.ParentCommit.Id)
		cHeadOld := cHead
		cHead, err = c.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		// there's no clear relationship between C's new state and any old one, so the new commit's parent should be the previous head
		require.NoError(t, err)
		require.NotNil(t, cHead.ParentCommit)
		require.NotEqual(t, cHeadOld.Commit.Id, cHead.Commit.Id)
		require.Equal(t, cHeadOld.Commit.Id, cHead.ParentCommit.Id)
	})

	suite.Run("CreateAndInspectRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		repoInfo, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, repo, repoInfo.Repo.Name)
		require.NotNil(t, repoInfo.Created)
		require.Equal(t, 0, int(repoInfo.Details.SizeBytes))

		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		_, err = env.PachClient.InspectRepo(pfs.DefaultProjectName, "nonexistent")
		require.YesError(t, err)

		_, err = env.PachClient.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
			Repo: client.NewRepo(pfs.DefaultProjectName, "somerepo1"),
		})
		require.NoError(t, err)
	})

	suite.Run("CreateRepoWithoutProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo123"
		_, err := env.PachClient.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
			Repo: &pfs.Repo{Name: repo},
		})
		require.NoError(t, err)

		repoInfo, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, repo, repoInfo.Repo.Name)
		require.NotNil(t, repoInfo.Created)
		require.Equal(t, 0, int(repoInfo.Details.SizeBytes))
	})

	suite.Run("CreateRepoWithEmptyProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo123"
		_, err := env.PachClient.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
			Repo: &pfs.Repo{Project: &pfs.Project{}, Name: repo},
		})
		require.NoError(t, err)

		repoInfo, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, repo, repoInfo.Repo.Name)
		require.NotNil(t, repoInfo.Created)
		require.Equal(t, 0, int(repoInfo.Details.SizeBytes))
	})

	suite.Run("ListRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numRepos := 10
		var repoNames []string
		for i := 0; i < numRepos; i++ {
			repo := fmt.Sprintf("repo%d", i)
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
			repoNames = append(repoNames, repo)
		}

		repoInfos, err := env.PachClient.ListRepo()
		require.NoError(t, err)
		require.ElementsEqualUnderFn(t, repoNames, repoInfos, RepoInfoToName)
	})

	// Make sure that artifacts of deleted repos do not resurface
	suite.Run("CreateDeletedRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		repoProto := client.NewRepo(pfs.DefaultProjectName, repo)

		systemRepo := client.NewSystemRepo(pfs.DefaultProjectName, repo, pfs.MetaRepoType)
		_, err := env.PachClient.PfsAPIClient.CreateRepo(env.PachClient.Ctx(), &pfs.CreateRepoRequest{
			Repo: systemRepo,
		})
		require.NoError(t, err)

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("foo")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		_, err = env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: systemRepo.NewBranch("master"),
		})
		require.NoError(t, err)

		commitInfos, err := env.PachClient.ListCommit(repoProto, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))

		commitInfos, err = env.PachClient.ListCommit(systemRepo, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))

		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, repo, false))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commitInfos, err = env.PachClient.ListCommit(repoProto, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 0, len(commitInfos))

		branchInfos, err := env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 0, len(branchInfos))

		repoInfos, err := env.PachClient.ListRepoByType("")
		require.NoError(t, err)
		require.Equal(t, 1, len(repoInfos))

	})

	// Make sure that commits of deleted repos do not resurface
	suite.Run("ListCommitLimit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("foo")))
		require.NoError(t, env.PachClient.PutFile(commit, "bar", strings.NewReader("bar")))
		commitInfos, err := env.PachClient.ListCommit(commit.Branch.Repo, nil, nil, 1)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
	})

	// The DAG looks like this before the update:
	// prov1 prov2
	//   \    /
	//    repo
	//   /    \
	// d1      d2
	//
	// Looks like this after the update:
	//
	// prov2 prov3
	//   \    /
	//    repo
	//   /    \
	// d1      d2
	suite.Run("UpdateProvenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		prov1 := "prov1"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, prov1))
		prov2 := "prov2"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, prov2))
		prov3 := "prov3"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, prov3))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, prov1, "master"), client.NewBranch(pfs.DefaultProjectName, prov2, "master")}))

		downstream1 := "downstream1"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, downstream1))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, downstream1, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, repo, "master")}))

		downstream2 := "downstream2"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, downstream2))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, downstream2, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, repo, "master")}))

		// Without the Update flag it should fail
		require.YesError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		_, err := env.PachClient.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{
			Repo:   client.NewRepo(pfs.DefaultProjectName, repo),
			Update: true,
		})
		require.NoError(t, err)

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, prov2, "master"), client.NewBranch(pfs.DefaultProjectName, prov3, "master")}))

		// We should be able to delete prov1 since it's no longer the provenance
		// of other repos.
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, prov1, false))

		// We shouldn't be able to delete prov3 since it's now a provenance
		// of other repos.
		require.YesError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, prov3, false))
	})

	suite.Run("PutFileIntoOpenCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		require.YesError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader("foo\n")))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "foo", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		require.YesError(t, env.PachClient.PutFile(commit2, "foo", strings.NewReader("foo\n")))
	})

	suite.Run("PutFileDirectoryTraversal", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "repo", "master")
		require.NoError(t, err)
		masterCommit := client.NewCommit(pfs.DefaultProjectName, "repo", "master", "")

		mfc, err := env.PachClient.NewModifyFileClient(masterCommit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("../foo", strings.NewReader("foo\n")))
		require.YesError(t, mfc.Close())

		fis, err := env.PachClient.ListFileAll(masterCommit, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fis))

		mfc, err = env.PachClient.NewModifyFileClient(masterCommit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("foo/../../bar", strings.NewReader("foo\n")))
		require.YesError(t, mfc.Close())

		mfc, err = env.PachClient.NewModifyFileClient(masterCommit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("foo/../bar", strings.NewReader("foo\n")))
		require.YesError(t, mfc.Close())

		fis, err = env.PachClient.ListFileAll(masterCommit, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fis))
	})

	suite.Run("CreateInvalidBranchName", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		// Create a branch that's the same length as a commit ID
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, uuid.NewWithoutDashes())
		require.YesError(t, err)
	})

	suite.Run("CreateBranchHeadOnOtherRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		// create two repos, and create a branch on one that tries to point on another's existing branch
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		otherRepo := "other"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, otherRepo))
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, otherRepo, "master")
		require.NoError(t, err)
		otherMasterCommit := client.NewCommit(pfs.DefaultProjectName, otherRepo, "master", "")

		mfc, err := env.PachClient.NewModifyFileClient(masterCommit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("/foo", strings.NewReader("foo\n")))
		require.NoError(t, mfc.Close())

		mfc, err = env.PachClient.NewModifyFileClient(otherMasterCommit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("/bar", strings.NewReader("bar\n")))
		require.NoError(t, mfc.Close())

		// Create a branch on one repo that points to a branch on another repo
		_, err = env.PachClient.PfsAPIClient.CreateBranch(
			env.PachClient.Ctx(),
			&pfs.CreateBranchRequest{
				Branch: client.NewBranch(pfs.DefaultProjectName, repo, "test"),
				Head:   client.NewCommit(pfs.DefaultProjectName, otherRepo, "master", ""),
			},
		)
		require.YesError(t, err)
		require.True(t, strings.Contains(err.Error(), "branch and head commit must belong to the same repo"))
	})

	suite.Run("DeleteProject", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		_, err := env.PachClient.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "test"}})
		require.NoError(t, err)
		_, err = env.PachClient.PfsAPIClient.DeleteProject(ctx, &pfs.DeleteProjectRequest{Project: &pfs.Project{Name: "test"}})
		require.NoError(t, err)
	})

	suite.Run("DeleteProjectWithRepos", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		_, err := env.PachClient.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "test"}})
		require.NoError(t, err)
		_, err = env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: &pfs.Repo{Project: &pfs.Project{Name: "default"}, Name: "test"}})
		require.NoError(t, err)
		_, err = env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: &pfs.Repo{Project: &pfs.Project{Name: "test"}, Name: "test"}})
		require.NoError(t, err)
		_, err = env.PachClient.PfsAPIClient.DeleteProject(ctx, &pfs.DeleteProjectRequest{Project: &pfs.Project{Name: "test"}})
		require.YesError(t, err)

		repoInfos, err := env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Len(t, repoInfos, 2) // should still have both repos

		// this should fail because there is still a repo
		_, err = env.PachClient.PfsAPIClient.DeleteProject(ctx,
			&pfs.DeleteProjectRequest{
				Project: &pfs.Project{Name: "test"},
				Force:   true,
			})
		require.YesError(t, err)

		_, err = env.PachClient.PfsAPIClient.DeleteRepo(ctx,
			&pfs.DeleteRepoRequest{
				Repo: &pfs.Repo{
					Project: &pfs.Project{Name: "test"},
					Name:    "test",
					Type:    pfs.UserRepoType,
				},
			})
		require.NoError(t, err)
		repoInfos, err = env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Len(t, repoInfos, 1) // should still have both repos
		_, err = env.PachClient.PfsAPIClient.DeleteProject(ctx,
			&pfs.DeleteProjectRequest{
				Project: &pfs.Project{Name: "test"},
			})
		require.NoError(t, err)

		repoInfos, err = env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Len(t, repoInfos, 1) // should only have one, in the default project
		for _, repoInfo := range repoInfos {
			require.Equal(t, repoInfo.GetRepo().GetProject().GetName(), "default")
		}
	})

	suite.Run("DeleteRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numRepos := 10
		repoNames := make(map[string]bool)
		for i := 0; i < numRepos; i++ {
			repo := fmt.Sprintf("repo%d", i)
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
			repoNames[repo] = true
		}

		reposToRemove := 5
		for i := 0; i < reposToRemove; i++ {
			// Pick one random element from repoNames
			for repoName := range repoNames {
				require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, repoName, false))
				delete(repoNames, repoName)
				break
			}
		}

		repoInfos, err := env.PachClient.ListRepo()
		require.NoError(t, err)

		for _, repoInfo := range repoInfos {
			require.True(t, repoNames[repoInfo.Repo.Name])
		}

		require.Equal(t, len(repoInfos), numRepos-reposToRemove)
	})

	suite.Run("DeleteRepoProvenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		// Create two repos, one as another's provenance
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit.Branch.Name, commit.Id))

		// Delete the provenance repo; that should fail.
		require.YesError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "A", false))

		// Delete the leaf repo, then the provenance repo; that should succeed
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "B", false))

		// Should be in a consistent state after B is deleted
		require.NoError(t, env.PachClient.FsckFastExit())

		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "A", false))

		repoInfos, err := env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Equal(t, 0, len(repoInfos))

		// Create two repos again
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))

		// Force delete should succeed
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "A", true))

		repoInfos, err = env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Equal(t, 1, len(repoInfos))

		// Everything should be consistent
		require.NoError(t, env.PachClient.FsckFastExit())
	})

	suite.Run("DeleteRepos", func(t *testing.T) {
		t.Parallel()
		var (
			ctx                  = pctx.TestContext(t)
			env                  = realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
			projectName          = tu.UniqueString("project")
			untouchedProjectName = tu.UniqueString("project")
			reposToDelete        = make(map[string]bool)
			untouchedRepos       []*pfs.Repo
		)
		_, err := env.PachClient.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: projectName}})
		require.NoError(t, err)
		_, err = env.PachClient.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: untouchedProjectName}})
		require.NoError(t, err)

		numRepos := 12
		for i := 0; i < numRepos; i++ {
			repoName := fmt.Sprintf("repo%d", i)
			repoToDelete := &pfs.Repo{
				Project: &pfs.Project{Name: projectName},
				Name:    repoName,
				Type:    pfs.UserRepoType,
			}
			_, err := env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: repoToDelete})
			require.NoError(t, err)
			reposToDelete[repoToDelete.String()] = true

			untouchedRepo := &pfs.Repo{
				Project: &pfs.Project{Name: untouchedProjectName},
				Name:    repoName,
				Type:    pfs.UserRepoType,
			}
			_, err = env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: untouchedRepo})
			require.NoError(t, err)
			untouchedRepos = append(untouchedRepos, untouchedRepo)
		}

		// DeleteRepos with no projects should not return an error.
		_, err = env.PachClient.PfsAPIClient.DeleteRepos(ctx, &pfs.DeleteReposRequest{})
		require.NoError(t, err)

		// DeleteRepos with a non-nil, zero-length list of projects should also not error.
		_, err = env.PachClient.PfsAPIClient.DeleteRepos(ctx, &pfs.DeleteReposRequest{Projects: make([]*pfs.Project, 0)})
		require.NoError(t, err)

		// DeleteRepos with an invalid project should not error because
		// there will simply be no repos to delete.
		resp, err := env.PachClient.PfsAPIClient.DeleteRepos(ctx, &pfs.DeleteReposRequest{Projects: []*pfs.Project{{Name: tu.UniqueString("noexist")}}})
		require.NoError(t, err)
		require.Len(t, resp.Repos, 0)

		// DeleteRepos should delete all repos in the given project and none in other projects.
		resp, err = env.PachClient.PfsAPIClient.DeleteRepos(ctx, &pfs.DeleteReposRequest{Projects: []*pfs.Project{{Name: projectName}}})
		require.NoError(t, err)
		require.Len(t, resp.Repos, len(reposToDelete))
		for _, repo := range resp.Repos {
			if !reposToDelete[repo.String()] {
				t.Errorf("deleted repo %v, which should not have been", repo)
			}
		}
		repoStream, err := env.PachClient.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{Projects: []*pfs.Project{{Name: untouchedProjectName}}})
		require.NoError(t, err)
		var repos = make(map[string]bool)
		for {
			repoInfo, err := repoStream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			if repoInfo.Repo.Project.Name != untouchedProjectName {
				t.Errorf("ListProject({Projects: []string{%s}}) return repo in project %v", untouchedProjectName, repoInfo.Repo.Project)
			}
			repos[repoInfo.Repo.String()] = true
		}
		for _, repo := range untouchedRepos {
			if !repos[repo.String()] {
				t.Errorf("repo %v not found in repos %v", repo.String(), repos)
			}
		}

		// DeleteRepos with all set should delete every single repo.
		resp, err = env.PachClient.PfsAPIClient.DeleteRepos(ctx, &pfs.DeleteReposRequest{All: true})
		require.NoError(t, err)
		require.Len(t, resp.Repos, len(untouchedRepos))

		repoStream, err = env.PachClient.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{})
		require.NoError(t, err)
		var count int
		for {
			_, err := repoStream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			count++
		}
		require.Equal(t, count, 0)
	})

	suite.Run("InspectCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		started := time.Now()
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		fileContent := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader(fileContent)))

		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id)
		require.NoError(t, err)
		tStarted := commitInfo.Started.AsTime()
		require.Equal(t, commit, commitInfo.Commit)
		require.Nil(t, commitInfo.Finished)
		require.Equal(t, &pfs.CommitInfo_Details{}, commitInfo.Details) // no details for an unfinished commit
		require.True(t, started.Before(tStarted))
		require.Nil(t, commitInfo.Finished)
		finished := time.Now()

		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		commitInfo, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, repo, commit.Branch.Name, commit.Id)
		require.NoError(t, err)
		tStarted = commitInfo.Started.AsTime()
		tFinished := commitInfo.Finished.AsTime()
		require.Equal(t, commit, commitInfo.Commit)
		require.NotNil(t, commitInfo.Finished)
		require.Equal(t, len(fileContent), int(commitInfo.Details.SizeBytes))
		require.True(t, started.Before(tStarted))
		require.True(t, finished.Before(tFinished))
	})

	suite.Run("InspectCommitWait", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		var eg errgroup.Group
		eg.Go(func() error {
			time.Sleep(2 * time.Second)
			return finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id)
		})

		commitInfo, err := env.PachClient.WaitCommit(pfs.DefaultProjectName, commit.Branch.Repo.Name, commit.Branch.Name, commit.Id)
		require.NoError(t, err)
		require.NotNil(t, commitInfo.Finished)

		require.NoError(t, eg.Wait())
	})

	suite.Run("DropCommitSet", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		project := tu.UniqueString("prj")
		require.NoError(t, env.PachClient.CreateProject(project))
		repo := tu.UniqueString("test")
		require.NoError(t, env.PachClient.CreateRepo(project, repo))

		commit1, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)

		fileContent := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader(fileContent)))

		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, "master", ""))

		commit2, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)

		// Squashing should fail because the commit has no children
		err = env.PachClient.SquashCommitSet(commit2.Id)
		require.YesError(t, err)
		require.True(t, pfsserver.IsSquashWithoutChildrenErr(err))

		require.NoError(t, env.PachClient.DropCommitSet(commit2.Id))

		_, err = env.PachClient.InspectCommit(project, repo, commit2.Branch.Name, commit2.Id)
		require.YesError(t, err)

		// Check that the head has been set to the parent
		commitInfo, err := env.PachClient.InspectCommit(project, repo, "master", "")
		require.NoError(t, err)
		require.Equal(t, commit1.Id, commitInfo.Commit.Id)

		// Check that the branch still exists
		branchInfos, err := env.PachClient.ListBranch(project, repo)
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfos))
	})

	suite.Run("DropCommitSetOnlyCommitInBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		repoProto := client.NewRepo(pfs.DefaultProjectName, repo)

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("foo\n")))

		commitInfos, err := env.PachClient.ListCommit(repoProto, repoProto.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, commit, commitInfos[0].Commit)

		require.NoError(t, env.PachClient.DropCommitSet(commit.Id))

		// The branch has not been deleted, though its head has been replaced with an empty commit
		branchInfos, err := env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfos))
		commitInfos, err = env.PachClient.ListCommit(repoProto, repoProto.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, branchInfos[0].Head, commitInfos[0].Commit)

		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		require.Equal(t, int64(0), commitInfo.Details.SizeBytes)

		// Check that repo size is back to 0
		repoInfo, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, int64(0), repoInfo.Details.SizeBytes)
	})

	suite.Run("DropCommitSetFinished", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		repoProto := client.NewRepo(pfs.DefaultProjectName, repo)

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		commitInfos, err := env.PachClient.ListCommit(repoProto, repoProto.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, commit, commitInfos[0].Commit)
		require.Equal(t, int64(4), commitInfos[0].Details.SizeBytes)

		require.NoError(t, env.PachClient.DropCommitSet(commit.Id))

		// The branch has not been deleted, though it only has an empty commit
		branchInfos, err := env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfos))
		commitInfos, err = env.PachClient.ListCommit(repoProto, repoProto.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, branchInfos[0].Head, commitInfos[0].Commit)
		require.NotEqual(t, commit, commitInfos[0].Commit)

		// Check that repo size is back to 0
		repoInfo, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 0, int(repoInfo.Details.SizeBytes))
	})

	suite.Run("BasicFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		file := "file"
		data := "data"
		require.NoError(t, env.PachClient.PutFile(commit, file, strings.NewReader(data)))

		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		var b bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit, "file", &b))
		require.Equal(t, data, b.String())
	})

	suite.Run("SimpleFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit1, "foo", &buffer))
		require.Equal(t, "foo\n", buffer.String())

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		err = finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id)
		require.NoError(t, err)

		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit1, "foo", &buffer))
		require.Equal(t, "foo\n", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit2, "foo", &buffer))
		require.Equal(t, "foo\nfoo\n", buffer.String())
	})

	suite.Run("StartCommitWithUnfinishedParent", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		// fails because the parent commit has not been finished
		require.YesError(t, err)

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
	})

	suite.Run("ProvenanceWithinSingleRepoDisallowed", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		// TODO: implement in terms of branch provenance
		// test: repo -> repo
		// test: a -> b -> a
	})

	suite.Run("AncestrySyntax", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "file", strings.NewReader("1")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "file", strings.NewReader("2")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "file", strings.NewReader("3")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^1")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~1")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^^")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~~")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^2")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~2")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.1")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.2")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.3")
		require.NoError(t, err)
		require.Equal(t, commit3, commitInfo.Commit)

		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^^^")
		require.YesError(t, err)

		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~~~")
		require.YesError(t, err)

		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master^3")
		require.YesError(t, err)

		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master~3")
		require.YesError(t, err)

		for i := 1; i <= 2; i++ {
			_, err := env.PachClient.InspectFile(client.NewCommit(pfs.DefaultProjectName, repo, "", fmt.Sprintf("%v^%v", commit3.Id, 3-i)), "file")
			require.NoError(t, err)
		}

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", 0)), "file", &buffer))
		require.Equal(t, "3", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", 1)), "file", &buffer))
		require.Equal(t, "2", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", 2)), "file", &buffer))
		require.Equal(t, "1", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", -1)), "file", &buffer))
		require.Equal(t, "1", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", -2)), "file", &buffer))
		require.Equal(t, "2", buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", ancestry.Add("master", -3)), "file", &buffer))
		require.Equal(t, "3", buffer.String())

		// Adding a bunch of commits to the head of the branch shouldn't change the forward references.
		// (It will change backward references.)
		for i := 0; i < 10; i++ {
			require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "file", strings.NewReader(fmt.Sprintf("%d", i+4))))
		}
		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.1")
		require.NoError(t, err)
		require.Equal(t, commit1, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.2")
		require.NoError(t, err)
		require.Equal(t, commit2, commitInfo.Commit)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "", "master.3")
		require.NoError(t, err)
		require.Equal(t, commit3, commitInfo.Commit)
	})

	// Provenance implements the following DAG
	//  A ─▶ B ─▶ C ─▶ D
	//            ▲
	//  E ────────╯

	suite.Run("Provenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "D"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "E"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "B", "master"), client.NewBranch(pfs.DefaultProjectName, "E", "master")}))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "D", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "C", "master")}))

		branchInfo, err := env.PachClient.InspectBranch(pfs.DefaultProjectName, "B", "master")
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfo.Provenance))
		branchInfo, err = env.PachClient.InspectBranch(pfs.DefaultProjectName, "C", "master")
		require.NoError(t, err)
		require.Equal(t, 3, len(branchInfo.Provenance))
		branchInfo, err = env.PachClient.InspectBranch(pfs.DefaultProjectName, "D", "master")
		require.NoError(t, err)
		require.Equal(t, 4, len(branchInfo.Provenance))

		ACommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", ACommit.Branch.Name, ACommit.Id))

		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "B", "master", "")
		require.NoError(t, err)
		require.Equal(t, ACommit.Id, commitInfo.Commit.Id)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		require.Equal(t, ACommit.Id, commitInfo.Commit.Id)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "D", "master", "")
		require.NoError(t, err)
		require.Equal(t, ACommit.Id, commitInfo.Commit.Id)

		ECommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "E", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "E", ECommit.Branch.Name, ECommit.Id))

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "B", "master", "")
		require.NoError(t, err)
		aliasCommitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "B", "", ECommit.Id)
		require.NoError(t, err)
		require.Equal(t, aliasCommitInfo.Commit.Id, commitInfo.Commit.Id)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)
		aliasCommitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "C", "", ECommit.Id)
		require.NoError(t, err)
		require.Equal(t, aliasCommitInfo.Commit.Id, commitInfo.Commit.Id)

		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "D", "master", "")
		require.NoError(t, err)
		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "D", "", ECommit.Id)
		require.NoError(t, err)
		require.Equal(t, ECommit.Id, commitInfo.Commit.Id)
	})

	suite.Run("CommitBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "input"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "output1"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "output2"))
		// Make two branches provenant on the master branch
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output1", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "input", "master")}))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output2", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "input", "master")}))
		// Now make a commit on the master branch, which should trigger a downstream commit on each of the two branches
		masterCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "input", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "input", masterCommit.Branch.Name, masterCommit.Id))
		// Check that the commit in output1 has the information and provenance we expect
		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output1", "master", "")
		require.NoError(t, err)
		require.Equal(t, masterCommit.Id, commitInfo.Commit.Id)
		// Check that the commit in output2 has the information and provenance we expect
		commitInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "output2", "master", "")
		require.NoError(t, err)
		require.Equal(t, masterCommit.Id, commitInfo.Commit.Id)
	})

	suite.Run("CommitBranchProvenanceMovement", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "input"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "output"))
		// create two commits on input@master
		parentCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "input", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "input", parentCommit.Branch.Name, parentCommit.Id))
		masterCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "input", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "input", masterCommit.Branch.Name, masterCommit.Id))
		// Make input@A set to input@master
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "input", "A", masterCommit.Branch.Name, masterCommit.Id, nil))
		// Now create a branch provenant on input@A
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output", "C", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "input", "A")}))
		require.NoError(t, finishCommit(env.PachClient, "output", "C", ""))
		// The head commit of the C branch should have a different ID than the new heads
		// of branches A and B, but should include them in its DirectProvenance
		cHead, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output", "C", "")
		require.NoError(t, err)
		aHead, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "input", "A", "")
		require.NoError(t, err)
		// Chain of C should only have one commit, differing from A & B
		require.NotEqual(t, cHead.Commit.Id, aHead.Commit.Id)
		require.Equal(t, 1, len(cHead.DirectProvenance))
		require.Equal(t, aHead.Commit.Id, cHead.DirectProvenance[0].Id)
		// We should be able to squash input@A^
		require.NoError(t, env.PachClient.SquashCommitSet(aHead.ParentCommit.Id))
		aHead, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "input", "A", "")
		require.NoError(t, err)
		require.Nil(t, aHead.ParentCommit)
		_, err = env.PFSServer.DropCommitSet(ctx, &pfs.DropCommitSetRequest{CommitSet: &pfs.CommitSet{Id: cHead.Commit.Id}})
		require.NoError(t, err)
		cHeadNew, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output", "C", "")
		require.NoError(t, err)
		aHeadNew, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "input", "A", "")
		require.NoError(t, err)
		// check C.Head's commit provenance
		require.Equal(t, 1, len(cHeadNew.DirectProvenance))
		require.Equal(t, "input", cHeadNew.DirectProvenance[0].Repo.Name)
		require.Equal(t, aHeadNew.Commit.Id, cHeadNew.DirectProvenance[0].Id)
		require.NotEqual(t, cHead.Commit.Id, cHeadNew.Commit.Id)
		// It should also be ok to make new commits on branch A
		aCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "input", "A")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "input", aCommit.Branch.Name, aCommit.Id))
	})

	// For the "V" shaped DAG:
	//     A   B
	//     |   |
	//      \ /
	//       C
	// When commit x propagates from A to C, verify that queries to B@x resolve to the B commit
	// that was used in C@x.
	suite.Run("ResolveAlias", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(pctx.TestContext(t), t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "",
			[]*pfs.Branch{
				client.NewBranch(pfs.DefaultProjectName, "A", "master"),
				client.NewBranch(pfs.DefaultProjectName, "B", "master")},
		))
		ci, err := c.InspectCommit(pfs.DefaultProjectName, "A", "master", "")
		require.NoError(t, err)
		// create another commit on A that will propagate to C
		commit1, err := c.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, c.PutFile(commit1, "file", strings.NewReader("foo")))
		require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, "A", commit1.Branch.Name, commit1.Id))
		// Assert that referencing "B" at the latest commit ID gives us the latest "B" commit
		cis, err := c.InspectCommitSet(commit1.Id)
		require.NoError(t, err)
		require.Equal(t, 3, len(cis))
		resolvedAliasCommit, err := c.InspectCommit(pfs.DefaultProjectName, "B", "", commit1.Id)
		require.NoError(t, err)
		require.Equal(t, ci.Commit.Id, resolvedAliasCommit.Commit.Id)

	})

	// In general a commit set can be squashed without use of force whenever the following commit set was initiated
	// by from the same branch.
	// For example, consider the following branch provenance graph (where A & B are input branches):
	// C: {A, B}
	// D: {B}
	// Now consider the following sequence of commits against input branches A & B, mapped to the expanded commit set they initiate.
	// A@1 -> { A@1, C@1, B@0 }
	// A@2 -> { A@2, C@2, B@0 }
	// A@3 -> { A@3, C@3, B@0 }
	// B@4 -> { B@4, C@4, D@4, A@3 }
	// B@5 -> { B@5, C@5, D@5, A@3 }
	// A@6 -> { A@6, C@6, B@5 }
	// A@7 -> { A@7, C@7, B@5 }
	// The commit sets that can be safely squashed here are: {1, 2, 4, 6}. The reason is that no other commit set referenes commits
	// from those commit sets except the commit sets that produced them. On the other hand, commit set 3 cannot be
	// independently deleted since commit set 4 references its commit A@3. Similarly, commit set 5 has its commit B@5 referenced
	// in commit set 6 & 7.
	// These commit sets that we'll consider troublesome for deletes, become troublesome when the next commit set originates from a
	// different branch; i.e. when one commit set in the sequence originates from branch A, and the following from B.
	// So conversely, we can conclude that whenever we have a sequence of commit sets that are originated from the
	// same branch, all but the last commit set in the sequence can be considered safe to delete.
	suite.Run("SquashComplex", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient
		var benches []string
		start := time.Now()
		checkpoint := func(label string) {
			benches = append(benches, fmt.Sprintf("%s: %v\n", label, time.Since(start)))
			start = time.Now()
		}
		makeCommit := func(c *client.APIClient, repos ...string) *pfs.CommitSet {
			totalSubvenance := make(map[string]*pfs.Branch)
			txn, err := c.StartTransaction()
			require.NoError(t, err)
			txnClient := env.PachClient.WithTransaction(txn)
			cs := []*pfs.Commit{}
			for _, r := range repos {
				commit, err := txnClient.StartCommit(pfs.DefaultProjectName, r, "master")
				require.NoError(t, err)
				cs = append(cs, commit)
			}
			txnInfo, err := env.PachClient.FinishTransaction(txn)
			require.NoError(t, err)
			for _, commit := range cs {
				require.NoError(t, c.PutFile(commit, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
				require.NoError(t, finishCommit(c, commit.Repo.Name, "master", ""))
				bi, err := c.InspectBranch(pfs.DefaultProjectName, commit.Repo.Name, "master")
				require.NoError(t, err)
				for _, subv := range bi.Subvenance {
					totalSubvenance[subv.String()] = subv
				}
			}
			for _, subv := range totalSubvenance {
				require.NoError(t, finishCommit(c, subv.Repo.Name, subv.Name, ""))
			}
			checkpoint(fmt.Sprintf("makeCommit: %v", repos))
			return &pfs.CommitSet{Id: txnInfo.Transaction.Id}
		}
		listCommitSets := func() []*pfs.CommitSetInfo {
			csClient, err := c.ListCommitSet(ctx, &pfs.ListCommitSetRequest{Project: client.NewProject(pfs.DefaultProjectName)})
			require.NoError(t, err)
			css, err := grpcutil.Collect[*pfs.CommitSetInfo](csClient, 1000)
			require.NoError(t, err)
			checkpoint("list commit sets")
			return css
		}
		buildDAG(t, c, "A")
		checkpoint("build A")
		buildDAG(t, c, "B")
		checkpoint("build B")
		buildDAG(t, c, "C", "B", "A")
		checkpoint("build C")
		buildDAG(t, c, "D", "B")
		checkpoint("build D")
		// first test a simple squash - removing an independent commit set from the middle of the DAG
		squashCandidate := makeCommit(c, "A")
		makeCommit(c, "A")
		require.Equal(t, 6, len(listCommitSets()))
		_, err := c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: squashCandidate})
		require.NoError(t, err)
		checkpoint("squash first")
		require.Equal(t, 5, len(listCommitSets()))
		badCandidate1 := makeCommit(c, "B")
		badCandidate2 := makeCommit(c, "A")
		squashCandidate = makeCommit(c, "B")
		squashCandidate2 := makeCommit(c, "B")
		latest := makeCommit(c, "B")
		start = time.Now()
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: badCandidate1})
		require.YesError(t, err)
		checkpoint("bad squash 1")
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: badCandidate2})
		require.YesError(t, err)
		checkpoint("bad squash 2")
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: squashCandidate2})
		require.NoError(t, err)
		checkpoint("good squash")
		require.Equal(t, 9, len(listCommitSets()))
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: squashCandidate})
		require.NoError(t, err)
		checkpoint("good squash")
		require.Equal(t, 8, len(listCommitSets()))
		// ok now lets do something crazy
		for i := 0; i < 5; i++ {
			for _, b := range []string{"A", "B"} {
				latest = makeCommit(c, b)
			}
		}
		require.Equal(t, 18, len(listCommitSets()))
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: latest})
		require.YesError(t, err)
		checkpoint("bad squash latest")
		// add slice to top so that we can successfully squash
		realLatest := makeCommit(c, "A", "B")
		_, err = c.PfsAPIClient.SquashCommitSet(ctx, &pfs.SquashCommitSetRequest{CommitSet: latest})
		require.NoError(t, err)
		checkpoint("good squash latest")
		css := listCommitSets()
		require.Equal(t, 18, len(css))
		require.Equal(t, realLatest.Id, css[0].CommitSet.Id)
		fmt.Println(benches)
	})

	suite.Run("Branch1", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(masterCommit, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(masterCommit, "foo", &buffer))
		require.Equal(t, "foo\n", buffer.String())
		branchInfos, err := env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfos))
		require.Equal(t, "master", branchInfos[0].Branch.Name)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(masterCommit, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		buffer = bytes.Buffer{}
		require.NoError(t, env.PachClient.GetFile(masterCommit, "foo", &buffer))
		require.Equal(t, "foo\nfoo\n", buffer.String())
		branchInfos, err = env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 1, len(branchInfos))
		require.Equal(t, "master", branchInfos[0].Branch.Name)

		// Check that moving the commit to other branches uses the same CommitSet ID and extends the existing CommitSet
		commitInfos, err := env.PachClient.InspectCommitSet(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master2", commit.Branch.Name, commit.Id, nil))

		commitInfos, err = env.PachClient.InspectCommitSet(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master3", commit.Branch.Name, commit.Id, nil))

		commitInfos, err = env.PachClient.InspectCommitSet(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))

		branchInfos, err = env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		require.Equal(t, 3, len(branchInfos))
		require.Equal(t, "master3", branchInfos[0].Branch.Name)
		require.Equal(t, "master2", branchInfos[1].Branch.Name)
		require.Equal(t, "master", branchInfos[2].Branch.Name)
	})

	suite.Run("PinBranchCommitsDAG", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		input, pinInput, output := "input", "pinInput", "output"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, input))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, pinInput))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, output))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, output, "master", "", "",
			[]*pfs.Branch{
				client.NewBranch(pfs.DefaultProjectName, input, "master"),
				client.NewBranch(pfs.DefaultProjectName, pinInput, "pin1"),
			}))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, pinInput, "master", "pin1", "", nil))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, pinInput, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, pinInput, "master", commit1.Id))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, pinInput, "pin2", "master", "", nil))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, output, "master", "", "",
			[]*pfs.Branch{
				client.NewBranch(pfs.DefaultProjectName, input, "master"),
				client.NewBranch(pfs.DefaultProjectName, pinInput, "pin2"),
			}))
	})

	suite.Run("PutFileBig", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		// Write a big blob that would normally not fit in a block
		fileSize := int(pfs.ChunkSize + 5*1024*1024)
		expectedOutputA := random.String(fileSize)
		r := strings.NewReader(string(expectedOutputA))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", r))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		fileInfo, err := env.PachClient.InspectFile(commit1, "foo")
		require.NoError(t, err)
		require.Equal(t, fileSize, int(fileInfo.SizeBytes))

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit1, "foo", &buffer))
		require.Equal(t, string(expectedOutputA), buffer.String())
	})

	suite.Run("PutFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, env.PachClient.PutFile(masterCommit, "file", strings.NewReader("foo")))
		var buf bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(masterCommit, "file", &buf))
		require.Equal(t, "foo", buf.String())
		require.NoError(t, env.PachClient.PutFile(masterCommit, "file", strings.NewReader("bar")))
		buf.Reset()
		require.NoError(t, env.PachClient.GetFile(masterCommit, "file", &buf))
		require.Equal(t, "bar", buf.String())
		require.NoError(t, env.PachClient.DeleteFile(masterCommit, "file"))
		require.NoError(t, env.PachClient.PutFile(masterCommit, "file", strings.NewReader("buzz")))
		buf.Reset()
		require.NoError(t, env.PachClient.GetFile(masterCommit, "file", &buf))
		require.Equal(t, "buzz", buf.String())
	})

	suite.Run("PutFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, env.PachClient.PutFile(commit1, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
		require.NoError(t, env.PachClient.PutFile(masterCommit, "file", strings.NewReader("buzz\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		expected := "foo\nbar\nbuzz\n"
		buffer := &bytes.Buffer{}
		require.NoError(t, env.PachClient.GetFile(commit1, "file", buffer))
		require.Equal(t, expected, buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(masterCommit, "file", buffer))
		require.Equal(t, expected, buffer.String())

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "file", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		require.NoError(t, env.PachClient.PutFile(commit2, "file", strings.NewReader("bar\n"), client.WithAppendPutFile()))
		require.NoError(t, env.PachClient.PutFile(masterCommit, "file", strings.NewReader("buzz\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit2, "file", buffer))
		require.Equal(t, expected, buffer.String())
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(masterCommit, "file", buffer))
		require.Equal(t, expected, buffer.String())

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "foo", "", commit3.Id, nil))

		commit4, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "foo")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit4, "file", strings.NewReader("foo\nbar\nbuzz\n"), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, commit4.Branch.Name, commit4.Id))

		// commit 3 should have remained unchanged
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit3, "file", buffer))
		require.Equal(t, expected, buffer.String())

		expected = "foo\nbar\nbuzz\nfoo\nbar\nbuzz\nfoo\nbar\nbuzz\n"
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit4, "file", buffer))
		require.Equal(t, expected, buffer.String())
	})

	suite.Run("PutFileBranchCommitID", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		err := env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, repo, "", "master"), "foo", strings.NewReader("foo\n"), client.WithAppendPutFile())
		require.NoError(t, err)
	})

	suite.Run("PutSameFileInParallel", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		var eg errgroup.Group
		for i := 0; i < 3; i++ {
			eg.Go(func() error {
				return env.PachClient.PutFile(commit, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile())
			})
		}
		require.NoError(t, eg.Wait())
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit, "foo", &buffer))
		require.Equal(t, "foo\nfoo\nfoo\n", buffer.String())
	})

	suite.Run("InspectFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		fileContent1 := "foo\n"
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader(fileContent1), client.WithAppendPutFile()))
		checks := func() {
			fileInfo, err := env.PachClient.InspectFile(commit1, "foo")
			require.NoError(t, err)
			require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
			require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		checks()

		fileContent2 := "barbar\n"
		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "foo", strings.NewReader(fileContent2), client.WithAppendPutFile()))

		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		fileInfo, err := env.PachClient.InspectFile(commit2, "foo")
		require.NoError(t, err)
		require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
		require.Equal(t, len(fileContent1+fileContent2), int(fileInfo.SizeBytes))

		fileInfo, err = env.PachClient.InspectFile(commit2, "foo")
		require.NoError(t, err)
		require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
		require.Equal(t, len(fileContent1)+len(fileContent2), int(fileInfo.SizeBytes))

		fileContent3 := "bar\n"
		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "bar", strings.NewReader(fileContent3), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		fis, err := env.PachClient.ListFileAll(commit3, "")
		require.NoError(t, err)
		require.Equal(t, 2, len(fis))

		require.Equal(t, len(fis), 2)
	})

	suite.Run("InspectFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		fileContent1 := "foo\n"
		fileContent2 := "buzz\n"

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader(fileContent1), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfo, err := env.PachClient.InspectFile(commit, "/file")
		require.NoError(t, err)
		require.Equal(t, len(fileContent1), int(fileInfo.SizeBytes))
		require.Equal(t, "/file", fileInfo.File.Path)
		require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader(fileContent1), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfo, err = env.PachClient.InspectFile(commit, "file")
		require.NoError(t, err)
		require.Equal(t, len(fileContent1)*2, int(fileInfo.SizeBytes))
		require.Equal(t, "/file", fileInfo.File.Path)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit, "file"))
		require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader(fileContent2), client.WithAppendPutFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfo, err = env.PachClient.InspectFile(commit, "file")
		require.NoError(t, err)
		require.Equal(t, len(fileContent2), int(fileInfo.SizeBytes))
	})

	suite.Run("InspectFile3", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		fileContent1 := "foo\n"
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "foo/bar", strings.NewReader(fileContent1)))
		fileInfo, err := env.PachClient.InspectFile(commit1, "foo")
		require.NoError(t, err)
		require.NotNil(t, fileInfo)

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		fi, err := env.PachClient.InspectFile(commit1, "foo/bar")
		require.NoError(t, err)
		require.NotNil(t, fi)

		fileContent2 := "barbar\n"
		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "foo", strings.NewReader(fileContent2)))

		fileInfo, err = env.PachClient.InspectFile(commit2, "foo")
		require.NoError(t, err)
		require.NotNil(t, fileInfo)

		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		fi, err = env.PachClient.InspectFile(commit2, "foo")
		require.NoError(t, err)
		require.NotNil(t, fi)

		fileContent3 := "bar\n"
		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "bar", strings.NewReader(fileContent3)))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))
		fi, err = env.PachClient.InspectFile(commit3, "bar")
		require.NoError(t, err)
		require.NotNil(t, fi)
	})

	suite.Run("InspectDir", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		fileContent := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit1, "dir/foo", strings.NewReader(fileContent)))

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		fileInfo, err := env.PachClient.InspectFile(commit1, "dir/foo")
		require.NoError(t, err)
		require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
		require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)

		fileInfo, err = env.PachClient.InspectFile(commit1, "dir")
		require.NoError(t, err)
		require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
		require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

		_, err = env.PachClient.InspectFile(commit1, "")
		require.NoError(t, err)
		require.Equal(t, len(fileContent), int(fileInfo.SizeBytes))
		require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)
	})

	suite.Run("InspectDir2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		fileContent := "foo\n"

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/1", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit, "dir/2", strings.NewReader(fileContent)))

		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfo, err := env.PachClient.InspectFile(commit, "/dir")
		require.NoError(t, err)
		require.Equal(t, "/dir/", fileInfo.File.Path)
		require.Equal(t, pfs.FileType_DIR, fileInfo.FileType)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/3", strings.NewReader(fileContent)))

		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		_, err = env.PachClient.InspectFile(commit, "dir")
		require.NoError(t, err)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		err = env.PachClient.DeleteFile(commit, "dir/2")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		_, err = env.PachClient.InspectFile(commit, "dir")
		require.NoError(t, err)
	})

	suite.Run("ListFileTwoCommits", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		numFiles := 5

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		for i := 0; i < numFiles; i++ {
			require.NoError(t, env.PachClient.PutFile(commit1, fmt.Sprintf("file%d", i), strings.NewReader("foo\n")))
		}

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		fis, err := env.PachClient.ListFileAll(commit1, "")
		require.NoError(t, err)
		require.Equal(t, numFiles, len(fis))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		for i := 0; i < numFiles; i++ {
			require.NoError(t, env.PachClient.PutFile(commit2, fmt.Sprintf("file2-%d", i), strings.NewReader("foo\n")))
		}

		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		fis, err = env.PachClient.ListFileAll(commit2, "")
		require.NoError(t, err)
		require.Equal(t, 2*numFiles, len(fis))

		fis, err = env.PachClient.ListFileAll(commit1, "")
		require.NoError(t, err)
		require.Equal(t, numFiles, len(fis))

		fis, err = env.PachClient.ListFileAll(commit2, "")
		require.NoError(t, err)
		require.Equal(t, 2*numFiles, len(fis))
	})

	suite.Run("ListFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		fileContent1 := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit, "dir/foo", strings.NewReader(fileContent1)))

		fileContent2 := "bar\n"
		require.NoError(t, env.PachClient.PutFile(commit, "dir/bar", strings.NewReader(fileContent2)))

		checks := func() {
			fileInfos, err := env.PachClient.ListFileAll(commit, "dir")
			require.NoError(t, err)
			require.Equal(t, 2, len(fileInfos))
			require.True(t, fileInfos[0].File.Path == "/dir/foo" && fileInfos[1].File.Path == "/dir/bar" || fileInfos[0].File.Path == "/dir/bar" && fileInfos[1].File.Path == "/dir/foo")
			require.True(t, fileInfos[0].SizeBytes == fileInfos[1].SizeBytes && fileInfos[0].SizeBytes == int64(len(fileContent1)))

		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		checks()
	})

	suite.Run("ListFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		fileContent := "foo\n"

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/1", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit, "dir/2", strings.NewReader(fileContent)))
		require.NoError(t, err)

		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err := env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/3", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err = env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 3, len(fileInfos))

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		err = env.PachClient.DeleteFile(commit, "dir/2")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err = env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))
	})

	suite.Run("ListFile3", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		fileContent := "foo\n"

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/1", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit, "dir/2", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err := env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/3/foo", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit, "dir/3/bar", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err = env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 3, len(fileInfos))
		require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent)*2)

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		err = env.PachClient.DeleteFile(commit, "dir/3/bar")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err = env.PachClient.ListFileAll(commit, "dir")
		require.NoError(t, err)
		require.Equal(t, 3, len(fileInfos))
		require.Equal(t, int(fileInfos[2].SizeBytes), len(fileContent))

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		fileInfos, err = env.PachClient.ListFileAll(commit, "/")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))
	})

	suite.Run("ListFile4", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.2", &bytes.Buffer{}))

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		// should list a directory but not siblings
		var fis []*pfs.FileInfo
		require.NoError(t, env.PachClient.ListFile(commit1, "/dir1", func(fi *pfs.FileInfo) error {
			fis = append(fis, fi)
			return nil
		}))
		require.ElementsEqual(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, finfosToPaths(fis))
		// should list the root
		fis = nil
		require.NoError(t, env.PachClient.ListFile(commit1, "/", func(fi *pfs.FileInfo) error {
			fis = append(fis, fi)
			return nil
		}))
		require.ElementsEqual(t, []string{"/dir1/", "/dir2/"}, finfosToPaths(fis))
	})

	suite.Run("RootDirectory", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		fileContent := "foo\n"

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader(fileContent)))

		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		fileInfos, err := env.PachClient.ListFileAll(commit, "")
		require.NoError(t, err)
		require.Equal(t, 1, len(fileInfos))
	})

	suite.Run("DeleteFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		project := tu.UniqueString("project")
		require.NoError(t, env.PachClient.CreateProject(project))
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(project, repo))

		// Commit 1: Add two files; delete one file within the commit
		commit1, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)

		fileContent1 := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit1, "foo", strings.NewReader(fileContent1)))

		fileContent2 := "bar\n"
		require.NoError(t, env.PachClient.PutFile(commit1, "bar", strings.NewReader(fileContent2)))

		require.NoError(t, env.PachClient.DeleteFile(commit1, "foo"))

		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit1.Branch.Name, commit1.Id))

		_, err = env.PachClient.InspectFile(commit1, "foo")
		require.YesError(t, err)

		// Should see one file
		fileInfos, err := env.PachClient.ListFileAll(commit1, "")
		require.NoError(t, err)
		require.Equal(t, 1, len(fileInfos))

		// Deleting a file in a finished commit should result in an error
		require.YesError(t, env.PachClient.DeleteFile(commit1, "bar"))

		// Empty commit
		commit2, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit2.Branch.Name, commit2.Id))

		// Should still see one files
		fileInfos, err = env.PachClient.ListFileAll(commit2, "")
		require.NoError(t, err)
		require.Equal(t, 1, len(fileInfos))

		// Delete bar
		commit3, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit3, "bar"))

		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit3.Branch.Name, commit3.Id))

		// Should see no file
		fileInfos, err = env.PachClient.ListFileAll(commit3, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))

		_, err = env.PachClient.InspectFile(commit3, "bar")
		require.YesError(t, err)

		// Delete a nonexistent file; it should be no-op
		commit4, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit4, "nonexistent"))
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit4.Branch.Name, commit4.Id))
	})

	suite.Run("DeleteFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "file", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		err = env.PachClient.DeleteFile(commit2, "file")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "file", strings.NewReader("bar\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		expected := "bar\n"
		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "file", &buffer))
		require.Equal(t, expected, buffer.String())

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "file", strings.NewReader("buzz\n")))
		err = env.PachClient.DeleteFile(commit3, "file")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "file", strings.NewReader("foo\n")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		expected = "foo\n"
		buffer.Reset()
		require.NoError(t, env.PachClient.GetFile(commit3, "file", &buffer))
		require.Equal(t, expected, buffer.String())
	})

	suite.Run("DeleteFile3", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		fileContent := "bar\n"
		require.NoError(t, env.PachClient.PutFile(commit1, "/bar", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/dir2/bar", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit2, "/"))
		require.NoError(t, env.PachClient.PutFile(commit2, "/bar", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit2, "/dir1/bar", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit2, "/dir1/dir2/bar", strings.NewReader(fileContent)))
		require.NoError(t, env.PachClient.PutFile(commit2, "/dir1/dir2/barbar", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit3, "/dir1/dir2/"))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		_, err = env.PachClient.InspectFile(commit3, "/dir1")
		require.NoError(t, err)
		_, err = env.PachClient.InspectFile(commit3, "/dir1/bar")
		require.NoError(t, err)
		_, err = env.PachClient.InspectFile(commit3, "/dir1/dir2")
		require.YesError(t, err)
		_, err = env.PachClient.InspectFile(commit3, "/dir1/dir2/bar")
		require.YesError(t, err)
		_, err = env.PachClient.InspectFile(commit3, "/dir1/dir2/barbar")
		require.YesError(t, err)

		commit4, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit4, "/dir1/dir2/bar", strings.NewReader(fileContent)))
		require.NoError(t, finishCommit(env.PachClient, repo, commit4.Branch.Name, commit4.Id))

		_, err = env.PachClient.InspectFile(commit4, "/dir1")
		require.NoError(t, err)
		_, err = env.PachClient.InspectFile(commit4, "/dir1/bar")
		require.NoError(t, err)
		_, err = env.PachClient.InspectFile(commit4, "/dir1/dir2")
		require.NoError(t, err)
		_, err = env.PachClient.InspectFile(commit4, "/dir1/dir2/bar")
		require.NoError(t, err)
	})

	suite.Run("DeleteDir", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		// Commit 1: Add two files into the same directory; delete the directory
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.PutFile(commit1, "dir/foo", strings.NewReader("foo1")))

		require.NoError(t, env.PachClient.PutFile(commit1, "dir/bar", strings.NewReader("bar1")))

		require.NoError(t, env.PachClient.DeleteFile(commit1, "/dir/"))

		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		fileInfos, err := env.PachClient.ListFileAll(commit1, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))

		// dir should not exist
		_, err = env.PachClient.InspectFile(commit1, "dir")
		require.YesError(t, err)

		// Commit 2: Delete the directory and add the same two files
		// The two files should reflect the new content
		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.PutFile(commit2, "dir/foo", strings.NewReader("foo2")))

		require.NoError(t, env.PachClient.PutFile(commit2, "dir/bar", strings.NewReader("bar2")))

		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		// Should see two files
		fileInfos, err = env.PachClient.ListFileAll(commit2, "dir")
		require.NoError(t, err)
		require.Equal(t, 2, len(fileInfos))

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit2, "dir/foo", &buffer))
		require.Equal(t, "foo2", buffer.String())

		var buffer2 bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit2, "dir/bar", &buffer2))
		require.Equal(t, "bar2", buffer2.String())

		// Commit 3: delete the directory
		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.DeleteFile(commit3, "/dir/"))

		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		// Should see zero files
		fileInfos, err = env.PachClient.ListFileAll(commit3, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))

		// One-off commit directory deletion
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, env.PachClient.PutFile(masterCommit, "/dir/foo", strings.NewReader("foo")))
		require.NoError(t, env.PachClient.DeleteFile(masterCommit, "/"))
		fileInfos, err = env.PachClient.ListFileAll(masterCommit, "/")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))
	})

	suite.Run("ListCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		repoProto := client.NewRepo(pfs.DefaultProjectName, repo)
		masterCommit := repoProto.NewCommit("master", "")

		numCommits := 10

		var midCommitID string
		for i := 0; i < numCommits; i++ {
			commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
			require.NoError(t, err)
			require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
			if i == numCommits/2 {
				midCommitID = commit.Id
			}
		}

		// list all commits
		commitInfos, err := env.PachClient.ListCommit(repoProto, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, numCommits, len(commitInfos))

		// Test that commits are sorted in newest-first order
		for i := 0; i < len(commitInfos)-1; i++ {
			require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
		}

		// Now list all commits up to the last commit
		commitInfos, err = env.PachClient.ListCommit(repoProto, masterCommit, nil, 0)
		require.NoError(t, err)
		require.Equal(t, numCommits, len(commitInfos))

		// Test that commits are sorted in newest-first order
		for i := 0; i < len(commitInfos)-1; i++ {
			require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
		}

		// Now list all commits up to the mid commit, excluding the mid commit
		// itself
		commitInfos, err = env.PachClient.ListCommit(repoProto, masterCommit, repoProto.NewCommit("", midCommitID), 0)
		require.NoError(t, err)
		require.Equal(t, numCommits-numCommits/2-1, len(commitInfos))

		// Test that commits are sorted in newest-first order
		for i := 0; i < len(commitInfos)-1; i++ {
			require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
		}

		// list commits by branch
		commitInfos, err = env.PachClient.ListCommit(repoProto, masterCommit, nil, 0)
		require.NoError(t, err)
		require.Equal(t, numCommits, len(commitInfos))

		// Test that commits are sorted in newest-first order
		for i := 0; i < len(commitInfos)-1; i++ {
			require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i+1].Commit)
		}

		// Try listing the commits in reverse order
		commitInfos = nil
		require.NoError(t, env.PachClient.ListCommitF(repoProto, nil, nil, 0, true, func(ci *pfs.CommitInfo) error {
			commitInfos = append(commitInfos, ci)
			return nil
		}))
		for i := 1; i < len(commitInfos); i++ {
			require.Equal(t, commitInfos[i].ParentCommit, commitInfos[i-1].Commit)
		}
	})

	suite.Run("OffsetRead", func(t *testing.T) {
		// TODO(2.0 optional): Decide on how to expose offset read.
		t.Skip("Offset read exists (inefficient), just need to decide on how to expose it in V2")
		// t.Parallel()
		// env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		// repo := "test"
		// require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		// commit, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "")
		// require.NoError(t, err)
		// fileData := "foo\n"
		// require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader(fileData)))
		// require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader(fileData)))

		// var buffer bytes.Buffer
		// require.NoError(t, env.PachClient.GetFile(commit, "foo", int64(len(fileData)*2)+1, 0, &buffer))
		// require.Equal(t, "", buffer.String())

		// require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.ID))

		// buffer.Reset()
		// require.NoError(t, env.PachClient.GetFile(commit, "foo", int64(len(fileData)*2)+1, 0, &buffer))
		// require.Equal(t, "", buffer.String())
	})

	suite.Run("Branch2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		project := tu.UniqueString("project")
		require.NoError(t, env.PachClient.CreateProject(project))
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(project, repo))

		commit, err := env.PachClient.StartCommit(project, repo, "branch1")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("bar")))
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit.Branch.Name, commit.Id))

		expectedBranches := []string{"branch1", "branch2", "branch3"}
		expectedCommits := []*pfs.Commit{}
		for _, branch := range expectedBranches {
			require.NoError(t, env.PachClient.CreateBranch(project, repo, branch, commit.Branch.Name, commit.Id, nil))
			commitInfo, err := env.PachClient.InspectCommit(project, repo, branch, "")
			require.NoError(t, err)
			expectedCommits = append(expectedCommits, commitInfo.Commit)
		}

		branchInfos, err := env.PachClient.ListBranch(project, repo)
		require.NoError(t, err)
		require.Equal(t, len(expectedBranches), len(branchInfos))
		for i, branchInfo := range branchInfos {
			// branches should return in newest-first order
			require.Equal(t, expectedBranches[len(branchInfos)-i-1], branchInfo.GetBranch().GetName())
			require.Equal(t, project, branchInfo.GetBranch().GetRepo().GetProject().GetName())
			// each branch should have a different commit id (from the transaction
			// that moved the branch head)
			headCommit := expectedCommits[len(branchInfos)-i-1]
			require.Equal(t, headCommit, branchInfo.Head)

			// ensure that the branch has the file from the original commit
			var buffer bytes.Buffer
			require.NoError(t, env.PachClient.GetFile(headCommit, "foo", &buffer))
			require.Equal(t, "bar", buffer.String())
		}

		commit2, err := env.PachClient.StartCommit(project, repo, "branch1")
		require.NoError(t, err)
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, "branch1", ""))

		commit2Info, err := env.PachClient.InspectCommit(project, repo, "branch1", "")
		require.NoError(t, err)
		require.Equal(t, expectedCommits[0], commit2Info.ParentCommit)

		// delete the last branch
		lastBranch := expectedBranches[len(expectedBranches)-1]
		require.YesError(t, env.PachClient.DeleteBranch(pfs.DefaultProjectName, repo, lastBranch, false))
		require.NoError(t, env.PachClient.DeleteBranch(project, repo, lastBranch, false))
		branchInfos, err = env.PachClient.ListBranch(project, repo)
		require.NoError(t, err)
		require.Equal(t, 2, len(branchInfos))
		require.Equal(t, "branch2", branchInfos[0].Branch.Name)
		require.Equal(t, expectedCommits[1], branchInfos[0].Head)
		require.Equal(t, "branch1", branchInfos[1].Branch.Name)
		require.Equal(t, commit2, branchInfos[1].Head)
	})

	suite.Run("DeleteNonexistentBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, env.PachClient.DeleteBranch(pfs.DefaultProjectName, repo, "doesnt_exist", false))
	})

	suite.Run("SubscribeCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		project := tu.UniqueString("project")
		require.NoError(t, env.PachClient.CreateProject(project))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(project, repo))

		numCommits := 10

		// create some commits that shouldn't affect the below SubscribeCommit call
		// reproduces #2469
		for i := 0; i < numCommits; i++ {
			commit, err := env.PachClient.StartCommit(project, repo, "master-v1")
			require.NoError(t, err)
			require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit.Branch.Name, commit.Id))
		}

		require.NoErrorWithinT(t, 60*time.Second, func() error {
			var eg errgroup.Group
			nextCommitChan := make(chan *pfs.Commit, numCommits)
			eg.Go(func() error {
				var count int
				err := env.PachClient.SubscribeCommit(client.NewRepo(project, repo), "master", "", pfs.CommitState_STARTED, func(ci *pfs.CommitInfo) error {
					commit := <-nextCommitChan
					require.Equal(t, commit, ci.Commit)
					count++
					if count == numCommits {
						return errutil.ErrBreak
					}
					return nil
				})
				return err
			})
			eg.Go(func() error {
				for i := 0; i < numCommits; i++ {
					commit, err := env.PachClient.StartCommit(project, repo, "master")
					require.NoError(t, err)
					require.NoError(t, finishProjectCommit(env.PachClient, project, repo, commit.Branch.Name, commit.Id))
					nextCommitChan <- commit
				}
				return nil
			})

			return errors.EnsureStack(eg.Wait())
		})
	})

	suite.Run("InspectRepoSimple", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "branch")
		require.NoError(t, err)

		file1Content := "foo\n"
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader(file1Content)))

		file2Content := "bar\n"
		require.NoError(t, env.PachClient.PutFile(commit, "bar", strings.NewReader(file2Content)))

		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		info, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)

		// Size should be 0 because the files were not added to master
		require.Equal(t, int(info.Details.SizeBytes), 0)
	})

	suite.Run("InspectRepoComplex", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		numFiles := 100
		minFileSize := 1000
		maxFileSize := 2000
		totalSize := 0

		for i := 0; i < numFiles; i++ {
			fileContent := random.String(rand.Intn(maxFileSize-minFileSize) + minFileSize)
			fileContent += "\n"
			fileName := fmt.Sprintf("file_%d", i)
			totalSize += len(fileContent)

			require.NoError(t, env.PachClient.PutFile(commit, fileName, strings.NewReader(fileContent)))

		}

		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		info, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)

		require.Equal(t, totalSize, int(info.Details.SizeBytes))

		infos, err := env.PachClient.ListRepo()
		require.NoError(t, err)
		require.Equal(t, 1, len(infos))
	})

	suite.Run("Create", func(t *testing.T) {
		// TODO: Implement put file split writer in V2?
		t.Skip("Put file split writer not implemented in V2")
		// t.Parallel()
		// env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		// repo := "test"
		// require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		// commit, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "")
		// require.NoError(t, err)
		// w, err := env.PachClient.PutFileSplitWriter(repo, commit.Branch.Name, commit.ID, "foo", pfs.Delimiter_NONE, 0, 0, 0, false)
		// require.NoError(t, err)
		// require.NoError(t, w.Close())
		// require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.ID))
		// _, err = env.PachClient.InspectFile(commit, "foo")
		// require.NoError(t, err)
	})

	suite.Run("GetFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := tu.UniqueString("test")
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/file", strings.NewReader("foo\n")))
		checks := func() {
			var buffer bytes.Buffer
			require.NoError(t, env.PachClient.GetFile(commit, "dir/file", &buffer))
			require.Equal(t, "foo\n", buffer.String())
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		checks()
		t.Run("InvalidCommit", func(t *testing.T) {
			buffer := bytes.Buffer{}
			err = env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "", "aninvalidcommitid"), "dir/file", &buffer)
			require.YesError(t, err)
		})
		t.Run("Directory", func(t *testing.T) {
			buffer := bytes.Buffer{}
			err = env.PachClient.GetFile(commit, "dir", &buffer)
			require.YesError(t, err)
		})
		t.Run("WithOffset", func(t *testing.T) {
			repo := "repo"
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

			commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
			require.NoError(t, err)

			file := "file"
			data := "data"
			require.NoError(t, env.PachClient.PutFile(commit, file, strings.NewReader(data)))

			require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

			for i := 0; i <= len(data); i++ {
				var b bytes.Buffer
				require.NoError(t, env.PachClient.GetFile(commit, "file", &b, client.WithOffset(int64(i))))
				if i < len(data) {
					require.Equal(t, data[i:], b.String())
				} else {
					require.Equal(t, "", b.String())
				}
			}
		})
	})

	suite.Run("ManyPutsSingleFileSingleCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		if testing.Short() {
			t.Skip("Skipping long tests in short mode")
		}
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		rawMessage := `{
		"level":"debug",
		"message":{
			"thing":"foo"
		},
		"timing":[1,3,34,6,7]
	}`
		numObjs := 500
		numGoros := 10
		var expectedOutput []byte
		var wg sync.WaitGroup
		for j := 0; j < numGoros; j++ {
			wg.Add(1)
			go func() {
				for i := 0; i < numObjs/numGoros; i++ {
					if err := env.PachClient.PutFile(commit1, "foo", strings.NewReader(rawMessage), client.WithAppendPutFile()); err != nil {
						panic(err)
					}
				}
				wg.Done()
			}()
		}
		for i := 0; i < numObjs; i++ {
			expectedOutput = append(expectedOutput, []byte(rawMessage)...)
		}
		wg.Wait()
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		var buffer bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit1, "foo", &buffer))
		require.Equal(t, string(expectedOutput), buffer.String())
	})

	suite.Run("PutFileValidCharacters", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		// null characters error because when you `ls` files with null characters
		// they truncate things after the null character leading to strange results
		require.YesError(t, env.PachClient.PutFile(commit, "foo\x00bar", strings.NewReader("foobar\n")))

		// Boundary tests for valid character range
		require.YesError(t, env.PachClient.PutFile(commit, "\x1ffoobar", strings.NewReader("foobar\n")))
		require.NoError(t, env.PachClient.PutFile(commit, "foo\x20bar", strings.NewReader("foobar\n")))
		require.NoError(t, env.PachClient.PutFile(commit, "foobar\x7e", strings.NewReader("foobar\n")))
		require.YesError(t, env.PachClient.PutFile(commit, "foo\x7fbar", strings.NewReader("foobar\n")))

		// Random character tests outside and inside valid character range
		require.YesError(t, env.PachClient.PutFile(commit, "foobar\x0b", strings.NewReader("foobar\n")))
		require.NoError(t, env.PachClient.PutFile(commit, "\x41foobar", strings.NewReader("foobar\n")))

		// Glob character test
		require.YesError(t, env.PachClient.PutFile(commit, "foobar*", strings.NewReader("foobar\n")))
	})

	suite.Run("PutFileValidPaths", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		// Duplicate paths, different tags.
		branch := "branch-1"
		require.NoError(t, env.PachClient.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, repo, branch, ""), func(mf client.ModifyFile) error {
			require.NoError(t, mf.PutFile("foo", strings.NewReader("foo\n"), client.WithDatumPutFile("tag1")))
			require.NoError(t, mf.PutFile("foo", strings.NewReader("foo\n"), client.WithDatumPutFile("tag2")))
			return nil
		}))
		commitInfo, err := env.PachClient.WaitCommit(pfs.DefaultProjectName, repo, branch, "")
		require.NoError(t, err)
		require.NotEqual(t, "", commitInfo.Error)
		// Directory and file path collision.
		branch = "branch-2"
		require.NoError(t, env.PachClient.WithModifyFileClient(client.NewCommit(pfs.DefaultProjectName, repo, branch, ""), func(mf client.ModifyFile) error {
			require.NoError(t, mf.PutFile("foo/bar", strings.NewReader("foo\n")))
			require.NoError(t, mf.PutFile("foo", strings.NewReader("foo\n")))
			return nil
		}))
		commitInfo, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, repo, branch, "")
		require.NoError(t, err)
		require.NotEqual(t, "", commitInfo.Error)
	})

	suite.Run("BigListFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		var eg errgroup.Group
		for i := 0; i < 25; i++ {
			for j := 0; j < 25; j++ {
				i := i
				j := j
				eg.Go(func() error {
					return env.PachClient.PutFile(commit, fmt.Sprintf("dir%d/file%d", i, j), strings.NewReader("foo\n"))
				})
			}
		}
		require.NoError(t, eg.Wait())
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		for i := 0; i < 25; i++ {
			files, err := env.PachClient.ListFileAll(commit, fmt.Sprintf("dir%d", i))
			require.NoError(t, err)
			require.Equal(t, 25, len(files))
		}
	})

	suite.Run("StartCommitLatestOnBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)
		require.Equal(t, commit3.Id, commitInfo.Commit.Id)
	})

	suite.Run("CreateBranchTwice", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "foo")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", commit1.Id, nil))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "foo")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", commit2.Id, nil))

		branchInfos, err := env.PachClient.ListBranch(pfs.DefaultProjectName, repo)
		require.NoError(t, err)

		// branches should be returned newest-first
		require.Equal(t, 2, len(branchInfos))
		require.Equal(t, "master", branchInfos[0].Branch.Name)
		require.Equal(t, commit2.Id, branchInfos[0].Head.Id) // aliased branch should have the same commit ID
		require.Equal(t, "foo", branchInfos[1].Branch.Name)
		require.Equal(t, commit2.Id, branchInfos[1].Head.Id) // original branch should remain unchanged
	})

	suite.Run("WaitCommitSet", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))

		ACommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		BCommit := client.NewCommit(pfs.DefaultProjectName, "B", "master", ACommit.Id)
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))

		commitInfos, err := env.PachClient.WaitCommitSetAll(ACommit.Id)
		require.NoError(t, err)
		require.Equal(t, 2, len(commitInfos))
		require.Equal(t, ACommit, commitInfos[0].Commit)
		require.Equal(t, BCommit, commitInfos[1].Commit)
	})

	// WaitCommitSet2 implements the following DAG:
	// A ─▶ B ─▶ C ─▶ D
	suite.Run("WaitCommitSet2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "D"))

		// Create branches and finish the default head commits on the downstream branches
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "D", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "C", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "D", "master", ""))

		ACommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))

		// do the other commits in a goro so we can block for them
		done := make(chan struct{})
		go func() {
			defer close(done)
			require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
			require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
			require.NoError(t, finishCommit(env.PachClient, "D", "master", ""))
		}()

		// Wait for the commits to finish
		commitInfos, err := env.PachClient.WaitCommitSetAll(ACommit.Id)
		require.NoError(t, err)
		BCommit := client.NewCommit(pfs.DefaultProjectName, "B", "master", ACommit.Id)
		CCommit := client.NewCommit(pfs.DefaultProjectName, "C", "master", ACommit.Id)
		DCommit := client.NewCommit(pfs.DefaultProjectName, "D", "master", ACommit.Id)
		require.Equal(t, 4, len(commitInfos))
		require.Equal(t, ACommit, commitInfos[0].Commit)
		require.Equal(t, BCommit, commitInfos[1].Commit)
		require.Equal(t, CCommit, commitInfos[2].Commit)
		require.Equal(t, DCommit, commitInfos[3].Commit)
		<-done
	})

	// A
	//  ╲
	//   ◀
	//    C
	//   ◀
	//  ╱
	// B
	suite.Run("WaitCommitSet3", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master"), client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		ACommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", ACommit.Branch.Name, ACommit.Id))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		BCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "B", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "B", BCommit.Branch.Name, BCommit.Id))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		BCommit, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "B", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "B", BCommit.Branch.Name, BCommit.Id))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		// The first two commits will be A and B, but they aren't deterministically sorted
		commitInfos, err := env.PachClient.WaitCommitSetAll(ACommit.Id)
		require.NoError(t, err)
		require.Equal(t, 3, len(commitInfos))
		require.Equal(t, client.NewCommit(pfs.DefaultProjectName, "C", "master", ACommit.Id), commitInfos[2].Commit)
		commitInfos, err = env.PachClient.WaitCommitSetAll(BCommit.Id)
		require.NoError(t, err)
		require.Equal(t, 3, len(commitInfos))
		require.Equal(t, client.NewCommit(pfs.DefaultProjectName, "C", "master", BCommit.Id), commitInfos[2].Commit)
	})

	suite.Run("WaitCommitSetWithNoDownstreamRepos", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		commitInfos, err := env.PachClient.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, commit, commitInfos[0].Commit)
	})

	suite.Run("WaitOpenCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)

		// do the other commits in a goro so we can block for them
		eg, _ := errgroup.WithContext(context.Background())
		eg.Go(func() error {
			time.Sleep(3 * time.Second)
			if err := finishCommit(env.PachClient, "A", "master", ""); err != nil {
				return err
			}
			return finishCommit(env.PachClient, "B", "master", "")
		})

		t.Cleanup(func() {
			require.NoError(t, eg.Wait())
		})

		// Wait for the commit to finish
		commitInfos, err := env.PachClient.WaitCommitSetAll(commit.Id)
		require.NoError(t, err)
		require.Equal(t, 2, len(commitInfos))
		require.Equal(t, commit, commitInfos[0].Commit)
		require.Equal(t, client.NewCommit(pfs.DefaultProjectName, "B", "master", commit.Id), commitInfos[1].Commit)
	})

	suite.Run("WaitUninvolvedBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", nil))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)

		// Blocking on a commit that doesn't exist does not work
		_, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, "B", "master", commit.Id)
		require.YesError(t, err)
	})

	suite.Run("WaitNonExistentBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		_, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, "A", "foo", "")
		require.YesError(t, err)
	})

	suite.Run("EmptyWait", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		_, err := env.PachClient.WaitCommitSetAll("")
		require.YesError(t, err)
	})

	suite.Run("WaitNonExistentCommitSet", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		_, err := env.PachClient.WaitCommitSetAll("fake-commitset")
		require.YesError(t, err)
		require.True(t, pfsserver.IsCommitSetNotFoundErr(err))
	})

	suite.Run("PutFileSplit", func(t *testing.T) {
		// TODO(2.0 optional): Implement put file split.
		t.Skip("Put file split not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	if testing.Short() {
		//		t.Skip("Skipping integration tests in short mode")
		//	}
		//
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//	commit, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "none", pfs.Delimiter_NONE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "line2", pfs.Delimiter_LINE, 2, 0, 0, false, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "line3", pfs.Delimiter_LINE, 0, 8, 0, false, strings.NewReader("foo\nbar\nbuz\nfiz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "json", pfs.Delimiter_JSON, 0, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "json2", pfs.Delimiter_JSON, 2, 0, 0, false, strings.NewReader("{}{}{}{}"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "json3", pfs.Delimiter_JSON, 0, 4, 0, false, strings.NewReader("{}{}{}{}"))
		//	require.NoError(t, err)
		//
		//	files, err := env.PachClient.ListFileAll(repo, commit.ID, "line2")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(8), fileInfo.SizeBytes)
		//	}
		//
		//	require.NoError(t, finishCommit(env.PachClient, repo, commit.ID))
		//	commit2, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit2.ID, "line", pfs.Delimiter_LINE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, commit2.ID, "json", pfs.Delimiter_JSON, 0, 0, 0, false, strings.NewReader("{}{}{}{}{}{}{}{}{}{}"))
		//	require.NoError(t, err)
		//
		//	files, err = env.PachClient.ListFileAll(repo, commit2.ID, "line")
		//	require.NoError(t, err)
		//	require.Equal(t, 9, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
		//
		//	require.NoError(t, finishCommit(env.PachClient, repo, commit2.ID))
		//	fileInfo, err := env.PachClient.InspectFile(repo, commit.ID, "none")
		//	require.NoError(t, err)
		//	require.Equal(t, pfs.FileType_FILE, fileInfo.FileType)
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "line")
		//	require.NoError(t, err)
		//	require.Equal(t, 6, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit2.ID, "line")
		//	require.NoError(t, err)
		//	require.Equal(t, 9, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "line2")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(8), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "line3")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(8), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "json")
		//	require.NoError(t, err)
		//	require.Equal(t, 20, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(2), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit2.ID, "json")
		//	require.NoError(t, err)
		//	require.Equal(t, 30, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(2), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "json2")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
		//	files, err = env.PachClient.ListFileAll(repo, commit.ID, "json3")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
	})

	suite.Run("PutFileSplitBig", func(t *testing.T) {
		// TODO(2.0 optional): Implement put file split.
		t.Skip("Put file split not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	if testing.Short() {
		//		t.Skip("Skipping integration tests in short mode")
		//	}
		//
		//	// create repos
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//	commit, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	w, err := env.PachClient.PutFileSplitWriter(repo, commit.ID, "line", pfs.Delimiter_LINE, 0, 0, 0, false)
		//	require.NoError(t, err)
		//	for i := 0; i < 1000; i++ {
		//		_, err = w.Write([]byte("foo\n"))
		//		require.NoError(t, err)
		//	}
		//	require.NoError(t, w.Close())
		//	require.NoError(t, finishCommit(env.PachClient, repo, commit.ID))
		//	files, err := env.PachClient.ListFileAll(repo, commit.ID, "line")
		//	require.NoError(t, err)
		//	require.Equal(t, 1000, len(files))
		//	for _, fileInfo := range files {
		//		require.Equal(t, uint64(4), fileInfo.SizeBytes)
		//	}
	})

	suite.Run("PutFileSplitCSV", func(t *testing.T) {
		// TODO(2.0 optional): Implement put file split.
		t.Skip("Put file split not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	// create repos
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//	_, err := env.PachClient.PutFileSplit(repo, "master", "data", pfs.Delimiter_CSV, 0, 0, 0, false,
		//		// Weird, but this is actually two lines ("is\na" is quoted, so one cell)
		//		strings.NewReader("this,is,a,test\n"+
		//			"\"\"\"this\"\"\",\"is\nonly\",\"a,test\"\n"))
		//	require.NoError(t, err)
		//	fileInfos, err := env.PachClient.ListFileAll(repo, "master", "/data")
		//	require.NoError(t, err)
		//	require.Equal(t, 2, len(fileInfos))
		//	var contents bytes.Buffer
		//	env.PachClient.GetFile(repo, "master", "/data/0000000000000000", &contents)
		//	require.Equal(t, "this,is,a,test\n", contents.String())
		//	contents.Reset()
		//	env.PachClient.GetFile(repo, "master", "/data/0000000000000001", &contents)
		//	require.Equal(t, "\"\"\"this\"\"\",\"is\nonly\",\"a,test\"\n", contents.String())
	})

	suite.Run("PutFileSplitSQL", func(t *testing.T) {
		// TODO(2.0 optional): Implement put file split.
		t.Skip("Put file split not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	// create repos
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//
		//	_, err := env.PachClient.PutFileSplit(repo, "master", "/sql", pfs.Delimiter_SQL, 0, 0, 0,
		//		false, strings.NewReader(tu.TestPGDump))
		//	require.NoError(t, err)
		//	fileInfos, err := env.PachClient.ListFileAll(repo, "master", "/sql")
		//	require.NoError(t, err)
		//	require.Equal(t, 5, len(fileInfos))
		//
		//	// Get one of the SQL records & validate it
		//	var contents bytes.Buffer
		//	env.PachClient.GetFile(repo, "master", "/sql/0000000000000000", &contents)
		//	// Validate that the recieved pgdump file creates the cars table
		//	require.Matches(t, "CREATE TABLE public\\.cars", contents.String())
		//	// Validate the SQL header more generally by passing the output of GetFile
		//	// back through the SQL library & confirm that it parses correctly but only
		//	// has one row
		//	pgReader := sql.NewPGDumpReader(bufio.NewReader(bytes.NewReader(contents.Bytes())))
		//	record, err := pgReader.ReadRow()
		//	require.NoError(t, err)
		//	require.Equal(t, "Tesla\tRoadster\t2008\tliterally a rocket\n", string(record))
		//	_, err = pgReader.ReadRow()
		//	require.YesError(t, err)
		//	require.True(t, errors.Is(err, io.EOF))
		//
		//	// Create a new commit that overwrites all existing data & puts it back with
		//	// --header-records=1
		//	commit, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	require.NoError(t, env.PachClient.DeleteFile(repo, commit.ID, "/sql"))
		//	_, err = env.PachClient.PutFileSplit(repo, commit.ID, "/sql", pfs.Delimiter_SQL, 0, 0, 1,
		//		false, strings.NewReader(tu.TestPGDump))
		//	require.NoError(t, err)
		//	require.NoError(t, finishCommit(env.PachClient, repo, commit.ID))
		//	fileInfos, err = env.PachClient.ListFileAll(repo, "master", "/sql")
		//	require.NoError(t, err)
		//	require.Equal(t, 4, len(fileInfos))
		//
		//	// Get one of the SQL records & validate it
		//	contents.Reset()
		//	env.PachClient.GetFile(repo, "master", "/sql/0000000000000003", &contents)
		//	// Validate a that the recieved pgdump file creates the cars table
		//	require.Matches(t, "CREATE TABLE public\\.cars", contents.String())
		//	// Validate the SQL header more generally by passing the output of GetFile
		//	// back through the SQL library & confirm that it parses correctly but only
		//	// has one row
		//	pgReader = sql.NewPGDumpReader(bufio.NewReader(strings.NewReader(contents.String())))
		//	record, err = pgReader.ReadRow()
		//	require.NoError(t, err)
		//	require.Equal(t, "Tesla\tRoadster\t2008\tliterally a rocket\n", string(record))
		//	record, err = pgReader.ReadRow()
		//	require.NoError(t, err)
		//	require.Equal(t, "Toyota\tCorolla\t2005\tgreatest car ever made\n", string(record))
		//	_, err = pgReader.ReadRow()
		//	require.YesError(t, err)
		//	require.True(t, errors.Is(err, io.EOF))
	})

	suite.Run("DiffFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		// Write foo
		c1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(c1, "foo", strings.NewReader("foo\n"), client.WithAppendPutFile()))
		checks := func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c1, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 0, len(oldFis))
			require.Equal(t, 2, len(newFis))
			require.Equal(t, "/foo", newFis[1].File.Path)
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c1.Branch.Name, c1.Id))
		checks()

		// Change the value of foo
		c2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(c2, "/foo"))
		require.NoError(t, env.PachClient.PutFile(c2, "foo", strings.NewReader("not foo\n"), client.WithAppendPutFile()))
		checks = func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c2, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 2, len(oldFis))
			require.Equal(t, "/foo", oldFis[1].File.Path)
			require.Equal(t, 2, len(newFis))
			require.Equal(t, "/foo", newFis[1].File.Path)
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c2.Branch.Name, c2.Id))
		checks()

		// Write bar
		c3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(c3, "/bar", strings.NewReader("bar\n"), client.WithAppendPutFile()))
		checks = func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c3, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 1, len(oldFis))
			require.Equal(t, 2, len(newFis))
			require.Equal(t, "/bar", newFis[1].File.Path)
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c3.Branch.Name, c3.Id))
		checks()

		// Delete bar
		c4, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(c4, "/bar"))
		checks = func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c4, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 2, len(oldFis))
			require.Equal(t, "/bar", oldFis[1].File.Path)
			require.Equal(t, 1, len(newFis))
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c4.Branch.Name, c4.Id))
		checks()

		// Write dir/fizz and dir/buzz
		c5, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(c5, "/dir/fizz", strings.NewReader("fizz\n"), client.WithAppendPutFile()))
		require.NoError(t, env.PachClient.PutFile(c5, "/dir/buzz", strings.NewReader("buzz\n"), client.WithAppendPutFile()))
		checks = func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c5, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 1, len(oldFis))
			require.Equal(t, 4, len(newFis))
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c5.Branch.Name, c5.Id))
		checks()

		// Modify dir/fizz
		c6, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(c6, "/dir/fizz", strings.NewReader("fizz\n"), client.WithAppendPutFile()))
		checks = func() {
			newFis, oldFis, err := env.PachClient.DiffFileAll(c6, "", nil, "", false)
			require.NoError(t, err)
			require.Equal(t, 3, len(oldFis))
			require.Equal(t, "/dir/fizz", oldFis[2].File.Path)
			require.Equal(t, 3, len(newFis))
			require.Equal(t, "/dir/fizz", newFis[2].File.Path)
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, c6.Branch.Name, c6.Id))
		checks()
	})

	suite.Run("GlobFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		if testing.Short() {
			t.Skip("Skipping integration tests in short mode")
		}
		project := tu.UniqueString("prj-")
		require.NoError(t, env.PachClient.CreateProject(project))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(project, repo))
		commit := client.NewCommit(project, repo, "master", "")

		// Write foo
		numFiles := 100
		_, err := env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)

		require.NoError(t, env.PachClient.WithModifyFileClient(commit, func(mf client.ModifyFile) error {
			for i := 0; i < numFiles; i++ {
				require.NoError(t, mf.PutFile(fmt.Sprintf("file%d", i), strings.NewReader("1")))
				require.NoError(t, mf.PutFile(fmt.Sprintf("dir1/file%d", i), strings.NewReader("2")))
				require.NoError(t, mf.PutFile(fmt.Sprintf("dir2/dir3/file%d", i), strings.NewReader("3")))
			}
			return nil
		}))

		checks := func() {
			fileInfos, err := env.PachClient.GlobFileAll(commit, "*")
			require.NoError(t, err)
			require.Equal(t, numFiles+2, len(fileInfos))
			fileInfos, err = env.PachClient.GlobFileAll(commit, "file*")
			require.NoError(t, err)
			require.Equal(t, numFiles, len(fileInfos))
			fileInfos, err = env.PachClient.GlobFileAll(commit, "dir1/*")
			require.NoError(t, err)
			require.Equal(t, numFiles, len(fileInfos))
			fileInfos, err = env.PachClient.GlobFileAll(commit, "dir2/dir3/*")
			require.NoError(t, err)
			require.Equal(t, numFiles, len(fileInfos))
			fileInfos, err = env.PachClient.GlobFileAll(commit, "*/*")
			require.NoError(t, err)
			require.Equal(t, numFiles+1, len(fileInfos))

			var output strings.Builder
			rc, err := env.PachClient.GetFileTAR(commit, "*")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))

			require.Equal(t, numFiles*3, len(output.String()))

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "dir2/dir3/file1?")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))
			require.Equal(t, 10, len(output.String()))

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "**file1?")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))
			require.Equal(t, 30, len(output.String()))

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "**file1")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))
			require.True(t, strings.Contains(output.String(), "1"))
			require.True(t, strings.Contains(output.String(), "2"))
			require.True(t, strings.Contains(output.String(), "3"))

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "**file1")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))
			match, err := regexp.Match("[123]", []byte(output.String()))
			require.NoError(t, err)
			require.True(t, match)

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "dir?")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))

			output = strings.Builder{}
			rc, err = env.PachClient.GetFileTAR(commit, "")
			require.NoError(t, err)
			defer rc.Close()
			require.NoError(t, tarutil.ConcatFileContent(&output, rc))

			output = strings.Builder{}
			err = env.PachClient.GetFile(commit, "garbage", &output)
			require.YesError(t, err)
		}
		checks()
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, "master", ""))
		checks()

		_, err = env.PachClient.StartCommit(project, repo, "master")
		require.NoError(t, err)

		err = env.PachClient.DeleteFile(commit, "/")
		require.NoError(t, err)
		checks = func() {
			fileInfos, err := env.PachClient.GlobFileAll(commit, "**")
			require.NoError(t, err)
			require.Equal(t, 0, len(fileInfos))
		}
		checks()
		require.NoError(t, finishProjectCommit(env.PachClient, project, repo, "master", ""))
		checks()
	})

	suite.Run("GlobFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		if testing.Short() {
			t.Skip("Skipping integration tests in short mode")
		}

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		expectedFileNames := []string{}
		for i := 0; i < 100; i++ {
			filename := fmt.Sprintf("/%d", i)
			require.NoError(t, env.PachClient.PutFile(commit, filename, strings.NewReader(filename)))

			if strings.HasPrefix(filename, "/1") {
				expectedFileNames = append(expectedFileNames, filename)
			}
		}
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		actualFileNames := []string{}
		require.NoError(t, env.PachClient.GlobFile(commit, "/1*", func(fileInfo *pfs.FileInfo) error {
			actualFileNames = append(actualFileNames, fileInfo.File.Path)
			return nil
		}))

		sort.Strings(expectedFileNames)
		sort.Strings(actualFileNames)
		require.Equal(t, expectedFileNames, actualFileNames)
	})

	suite.Run("GlobFile3", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.2", &bytes.Buffer{}))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		globFile := func(pattern string) []string {
			var fis []*pfs.FileInfo
			require.NoError(t, env.PachClient.GlobFile(commit1, pattern, func(fi *pfs.FileInfo) error {
				fis = append(fis, fi)
				return nil
			}))
			return finfosToPaths(fis)
		}
		assert.ElementsMatch(t, []string{"/dir1/file1.2", "/dir2/file2.2"}, globFile("**.2"))
		assert.ElementsMatch(t, []string{"/dir1/file1.1", "/dir1/file1.2"}, globFile("/dir1/*"))
		assert.ElementsMatch(t, []string{"/dir1/", "/dir2/"}, globFile("/*"))
		assert.ElementsMatch(t, []string{"/"}, globFile("/"))
	})

	// GetFileGlobOrder checks that GetFile(glob) streams data back in the
	// right order. GetFile(glob) is supposed to return a stream of data of the
	// form file1 + file2 + .. + fileN, where file1 is the lexicographically lowest
	// file matching 'glob', file2 is the next lowest, etc.
	suite.Run("GetFileTARGlobOrder", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		var expected bytes.Buffer
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		for i := 0; i < 25; i++ {
			next := fmt.Sprintf("%d,%d,%d,%d\n", 4*i, (4*i)+1, (4*i)+2, (4*i)+3)
			expected.WriteString(next)
			require.NoError(t, env.PachClient.PutFile(commit, fmt.Sprintf("/data/%010d", i), strings.NewReader(next)))
		}
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))

		var output bytes.Buffer
		rc, err := env.PachClient.GetFileTAR(commit, "/data/*")
		require.NoError(t, err)
		defer rc.Close()
		require.NoError(t, tarutil.ConcatFileContent(&output, rc))

		require.Equal(t, expected.String(), output.String())
	})

	suite.Run("PathRange", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		masterCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		var paths []string
		for i := 0; i < 3; i++ {
			paths = append(paths, fmt.Sprintf("/dir%v/", i))
			for j := 0; j < 3; j++ {
				path := fmt.Sprintf("/dir%v/file%v", i, j)
				require.NoError(t, env.PachClient.PutFile(masterCommit, path, &bytes.Buffer{}))
				paths = append(paths, path)
			}
		}

		type test struct {
			pathRange     *pfs.PathRange
			expectedPaths []string
		}
		var tests []*test
		for i := 1; i < len(paths); i++ {
			for j := 0; j+i < len(paths); j += i {
				tests = append(tests, &test{
					pathRange: &pfs.PathRange{
						Lower: paths[j],
						Upper: paths[j+i],
					},
					expectedPaths: paths[j : j+i],
				})
			}
		}
		for i, test := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				c, err := env.PachClient.PfsAPIClient.GlobFile(ctx, &pfs.GlobFileRequest{
					Commit:    masterCommit,
					Pattern:   "**",
					PathRange: test.pathRange,
				})
				require.NoError(t, err)
				expectedPaths := test.expectedPaths
				require.NoError(t, grpcutil.ForEach[*pfs.FileInfo](c, func(fi *pfs.FileInfo) error {
					require.Equal(t, expectedPaths[0], fi.File.Path)
					expectedPaths = expectedPaths[1:]
					return nil
				}))
				require.Equal(t, 0, len(expectedPaths))
			})
		}
	})

	suite.Run("ApplyWriteOrder", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		if testing.Short() {
			t.Skip("Skipping integration tests in short mode")
		}

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		// Test that fails when records are applied in lexicographic order
		// rather than mod revision order.
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "/file", strings.NewReader("")))
		err = env.PachClient.DeleteFile(commit, "/")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		fileInfos, err := env.PachClient.GlobFileAll(commit, "**")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))
	})

	suite.Run("Overwrite", func(t *testing.T) {
		// TODO(2.0 optional): Implement put file split.
		t.Skip("Put file split not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	if testing.Short() {
		//		t.Skip("Skipping integration tests in short mode")
		//	}
		//
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//
		//	// Write foo
		//	_, err := env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFile(repo, "master", "file1", strings.NewReader("foo"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, "master", "file2", pfs.Delimiter_LINE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, "master", "file3", pfs.Delimiter_LINE, 0, 0, 0, false, strings.NewReader("foo\nbar\nbuz\n"))
		//	require.NoError(t, err)
		//	require.NoError(t, finishCommit(env.PachClient, repo, "master"))
		//	_, err = env.PachClient.StartProjectCommit(pfs.DefaultProjectName,repo, "master")
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFile(repo, "master", "file1", strings.NewReader("bar"))
		//	require.NoError(t, err)
		//	require.NoError(t, env.PachClient.PutFile(repo, "master", "file2", strings.NewReader("buzz")))
		//	require.NoError(t, err)
		//	_, err = env.PachClient.PutFileSplit(repo, "master", "file3", pfs.Delimiter_LINE, 0, 0, 0, true, strings.NewReader("0\n1\n2\n"))
		//	require.NoError(t, err)
		//	require.NoError(t, finishCommit(env.PachClient, repo, "master"))
		//	var buffer bytes.Buffer
		//	require.NoError(t, env.PachClient.GetFile(repo, "master", "file1", &buffer))
		//	require.Equal(t, "bar", buffer.String())
		//	buffer.Reset()
		//	require.NoError(t, env.PachClient.GetFile(repo, "master", "file2", &buffer))
		//	require.Equal(t, "buzz", buffer.String())
		//	fileInfos, err := env.PachClient.ListFileAll(repo, "master", "file3")
		//	require.NoError(t, err)
		//	require.Equal(t, 3, len(fileInfos))
		//	for i := 0; i < 3; i++ {
		//		buffer.Reset()
		//		require.NoError(t, env.PachClient.GetFile(repo, "master", fmt.Sprintf("file3/%016x", i), &buffer))
		//		require.Equal(t, fmt.Sprintf("%d\n", i), buffer.String())
		//	}
	})

	suite.Run("FindCommits", func(t *testing.T) {
		t.Parallel()
		ctx, cf := context.WithTimeout(pctx.TestContext(t), time.Minute)
		defer cf()
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "/files/a", strings.NewReader("foo")))
		require.NoError(t, env.PachClient.PutFile(commit1, "/files/b", strings.NewReader("bar")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "/files/c", strings.NewReader("baz")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		commit3, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit3, "/files/b", strings.NewReader("foo")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit3.Branch.Name, commit3.Id))

		commit4, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.DeleteFile(commit4, "/files/b"))
		require.NoError(t, finishCommit(env.PachClient, repo, commit4.Branch.Name, commit4.Id))

		resp, err := env.PachClient.FindCommits(&pfs.FindCommitsRequest{FilePath: "/files/b", Start: commit4, Limit: 0})
		require.NoError(t, err)

		require.Equal(t, []*pfs.Commit{commit4, commit3, commit1}, resp.FoundCommits)
		require.Equal(t, uint32(4), resp.CommitsSearched)
		require.Equal(t, commit1, resp.LastSearchedCommit)
	})

	suite.Run("FindCommitsLimit", func(t *testing.T) {
		t.Parallel()
		ctx, cf := context.WithTimeout(pctx.TestContext(t), time.Minute)
		defer cf()
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "/files/a", strings.NewReader("foo")))
		require.NoError(t, env.PachClient.PutFile(commit1, "/files/b", strings.NewReader("bar")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit2, "/files/b", strings.NewReader("foo")))
		require.NoError(t, finishCommit(env.PachClient, repo, commit2.Branch.Name, commit2.Id))

		resp, err := env.PachClient.FindCommits(&pfs.FindCommitsRequest{FilePath: "/files/b", Start: commit2, Limit: 1})
		require.NoError(t, err)

		require.Equal(t, []*pfs.Commit{commit2}, resp.FoundCommits)
		require.Equal(t, uint32(1), resp.CommitsSearched)
		require.Equal(t, commit2, resp.LastSearchedCommit)
	})

	suite.Run("FindCommitsOpenCommit", func(t *testing.T) {
		t.Parallel()
		ctx, cf := context.WithTimeout(pctx.TestContext(t), time.Minute)
		defer cf()
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		_, err = env.PachClient.FindCommits(&pfs.FindCommitsRequest{FilePath: "/files/b", Start: commit1, Limit: 1})
		require.YesError(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok, "returned error must be a gRPC status")
		var sawResourceInfo bool
		for _, d := range s.Details() {
			switch d := d.(type) {
			case *errdetails.ResourceInfo:
				require.Equal(t, d.ResourceType, "pfs:commit")
				require.Equal(t, d.ResourceName, commit1.Id)
				require.Equal(t, d.Description, "commit not finished")
				sawResourceInfo = true
			}
		}
		require.True(t, sawResourceInfo, "must have seen resource info in error details")
	})

	suite.Run("CopyFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		masterCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		numFiles := 5
		for i := 0; i < numFiles; i++ {
			require.NoError(t, env.PachClient.PutFile(masterCommit, fmt.Sprintf("files/%d", i), strings.NewReader(fmt.Sprintf("foo %d\n", i))))
		}
		require.NoError(t, finishCommit(env.PachClient, repo, masterCommit.Branch.Name, masterCommit.Id))

		for i := 0; i < numFiles; i++ {
			_, err = env.PachClient.InspectFile(masterCommit, fmt.Sprintf("files/%d", i))
			require.NoError(t, err)
		}

		otherCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "other")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.CopyFile(otherCommit, "files", masterCommit, "files", client.WithAppendCopyFile()))
		require.NoError(t, env.PachClient.CopyFile(otherCommit, "file0", masterCommit, "files/0", client.WithAppendCopyFile()))
		require.NoError(t, env.PachClient.CopyFile(otherCommit, "all", masterCommit, "/", client.WithAppendCopyFile()))
		require.NoError(t, finishCommit(env.PachClient, repo, otherCommit.Branch.Name, otherCommit.Id))

		for i := 0; i < numFiles; i++ {
			_, err = env.PachClient.InspectFile(otherCommit, fmt.Sprintf("files/%d", i))
			require.NoError(t, err)
		}
		for i := 0; i < numFiles; i++ {
			_, err = env.PachClient.InspectFile(otherCommit, fmt.Sprintf("all/files/%d", i))
			require.NoError(t, err)
		}
		_, err = env.PachClient.InspectFile(otherCommit, "files/0")
		require.NoError(t, err)
	})

	suite.Run("PropagateBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo1 := "test1"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo1))
		repo2 := "test2"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo2))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo2, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, repo1, "master")}))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo1, "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, repo1, commit.Branch.Name, commit.Id))
		commits, err := env.PachClient.ListCommitByRepo(client.NewRepo(pfs.DefaultProjectName, repo2))
		require.NoError(t, err)
		require.Equal(t, 2, len(commits))
	})

	// BackfillBranch implements the following DAG:
	//
	// A ──▶ C
	//  ╲   ◀
	//   ╲ ╱
	//    ╳
	//   ╱ ╲
	// 	╱   ◀
	// B ──▶ D
	suite.Run("BackfillBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "D"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master"), client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "B", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		commits, err := env.PachClient.ListCommitByRepo(client.NewRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, err)
		require.Equal(t, 3, len(commits))

		// Create a branch in D, it should receive a single commit for the heads of `A` and `B`.
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "D", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master"), client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		commits, err = env.PachClient.ListCommitByRepo(client.NewRepo(pfs.DefaultProjectName, "D"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commits))
	})

	// UpdateBranch tests the following DAG:
	//
	// A ─▶ B ─▶ C
	//
	// Then updates it to:
	//
	// A ─▶ B ─▶ C
	//      ▲
	// D ───╯
	//
	suite.Run("UpdateBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "D"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "D", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "D", "master", ""))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master"), client.NewBranch(pfs.DefaultProjectName, "D", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))
		cCommitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "C", "master", "")
		require.NoError(t, err)

		commitInfos, err := env.PachClient.InspectCommitSet(cCommitInfo.Commit.Id)
		require.NoError(t, err)
		require.Equal(t, 4, len(commitInfos))
	})

	suite.Run("BranchProvenance", func(t *testing.T) {
		t.Parallel()

		tests := [][]struct {
			name       string
			directProv []string
			err        bool
			expectProv map[string][]string
			expectSubv map[string][]string
		}{{
			{name: "A"},
			{name: "B", directProv: []string{"A"}},
			{name: "C", directProv: []string{"B"}},
			{name: "D", directProv: []string{"C", "A"},
				expectProv: map[string][]string{"A": nil, "B": {"A"}, "C": {"B", "A"}, "D": {"A", "B", "C"}},
				expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
			// A ─▶ B ─▶ C ─▶ D
			// ╰─────────────⬏
			{name: "B",
				expectProv: map[string][]string{"A": {}, "B": {}, "C": {"B"}, "D": {"A", "B", "C"}},
				expectSubv: map[string][]string{"A": {"D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
			// A    B ─▶ C ─▶ D
			// ╰─────────────⬏
		}, {
			{name: "A"},
			{name: "B", directProv: []string{"A"}},
			{name: "C", directProv: []string{"A", "B"}},
			{name: "D", directProv: []string{"C"},
				expectProv: map[string][]string{"A": {}, "B": {"A"}, "C": {"A", "B"}, "D": {"A", "B", "C"}},
				expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
			// A ─▶ B ─▶ C ─▶ D
			// ╰────────⬏
			{name: "C", directProv: []string{"B"},
				expectProv: map[string][]string{"A": {}, "B": {"A"}, "C": {"A", "B"}, "D": {"A", "B", "C"}},
				expectSubv: map[string][]string{"A": {"B", "C", "D"}, "B": {"C", "D"}, "C": {"D"}, "D": {}}},
			// A ─▶ B ─▶ C ─▶ D
		}, {
			{name: "A"},
			{name: "B"},
			{name: "C", directProv: []string{"A", "B"}},
			{name: "D", directProv: []string{"C"}},
			{name: "E", directProv: []string{"A", "D"},
				expectProv: map[string][]string{"A": {}, "B": {}, "C": {"A", "B"}, "D": {"A", "B", "C"}, "E": {"A", "B", "C", "D"}},
				expectSubv: map[string][]string{"A": {"C", "D", "E"}, "B": {"C", "D", "E"}, "C": {"D", "E"}, "D": {"E"}, "E": {}}},
			// A    B ─▶ C ─▶ D ─▶ E
			// ├────────⬏          ▲
			// ╰───────────────────╯
			{name: "C", directProv: []string{"B"},
				expectProv: map[string][]string{"A": {}, "B": {}, "C": {"B"}, "D": {"B", "C"}, "E": {"A", "B", "C", "D"}},
				expectSubv: map[string][]string{"A": {"E"}, "B": {"C", "D", "E"}, "C": {"D", "E"}, "D": {"E"}, "E": {}}},
			// A    B ─▶ C ─▶ D ─▶ E
			// ╰──────────────────⬏
		},
		// {
		// 	{name: "A", directProv: []string{"A"}, err: true},
		// 	{name: "A"},
		// 	{name: "A", directProv: []string{"A"}, err: true},
		// 	{name: "B", directProv: []string{"A"}},
		// 	{name: "A", directProv: []string{"B"}, err: true},
		// },
		}
		for i, test := range tests {
			test := test
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				t.Parallel()
				ctx := pctx.TestContext(t)
				env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

				for _, repo := range []string{"A", "B", "C", "D", "E"} {
					require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
				}
				for iStep, step := range test {
					var provenance []*pfs.Branch
					for _, repo := range step.directProv {
						provenance = append(provenance, client.NewBranch(pfs.DefaultProjectName, repo, "master"))
					}
					err := env.PachClient.CreateBranch(pfs.DefaultProjectName, step.name, "master", "", "", provenance)
					if step.err {
						require.YesError(t, err, "%d> CreateBranch(\"%s\", %v)", iStep, step.name, step.directProv)
					} else {
						require.NoError(t, err, "%d> CreateBranch(\"%s\", %v)", iStep, step.name, step.directProv)
					}
					require.NoError(t, env.PachClient.FsckFastExit())
					for repo, expectedProv := range step.expectProv {
						bi, err := env.PachClient.InspectBranch(pfs.DefaultProjectName, repo, "master")
						require.NoError(t, err)
						sort.Strings(expectedProv)
						require.Equal(t, len(expectedProv), len(bi.Provenance))
						for _, b := range bi.Provenance {
							i := sort.SearchStrings(expectedProv, b.Repo.Name)
							if i >= len(expectedProv) || expectedProv[i] != b.Repo.Name {
								t.Fatalf("provenance for %s contains: %s, but should only contain: %v", repo, b.Repo.Name, expectedProv)
							}
						}
					}
					for repo, expectedSubv := range step.expectSubv {
						bi, err := env.PachClient.InspectBranch(pfs.DefaultProjectName, repo, "master")
						require.NoError(t, err)
						sort.Strings(expectedSubv)
						require.Equal(t, len(expectedSubv), len(bi.Subvenance))
						for _, b := range bi.Subvenance {
							i := sort.SearchStrings(expectedSubv, b.Repo.Name)
							if i >= len(expectedSubv) || expectedSubv[i] != b.Repo.Name {
								t.Fatalf("subvenance for %s contains: %s, but should only contain: %v", repo, b.Repo.Name, expectedSubv)
							}
						}
					}
				}
			})
		}

		// t.Run("1", func(t *testing.T) {
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"A"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"B"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"C"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"D"))

		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"B", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"A", "master")}))
		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"C", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"B", "master")}))
		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"D", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"C", "master"), client.NewProjectBranch(pfs.DefaultProjectName,"A", "master")}))

		// 	aMaster, err := env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"A", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 3, len(aMaster.Subvenance))

		// 	cMaster, err := env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"C", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 2, len(cMaster.Provenance))

		// 	dMaster, err := env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"D", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 3, len(dMaster.Provenance))

		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"B", "master", "", nil))

		// 	aMaster, err = env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"A", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 1, len(aMaster.Subvenance))

		// 	cMaster, err = env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"C", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 1, len(cMaster.Provenance))

		// 	dMaster, err = env.PachClient.InspectProjectBranch(pfs.DefaultProjectName,"D", "master")
		// 	require.NoError(t, err)
		// 	require.Equal(t, 3, len(dMaster.Provenance))
		// })
		// t.Run("2", func(t *testing.T) {
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"A"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"B"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"C"))
		// 	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"D"))

		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"B", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"A", "master")}))
		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"C", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"B", "master"), client.NewProjectBranch(pfs.DefaultProjectName,"A", "master")}))
		// 	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,"D", "master", "", []*pfs.Branch{client.NewProjectBranch(pfs.DefaultProjectName,"C", "master")}))
		// })
	})

	suite.Run("ChildCommits", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "A", "master", "", "", nil))
		aRepo := client.NewRepo(pfs.DefaultProjectName, "A")

		// Small helper function wrapping env.PachClient.InspectCommit, because it's called a lot
		inspect := func(repo, branch, commit string) *pfs.CommitInfo {
			commitInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, repo, branch, commit)
			require.NoError(t, err)
			return commitInfo
		}

		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		commits, err := env.PachClient.ListCommit(aRepo, aRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		t.Logf("%v", commits)
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))

		commit2, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)

		// Inspect commit 1 and 2
		commit1Info, commit2Info := inspect("A", commit1.Branch.Name, commit1.Id), inspect("A", commit2.Branch.Name, commit2.Id)
		require.Equal(t, commit1.Id, commit2Info.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{commit2}, commit1Info.ChildCommits, CommitToID)

		// Delete commit 2 and make sure it's removed from commit1.ChildCommits
		require.NoError(t, env.PachClient.DropCommitSet(commit2.Id))
		commit1Info = inspect("A", commit1.Branch.Name, commit1.Id)
		require.ElementsEqualUnderFn(t, nil, commit1Info.ChildCommits, CommitToID)

		// Re-create commit2, and create a third commit also extending from commit1.
		// Make sure both appear in commit1.children
		commit2, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit2.Branch.Name, commit2.Id))
		commit3, err := env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(pfs.DefaultProjectName, "A", "foo"),
			Parent: commit1,
		})
		require.NoError(t, err)
		commit1Info = inspect("A", commit1.Branch.Name, commit1.Id)
		require.ImagesEqual(t, []*pfs.Commit{commit2, commit3}, commit1Info.ChildCommits, CommitToID)

		// Delete commit3 and make sure commit1 has the right children
		require.NoError(t, env.PachClient.DropCommitSet(commit3.Id))
		commit1Info = inspect("A", commit1.Branch.Name, commit1.Id)
		require.ImagesEqual(t, []*pfs.Commit{commit2}, commit1Info.ChildCommits, CommitToID)

		// Create a downstream branch in a different repo, then commit to "A" and
		// make sure the new HEAD commit is in the parent's children (i.e. test
		// propagateBranches)
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "A", "master"),
		}))
		bCommit1 := inspect("B", "master", "")
		commit3, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit3.Branch.Name, commit3.Id))
		// Re-inspect bCommit1, which has been updated by StartCommit
		bCommit1, bCommit2 := inspect("B", bCommit1.Commit.Branch.Name, bCommit1.Commit.Id), inspect("B", "master", "")
		require.Equal(t, bCommit1.Commit.Id, bCommit2.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{bCommit2.Commit}, bCommit1.ChildCommits, CommitToID)

		// create a new branch in a different repo, then update it so that two commits
		// are generated. Make sure the second commit is in the parent's children
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "A", "master"),
		}))
		cCommit1 := inspect("C", "master", "") // Get new commit's ID
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "A", "master"),
			client.NewBranch(pfs.DefaultProjectName, "B", "master"),
		}))
		// Re-inspect cCommit1, which has been updated by CreateBranch
		cCommit1, cCommit2 := inspect("C", cCommit1.Commit.Branch.Name, cCommit1.Commit.Id), inspect("C", "master", "")
		require.Equal(t, cCommit1.Commit.Id, cCommit2.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{cCommit2.Commit}, cCommit1.ChildCommits, CommitToID)
	})

	suite.Run("StartCommitFork", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit.Branch.Name, commit.Id))
		commit2, err := env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(pfs.DefaultProjectName, "A", "master2"),
			Parent: client.NewCommit(pfs.DefaultProjectName, "A", "master", ""),
		})
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit2.Branch.Name, commit2.Id))

		aRepo := client.NewRepo(pfs.DefaultProjectName, "A")
		commitInfos, err := env.PachClient.ListCommit(aRepo, nil, nil, 0)
		require.NoError(t, err)
		commits := []*pfs.Commit{}
		for _, ci := range commitInfos {
			commits = append(commits, ci.Commit)
		}
		require.ImagesEqual(t, []*pfs.Commit{commit, commit2}, commits, CommitToID)
	})

	// UpdateBranchNewOutputCommit tests the following corner case:
	// A ──▶ C
	// B
	//
	// Becomes:
	//
	// A  ╭▶ C
	// B ─╯
	//
	// C should create a new output commit to process its unprocessed inputs in B
	suite.Run("UpdateBranchNewOutputCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "A", "master", "", "", nil))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", nil))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "",
			[]*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		// Create commits in A and B
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "A", commit.Branch.Name, commit.Id))
		commit, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "B", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "B", commit.Branch.Name, commit.Id))
		// Check for first output commit in C (plus the old empty head commit)
		cRepo := client.NewRepo(pfs.DefaultProjectName, "C")
		commits, err := env.PachClient.ListCommit(cRepo, cRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(commits))
		// Update the provenance of C/master and make sure it creates a new commit
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "",
			[]*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		commits, err = env.PachClient.ListCommit(cRepo, cRepo.NewCommit("master", ""), nil, 0)
		require.NoError(t, err)
		require.Equal(t, 3, len(commits))
	})

	// SquashCommitSetMultipleChildrenSingleCommit tests that when you have the
	// following commit graph in a repo:
	// c   d
	//  ↘ ↙
	//   b
	//   ↓
	//   a
	//
	// and you delete commit 'b', what you end up with is:
	//
	// c   d
	//  ↘ ↙
	//   a
	suite.Run("SquashCommitSetMultipleChildrenSingleCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "repo", "master", "", "", nil))

		// Create commits 'a' and 'b'
		a, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "repo", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "repo", a.Branch.Name, a.Id))
		b, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "repo", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "repo", b.Branch.Name, b.Id))

		// Create 'd' by aliasing 'b' into another branch (force a new CommitSet rather than extending 'b')
		_, err = env.PachClient.PfsAPIClient.CreateBranch(env.PachClient.Ctx(), &pfs.CreateBranchRequest{
			Branch:       client.NewBranch(pfs.DefaultProjectName, "repo", "master2"),
			Head:         client.NewCommit(pfs.DefaultProjectName, "repo", "master", ""),
			NewCommitSet: true,
		})
		require.NoError(t, err)
		d, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "repo", "master2")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "repo", d.Branch.Name, d.Id))

		// Create 'c'
		c, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "repo", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "repo", c.Branch.Name, c.Id))

		// Collect info re: a, b, c, and d, and make sure that the parent/child
		// relationships are all correct
		aInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		bInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", b.Branch.Name, b.Id)
		require.NoError(t, err)
		cInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", c.Branch.Name, c.Id)
		require.NoError(t, err)
		dInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", "master2", "")
		require.NoError(t, err)

		require.NotNil(t, aInfo.ParentCommit) // this should be the empty default head
		require.ImagesEqual(t, []*pfs.Commit{b}, aInfo.ChildCommits, CommitToID)

		require.Equal(t, a.Id, bInfo.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{c, d}, bInfo.ChildCommits, CommitToID)

		require.Equal(t, b.Id, cInfo.ParentCommit.Id)
		require.Equal(t, 0, len(cInfo.ChildCommits))

		require.Equal(t, b.Id, dInfo.ParentCommit.Id)
		require.Equal(t, 0, len(dInfo.ChildCommits))

		// Delete commit 'b'
		require.NoError(t, env.PachClient.SquashCommitSet(b.Id))

		// Collect info re: a, c, and d, and make sure that the parent/child
		// relationships are still correct
		aInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		cInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", c.Branch.Name, c.Id)
		require.NoError(t, err)
		dInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", d.Branch.Name, d.Id)
		require.NoError(t, err)

		require.NotNil(t, aInfo.ParentCommit)
		require.ImagesEqual(t, []*pfs.Commit{c, d}, aInfo.ChildCommits, CommitToID)

		require.Equal(t, a.Id, cInfo.ParentCommit.Id)
		require.Equal(t, 0, len(cInfo.ChildCommits))

		require.Equal(t, a.Id, dInfo.ParentCommit.Id)
		require.Equal(t, 0, len(dInfo.ChildCommits))
	})

	// Tests that when you have the following commit graph in a *downstream* repo:
	//
	//    ↙f
	//   c
	//   ↓↙e
	//   b
	//   ↓↙d
	//   a
	//
	// and you delete commits 'b', what you end up with
	// is:
	//     f
	//     ↓
	// d e c
	//  ↘↓↙
	//   a
	// This makes sure that multiple live children are re-pointed at a live parent
	// if appropriate
	suite.Run("SquashCommitSetMultiLevelChildrenSimple", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		project := tu.UniqueString("prj-")
		require.NoError(t, env.PachClient.CreateProject(project))
		// Create main repo (will have the commit graphs above)
		require.NoError(t, env.PachClient.CreateRepo(project, "repo"))
		repoProto := client.NewRepo(project, "repo")
		_, err := env.PachClient.StartCommit(project, "repo", "master")
		require.NoError(t, err)
		// Create commit 'a'
		aInfo, err := env.PachClient.InspectCommit(project, "repo", "master", "")
		require.NoError(t, err)
		a := aInfo.Commit
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", a.Branch.Name, a.Id))
		// Create 'd'
		resp, err := env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(project, "repo", "fod"),
			Parent: a,
		})
		require.NoError(t, err)
		d := client.NewCommit(project, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", resp.Branch.Name, resp.Id))
		// Create 'b'
		// (a & b have same prov commit in upstream2, so this is the commit that will
		// be deleted, as both b and c are provenant on it)
		squashMeCommit, err := env.PachClient.StartCommit(project, "repo", "master")
		require.NoError(t, err)
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", "master", ""))
		bInfo, err := env.PachClient.InspectCommit(project, "repo", "master", "")
		require.NoError(t, err)
		b := bInfo.Commit
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", b.Branch.Name, b.Id))
		require.Equal(t, b.Id, squashMeCommit.Id)
		// Create 'e'
		resp, err = env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(project, "repo", "foe"),
			Parent: b,
		})
		require.NoError(t, err)
		e := client.NewCommit(project, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", resp.Branch.Name, resp.Id))
		// Create 'c'
		_, err = env.PachClient.StartCommit(project, "repo", "master")
		require.NoError(t, err)
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", "master", ""))
		cInfo, err := env.PachClient.InspectCommit(project, "repo", "master", "")
		require.NoError(t, err)
		c := cInfo.Commit
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", c.Branch.Name, c.Id))
		// Create 'f'
		resp, err = env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(project, "repo", "fof"),
			Parent: c,
		})
		require.NoError(t, err)
		f := client.NewCommit(project, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishProjectCommit(env.PachClient, project, "repo", resp.Branch.Name, resp.Id))
		// Make sure child/parent relationships are as shown in first diagram
		commits, err := env.PachClient.ListCommit(repoProto, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 6, len(commits))
		aInfo, err = env.PachClient.InspectCommit(project, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		bInfo, err = env.PachClient.InspectCommit(project, "repo", b.Branch.Name, b.Id)
		require.NoError(t, err)
		cInfo, err = env.PachClient.InspectCommit(project, "repo", c.Branch.Name, c.Id)
		require.NoError(t, err)
		dInfo, err := env.PachClient.InspectCommit(project, "repo", d.Branch.Name, d.Id)
		require.NoError(t, err)
		eInfo, err := env.PachClient.InspectCommit(project, "repo", e.Branch.Name, e.Id)
		require.NoError(t, err)
		fInfo, err := env.PachClient.InspectCommit(project, "repo", f.Branch.Name, f.Id)
		require.NoError(t, err)
		require.Nil(t, aInfo.ParentCommit)
		require.Equal(t, a.Id, bInfo.ParentCommit.Id)
		require.Equal(t, a.Id, dInfo.ParentCommit.Id)
		require.Equal(t, b.Id, cInfo.ParentCommit.Id)
		require.Equal(t, b.Id, eInfo.ParentCommit.Id)
		require.Equal(t, c.Id, fInfo.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{b, d}, aInfo.ChildCommits, CommitToID)
		require.ImagesEqual(t, []*pfs.Commit{c, e}, bInfo.ChildCommits, CommitToID)
		require.ImagesEqual(t, []*pfs.Commit{f}, cInfo.ChildCommits, CommitToID)
		require.Nil(t, dInfo.ChildCommits)
		require.Nil(t, eInfo.ChildCommits)
		require.Nil(t, fInfo.ChildCommits)
		// Delete b
		require.NoError(t, env.PachClient.SquashCommitSet(squashMeCommit.Id))
		// Re-read commit info to get new parents/children
		aInfo, err = env.PachClient.InspectCommit(project, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		cInfo, err = env.PachClient.InspectCommit(project, "repo", c.Branch.Name, c.Id)
		require.NoError(t, err)
		dInfo, err = env.PachClient.InspectCommit(project, "repo", d.Branch.Name, d.Id)
		require.NoError(t, err)
		eInfo, err = env.PachClient.InspectCommit(project, "repo", e.Branch.Name, e.Id)
		require.NoError(t, err)
		fInfo, err = env.PachClient.InspectCommit(project, "repo", f.Branch.Name, f.Id)
		require.NoError(t, err)
		// The head of master should be 'c'
		// Make sure child/parent relationships are as shown in second diagram. Note
		// that after 'b' is deleted, SquashCommitSet does not create a new commit (c has
		// an alias for the deleted commit in upstream1)
		commits, err = env.PachClient.ListCommit(client.NewRepo(project, "repo"), nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 5, len(commits))
		require.Nil(t, aInfo.ParentCommit)
		require.Equal(t, a.Id, cInfo.ParentCommit.Id)
		require.Equal(t, a.Id, dInfo.ParentCommit.Id)
		require.Equal(t, a.Id, eInfo.ParentCommit.Id)
		require.Equal(t, c.Id, fInfo.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{d, e, c}, aInfo.ChildCommits, CommitToID)
		require.ImagesEqual(t, []*pfs.Commit{f}, cInfo.ChildCommits, CommitToID)
		require.Nil(t, dInfo.ChildCommits)
		require.Nil(t, eInfo.ChildCommits)
		require.Nil(t, fInfo.ChildCommits)
		masterInfo, err := env.PachClient.InspectBranch(project, "repo", "master")
		require.NoError(t, err)
		require.Equal(t, c.Id, masterInfo.Head.Id)
	})

	// Tests that when you have the following commit graph in a *downstream* repo:
	//
	//   g
	//   ↓↙f
	//   c
	//   ↓↙e
	//   b
	//   ↓↙d
	//   a
	//
	// and you delete commits 'b' and 'c' (in a single call), what you end up with
	// is:
	//
	// d e f g
	//  ↘↓↙ ↙
	//   a
	// This makes sure that multiple live children are re-pointed at a live parent
	// if appropriate
	suite.Run("SquashCommitSetMultiLevelChildrenComplex", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(pctx.TestContext(t), t, dockertestenv.NewTestDBConfig(t))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "upstream1"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "upstream2"))
		// commit to both inputs
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "upstream1", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "upstream1", "master", ""))
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "upstream2", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "upstream2", "master", ""))
		// Create main repo (will have the commit graphs above)
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "repo", "master", "", "", []*pfs.Branch{
			client.NewBranch(pfs.DefaultProjectName, "upstream1", "master"),
			client.NewBranch(pfs.DefaultProjectName, "upstream2", "master"),
		}))
		repoProto := client.NewRepo(pfs.DefaultProjectName, "repo")
		// Create commit 'a'
		aInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
		require.NoError(t, err)
		a := aInfo.Commit
		require.NoError(t, finishCommit(env.PachClient, "repo", a.Branch.Name, a.Id))
		// Create 'd'
		resp, err := env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(pfs.DefaultProjectName, "repo", "fod"),
			Parent: a,
		})
		require.NoError(t, err)
		d := client.NewCommit(pfs.DefaultProjectName, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishCommit(env.PachClient, "repo", resp.Branch.Name, resp.Id))
		// Create 'b'
		// (a & b have same prov commit in upstream2, so this is the commit that will
		// be deleted, as both b and c are provenant on it)
		bSquashMeCommit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "upstream1", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "upstream1", "master", ""))
		bInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
		require.NoError(t, err)
		b := bInfo.Commit
		require.NoError(t, finishCommit(env.PachClient, "repo", b.Branch.Name, b.Id))
		require.Equal(t, b.Id, bSquashMeCommit.Id)
		// Create 'e'
		resp, err = env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(pfs.DefaultProjectName, "repo", "foe"),
			Parent: b,
		})
		require.NoError(t, err)
		e := client.NewCommit(pfs.DefaultProjectName, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishCommit(env.PachClient, "repo", resp.Branch.Name, resp.Id))
		// Create 'c'
		_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, "upstream2", "master")
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "upstream2", "master", ""))
		cInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
		require.NoError(t, err)
		cSquashMeToo := cInfo.Commit
		require.NoError(t, finishCommit(env.PachClient, "repo", cSquashMeToo.Branch.Name, cSquashMeToo.Id))
		// Create 'f'
		resp, err = env.PachClient.PfsAPIClient.StartCommit(env.PachClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch(pfs.DefaultProjectName, "repo", "fof"),
			Parent: cSquashMeToo,
		})
		require.NoError(t, err)
		f := client.NewCommit(pfs.DefaultProjectName, "repo", resp.Branch.Name, resp.Id)
		require.NoError(t, finishCommit(env.PachClient, "repo", resp.Branch.Name, resp.Id))
		// Make sure child/parent relationships are as shown in first diagram
		commits, err := env.PachClient.ListCommit(repoProto, nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 6, len(commits))
		aInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		bInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", b.Branch.Name, b.Id)
		require.NoError(t, err)
		cInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", cSquashMeToo.Branch.Name, cSquashMeToo.Id)
		require.NoError(t, err)
		dInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", d.Branch.Name, d.Id)
		require.NoError(t, err)
		eInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", e.Branch.Name, e.Id)
		require.NoError(t, err)
		fInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", f.Branch.Name, f.Id)
		require.NoError(t, err)
		require.Nil(t, aInfo.ParentCommit)
		require.Equal(t, a.Id, bInfo.ParentCommit.Id)
		require.Equal(t, a.Id, dInfo.ParentCommit.Id)
		require.Equal(t, b.Id, cInfo.ParentCommit.Id)
		require.Equal(t, b.Id, eInfo.ParentCommit.Id)
		require.Equal(t, cSquashMeToo.Id, fInfo.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{b, d}, aInfo.ChildCommits, CommitToID)
		require.ImagesEqual(t, []*pfs.Commit{cSquashMeToo, e}, bInfo.ChildCommits, CommitToID)
		require.ImagesEqual(t, []*pfs.Commit{f}, cInfo.ChildCommits, CommitToID)
		require.Nil(t, dInfo.ChildCommits)
		require.Nil(t, eInfo.ChildCommits)
		require.Nil(t, fInfo.ChildCommits)
		// Delete second commit in upstream1, which deletes b
		// didn't work, because commit set c depends on commit set b via repo.'c' on upstream1.'b'
		require.YesError(t, env.PachClient.SquashCommitSet(bSquashMeCommit.Id))
		// now squash commit sets c
		require.YesError(t, env.PachClient.SquashCommitSet(cSquashMeToo.Id))
		// still doesn't work, because we would be deleting the heads of upstream1 & upstream2. To successfully squash, we need to push another commit set.
		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)
		txnClient := env.PachClient.WithTransaction(txn)
		_, err = txnClient.StartCommit(pfs.DefaultProjectName, "upstream1", "master")
		require.NoError(t, err)
		_, err = txnClient.StartCommit(pfs.DefaultProjectName, "upstream2", "master")
		require.NoError(t, err)
		_, err = env.PachClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.NoError(t, finishCommit(env.PachClient, "upstream1", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "upstream2", "master", ""))
		require.NoError(t, finishCommit(env.PachClient, "repo", "master", ""))
		// try squashing b & c again
		// now squash commit sets b & c
		require.NoError(t, env.PachClient.SquashCommitSet(cSquashMeToo.Id))
		require.NoError(t, env.PachClient.SquashCommitSet(bSquashMeCommit.Id))
		// Re-read commit info to get new parents/children
		aInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", a.Branch.Name, a.Id)
		require.NoError(t, err)
		dInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", d.Branch.Name, d.Id)
		require.NoError(t, err)
		eInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", e.Branch.Name, e.Id)
		require.NoError(t, err)
		fInfo, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", f.Branch.Name, f.Id)
		require.NoError(t, err)
		gInfo, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "repo", "master", "")
		require.NoError(t, err)
		// The head of master should be 'c'
		// Make sure child/parent relationships are as shown in second diagram. Note
		// that after 'b' is deleted, SquashCommitSet does not create a new commit (c has
		// an alias for the deleted commit in upstream1)
		commits, err = env.PachClient.ListCommit(client.NewRepo(pfs.DefaultProjectName, "repo"), nil, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 5, len(commits))
		require.Nil(t, aInfo.ParentCommit)
		require.Equal(t, a.Id, dInfo.ParentCommit.Id)
		require.Equal(t, a.Id, eInfo.ParentCommit.Id)
		require.Equal(t, a.Id, fInfo.ParentCommit.Id)
		require.Equal(t, a.Id, gInfo.ParentCommit.Id)
		require.ImagesEqual(t, []*pfs.Commit{d, e, f, gInfo.Commit}, aInfo.ChildCommits, CommitToID)
		require.Nil(t, dInfo.ChildCommits)
		require.Nil(t, eInfo.ChildCommits)
		require.Nil(t, fInfo.ChildCommits)
		require.Nil(t, gInfo.ChildCommits)
		masterInfo, err := env.PachClient.InspectBranch(pfs.DefaultProjectName, "repo", "master")
		require.NoError(t, err)
		require.Equal(t, gInfo.Commit.Id, masterInfo.Head.Id)
	})

	suite.Run("CommitState", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		// two input repos, one with many commits (logs), and one with few (schema)
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))

		// Start a commit on A/master, this will create a non-ready commit on B/master.
		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = env.PachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: client.NewCommit(pfs.DefaultProjectName, "B", "master", ""),
			Wait:   pfs.CommitState_READY,
		})
		require.YesError(t, err)

		// Finish the commit on A/master, that will make the B/master ready.
		require.NoError(t, finishCommit(env.PachClient, "A", "master", ""))

		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = env.PachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: client.NewCommit(pfs.DefaultProjectName, "B", "master", ""),
			Wait:   pfs.CommitState_READY,
		})
		require.NoError(t, err)

		// Create a new branch C/master with A/master as provenance. It should start out ready.
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))

		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = env.PachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: client.NewCommit(pfs.DefaultProjectName, "C", "master", ""),
			Wait:   pfs.CommitState_READY,
		})
		require.NoError(t, err)
	})

	suite.Run("SubscribeStates", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "A"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "B"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "C"))

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "B", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "A", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "B", "master", ""))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "C", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "B", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "C", "master", ""))

		ctx, cancel := pctx.WithCancel(env.PachClient.Ctx())
		defer cancel()
		pachClient := env.PachClient.WithCtx(ctx)

		var readyCommitsB, readyCommitsC int64
		go func() {
			_ = pachClient.SubscribeCommit(client.NewRepo(pfs.DefaultProjectName, "B"), "master", "", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
				atomic.AddInt64(&readyCommitsB, 1)
				return nil
			})
		}()
		go func() {
			_ = pachClient.SubscribeCommit(client.NewRepo(pfs.DefaultProjectName, "C"), "master", "", pfs.CommitState_READY, func(ci *pfs.CommitInfo) error {
				atomic.AddInt64(&readyCommitsC, 1)
				return nil
			})
		}()
		_, err := pachClient.StartCommit(pfs.DefaultProjectName, "A", "master")
		require.NoError(t, err)
		require.NoError(t, pachClient.FinishCommit(pfs.DefaultProjectName, "A", "master", ""))

		require.NoErrorWithinTRetry(t, time.Second*10, func() error {
			if atomic.LoadInt64(&readyCommitsB) != 2 {
				return errors.Errorf("wrong number of ready commits")
			}
			return nil
		})

		require.NoError(t, pachClient.FinishCommit(pfs.DefaultProjectName, "B", "master", ""))

		require.NoErrorWithinTRetry(t, time.Second*10, func() error {
			if atomic.LoadInt64(&readyCommitsC) != 2 {
				return errors.Errorf("wrong number of ready commits")
			}
			return nil
		})
	})

	suite.Run("PutFileCommit", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numFiles := 25
		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		var eg errgroup.Group
		for i := 0; i < numFiles; i++ {
			i := i
			eg.Go(func() error {
				return env.PachClient.PutFile(commit, fmt.Sprintf("%d", i), strings.NewReader(fmt.Sprintf("%d", i)))
			})
		}
		require.NoError(t, eg.Wait())

		for i := 0; i < numFiles; i++ {
			var b bytes.Buffer
			require.NoError(t, env.PachClient.GetFile(commit, fmt.Sprintf("%d", i), &b))
			require.Equal(t, fmt.Sprintf("%d", i), b.String())
		}

		bi, err := env.PachClient.InspectBranch(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		eg = errgroup.Group{}
		for i := 0; i < numFiles; i++ {
			i := i
			eg.Go(func() error {
				return env.PachClient.CopyFile(commit, fmt.Sprintf("%d", (i+1)%numFiles), bi.Head, fmt.Sprintf("%d", i))
			})
		}
		require.NoError(t, eg.Wait())

		for i := 0; i < numFiles; i++ {
			var b bytes.Buffer
			require.NoError(t, env.PachClient.GetFile(commit, fmt.Sprintf("%d", (i+1)%numFiles), &b))
			require.Equal(t, fmt.Sprintf("%d", i), b.String())
		}

		eg = errgroup.Group{}
		for i := 0; i < numFiles; i++ {
			i := i
			eg.Go(func() error {
				return env.PachClient.DeleteFile(commit, fmt.Sprintf("%d", i))
			})
		}
		require.NoError(t, eg.Wait())

		fileInfos, err := env.PachClient.ListFileAll(commit, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(fileInfos))
	})

	suite.Run("PutFileCommitNilBranch", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader("file")))
	})

	suite.Run("PutFileCommitOverwrite", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		numFiles := 5
		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")

		for i := 0; i < numFiles; i++ {
			require.NoError(t, env.PachClient.PutFile(commit, "file", strings.NewReader(fmt.Sprintf("%d", i))))
		}

		var b bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commit, "file", &b))
		require.Equal(t, fmt.Sprintf("%d", numFiles-1), b.String())
	})

	suite.Run("WalkFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit, "dir/bar", strings.NewReader("bar")))
		require.NoError(t, env.PachClient.PutFile(commit, "dir/dir2/buzz", strings.NewReader("buzz")))
		require.NoError(t, env.PachClient.PutFile(commit, "foo", strings.NewReader("foo")))

		expectedPaths := []string{"/", "/dir/", "/dir/bar", "/dir/dir2/", "/dir/dir2/buzz", "/foo"}
		checks := func() {
			i := 0
			require.NoError(t, env.PachClient.WalkFile(commit, "", func(fi *pfs.FileInfo) error {
				require.Equal(t, expectedPaths[i], fi.File.Path)
				i++
				return nil
			}))
			require.Equal(t, len(expectedPaths), i)
		}
		checks()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		checks()
	})

	suite.Run("WalkFile2", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "WalkFile2"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir1/file1.2", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.1", &bytes.Buffer{}))
		require.NoError(t, env.PachClient.PutFile(commit1, "/dir2/file2.2", &bytes.Buffer{}))
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
		walkFile := func(path string) []string {
			var fis []*pfs.FileInfo
			require.NoError(t, env.PachClient.WalkFile(commit1, path, func(fi *pfs.FileInfo) error {
				fis = append(fis, fi)
				return nil
			}))
			return finfosToPaths(fis)
		}
		assert.ElementsMatch(t, []string{"/dir1/", "/dir1/file1.1", "/dir1/file1.2"}, walkFile("/dir1"))
		assert.ElementsMatch(t, []string{"/dir1/file1.1"}, walkFile("/dir1/file1.1"))
		assert.Len(t, walkFile("/"), 7)
	})

	suite.Run("WalkFileEmpty", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		latestCommit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
		checks := func() {
			cb := func(fi *pfs.FileInfo) error {
				if assert.Equal(t, fi.FileType, pfs.FileType_DIR) && assert.Equal(t, fi.File.Path, "/") {
					return nil
				}
				return errors.New("should not have returned any file results for an empty commit")
			}
			checkNotFound := func(path string) {
				err := env.PachClient.WalkFile(latestCommit, path, cb)
				require.YesError(t, err)
				s := status.Convert(err)
				require.Equal(t, codes.NotFound, s.Code())
			}
			require.NoError(t, env.PachClient.WalkFile(latestCommit, "", cb))
			require.NoError(t, env.PachClient.WalkFile(latestCommit, "/", cb))
			checkNotFound("foo")
			checkNotFound("/foo")
			checkNotFound("foo/bar")
			checkNotFound("/foo/bar")
		}

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
		checks() // Test the default empty head commit

		_, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		checks() // Test an empty open commit

		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		checks() // Test an empty closed commit
	})

	suite.Run("ReadSizeLimited", func(t *testing.T) {
		// TODO(2.0 optional): Decide on how to expose offset read.
		t.Skip("Offset read exists (inefficient), just need to decide on how to expose it in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,"test"))
		//	require.NoError(t, env.PachClient.PutFile("test", "master", "", "file", strings.NewReader(strings.Repeat("a", 100*units.MB))))
		//
		//	var b bytes.Buffer
		//	require.NoError(t, env.PachClient.GetFile("test", "master", "", "file", 0, 2*units.MB, &b))
		//	require.Equal(t, 2*units.MB, b.Len())
		//
		//	b.Reset()
		//	require.NoError(t, env.PachClient.GetFile("test", "master", "", "file", 2*units.MB, 2*units.MB, &b))
		//	require.Equal(t, 2*units.MB, b.Len())
	})

	suite.Run("PutFileURL", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		if testing.Short() {
			t.Skip("Skipping integration tests in short mode")
		}

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFileURL(commit, "readme", "https://raw.githubusercontent.com/pachyderm/pachyderm/master/README.md", false))
		check := func() {
			fileInfo, err := env.PachClient.InspectFile(commit, "readme")
			require.NoError(t, err)
			require.True(t, fileInfo.SizeBytes > 0)
		}
		check()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		check()
	})

	suite.Run("PutFilesURL", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		repoProto := client.NewRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		paths := []string{"README.md", "CHANGELOG.md", "CONTRIBUTING.md"}
		for _, path := range paths {
			url := fmt.Sprintf("https://raw.githubusercontent.com/pachyderm/pachyderm/master/%s", path)
			require.NoError(t, env.PachClient.PutFileURL(commit, path, url, false))
		}
		check := func() {
			cis, err := env.PachClient.ListCommit(repoProto, nil, nil, 0)
			require.NoError(t, err)
			require.Equal(t, 1, len(cis))

			for _, path := range paths {
				fileInfo, err := env.PachClient.InspectFile(repoProto.NewCommit("master", ""), path)
				require.NoError(t, err)
				require.True(t, fileInfo.SizeBytes > 0)
			}
		}
		check()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		check()
	})

	suite.Run("PutFilesObjURL", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		bucket, url := dockertestenv.NewTestBucket(env.Context, t)
		paths := []string{"files/foo", "files/bar", "files/fizz"}
		for _, path := range paths {
			require.NoError(t, bucket.WriteAll(ctx, path, []byte(path), nil))
		}
		for _, p := range paths {
			objURL := url + p
			require.NoError(t, env.PachClient.PutFileURL(commit, p, objURL, false))
		}
		srcURL := url + "files"
		require.NoError(t, env.PachClient.PutFileURL(commit, "recursive", srcURL, true))
		check := func() {
			cis, err := env.PachClient.ListCommit(client.NewRepo(pfs.DefaultProjectName, repo), nil, nil, 0)
			require.NoError(t, err)
			require.Equal(t, 1, len(cis))

			for _, path := range paths {
				var b bytes.Buffer
				require.NoError(t, env.PachClient.GetFile(commit, path, &b))
				require.Equal(t, path, b.String())
				b.Reset()
				require.NoError(t, env.PachClient.GetFile(commit, filepath.Join("recursive", filepath.Base(path)), &b))
				require.Equal(t, path, b.String())
			}
		}
		check()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		check()
	})

	suite.Run("GetFilesObjURL", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		paths := []string{"files/foo", "files/bar", "files/fizz"}
		for _, path := range paths {
			require.NoError(t, env.PachClient.PutFile(commit, path, strings.NewReader(path)))
		}
		check := func() {
			bucket, url := dockertestenv.NewTestBucket(ctx, t)
			for _, path := range paths {
				require.NoError(t, env.PachClient.GetFileURL(commit, path, url))
			}
			for _, path := range paths {
				data, err := bucket.ReadAll(ctx, path)
				require.NoError(t, err)
				require.True(t, bytes.Equal([]byte(path), data))
				require.NoError(t, bucket.Delete(ctx, path))
			}
			require.NoError(t, env.PachClient.GetFileURL(commit, "files/*", url))
			for _, path := range paths {
				data, err := bucket.ReadAll(ctx, path)
				require.NoError(t, err)
				require.True(t, bytes.Equal([]byte(path), data))
				require.NoError(t, bucket.Delete(ctx, path))
			}
			prefix := "/prefix"
			require.NoError(t, env.PachClient.GetFileURL(commit, "files/*", url+prefix))
			for _, path := range paths {
				data, err := bucket.ReadAll(ctx, prefix+"/"+path)
				require.NoError(t, err)
				require.True(t, bytes.Equal([]byte(path), data))
			}
		}
		check()
		require.NoError(t, finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id))
		check()
	})

	suite.Run("PutFileOutputRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		inputRepo, outputRepo := "input", "output"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, inputRepo))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, outputRepo))
		inCommit := client.NewCommit(pfs.DefaultProjectName, inputRepo, "master", "")
		outCommit := client.NewCommit(pfs.DefaultProjectName, outputRepo, "master", "")
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, outputRepo, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, inputRepo, "master")}))
		require.NoError(t, finishCommit(env.PachClient, outputRepo, "master", ""))
		require.NoError(t, env.PachClient.PutFile(inCommit, "foo", strings.NewReader("foo\n")))
		require.NoError(t, env.PachClient.PutFile(outCommit, "bar", strings.NewReader("bar\n")))
		require.NoError(t, finishCommit(env.PachClient, outputRepo, "master", ""))
		fileInfos, err := env.PachClient.ListFileAll(outCommit, "")
		require.NoError(t, err)
		require.Equal(t, 1, len(fileInfos))
		buf := &bytes.Buffer{}
		require.NoError(t, env.PachClient.GetFile(outCommit, "bar", buf))
		require.Equal(t, "bar\n", buf.String())
	})

	suite.Run("FileHistory", func(t *testing.T) {
		// TODO: There is no notion of file history in V2. We could potentially implement this, but
		// we would need to spend some time thinking about the performance characteristics.
		t.Skip("File history is not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	var err error
		//
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//	numCommits := 10
		//	for i := 0; i < numCommits; i++ {
		//		_, err = env.PachClient.PutFile(repo, "master", "file", strings.NewReader("foo\n"))
		//		require.NoError(t, err)
		//	}
		//	fileInfos, err := env.PachClient.ListFileHistory(repo, "master", "file", -1)
		//	require.NoError(t, err)
		//	require.Equal(t, numCommits, len(fileInfos))
		//
		//	for i := 1; i < numCommits; i++ {
		//		fileInfos, err := env.PachClient.ListFileHistory(repo, "master", "file", int64(i))
		//		require.NoError(t, err)
		//		require.Equal(t, i, len(fileInfos))
		//	}
		//
		//	require.NoError(t, env.PachClient.DeleteFile(repo, "master", "file"))
		//	for i := 0; i < numCommits; i++ {
		//		_, err = env.PachClient.PutFile(repo, "master", "file", strings.NewReader("foo\n"))
		//		require.NoError(t, err)
		//		_, err = env.PachClient.PutFile(repo, "master", "unrelated", strings.NewReader("foo\n"))
		//		require.NoError(t, err)
		//	}
		//	fileInfos, err = env.PachClient.ListFileHistory(repo, "master", "file", -1)
		//	require.NoError(t, err)
		//	require.Equal(t, numCommits, len(fileInfos))
		//
		//	for i := 1; i < numCommits; i++ {
		//		fileInfos, err := env.PachClient.ListFileHistory(repo, "master", "file", int64(i))
		//		require.NoError(t, err)
		//		require.Equal(t, i, len(fileInfos))
		//	}
	})

	suite.Run("UpdateRepo", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		var err error
		repo := "test"
		_, err = env.PachClient.PfsAPIClient.CreateRepo(
			env.PachClient.Ctx(),
			&pfs.CreateRepoRequest{
				Repo:   client.NewRepo(pfs.DefaultProjectName, repo),
				Update: true,
			},
		)
		require.NoError(t, err)
		ri, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		created := ri.Created.AsTime()
		desc := "foo"
		_, err = env.PachClient.PfsAPIClient.CreateRepo(
			env.PachClient.Ctx(),
			&pfs.CreateRepoRequest{
				Repo:        client.NewRepo(pfs.DefaultProjectName, repo),
				Update:      true,
				Description: desc,
			},
		)
		require.NoError(t, err)
		ri, err = env.PachClient.InspectRepo(pfs.DefaultProjectName, repo)
		require.NoError(t, err)
		newCreated := ri.Created.AsTime()
		require.Equal(t, created, newCreated)
		require.Equal(t, desc, ri.Description)
	})

	suite.Run("DeferredProcessing", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "input"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "output1"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "output2"))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output1", "staging", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "input", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "output1", "staging", ""))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output2", "staging", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "output1", "master")}))
		require.NoError(t, finishCommit(env.PachClient, "output2", "staging", ""))
		require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, "input", "staging", ""), "file", strings.NewReader("foo")))
		commitInfoA, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "input", "staging", "")
		require.NoError(t, err)
		commitsetID := commitInfoA.Commit.Id

		commitInfos, err := env.PachClient.WaitCommitSetAll(commitsetID)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, commitInfoA.Commit, commitInfos[0].Commit)

		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "input", "master", "staging", "", nil))
		require.NoError(t, finishCommit(env.PachClient, "output1", "staging", ""))
		commitInfoB, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "input", "master", "")
		require.NoError(t, err)
		require.Equal(t, commitsetID, commitInfoB.Commit.Id)
		commitInfos, err = env.PachClient.WaitCommitSetAll(commitsetID)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos)) // only input@master. output1@staging should now have a commit that's provenant on input@master
		outputStagingCi, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output1", "staging", "")
		require.NoError(t, err)
		require.Equal(t, commitInfoB.Commit.Repo, outputStagingCi.DirectProvenance[0].Repo)
		require.Equal(t, commitInfoB.Commit.Id, outputStagingCi.DirectProvenance[0].Id)
		// now kick off output2@staging by fast-forwarding output1@master -> output1@staging
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, "output1", "master", "staging", "", nil))
		require.NoError(t, finishCommit(env.PachClient, "output2", "staging", ""))
		output1Master, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output1", "master", "")
		require.NoError(t, err)
		commitInfos, err = env.PachClient.WaitCommitSetAll(output1Master.Commit.Id)
		require.NoError(t, err)
		require.Equal(t, 2, len(commitInfos))
		output2StagingCi, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, "output2", "staging", "")
		require.NoError(t, err)
		require.Equal(t, outputStagingCi.Commit.Repo, output2StagingCi.DirectProvenance[0].Repo)
		require.Equal(t, outputStagingCi.Commit.Id, output2StagingCi.DirectProvenance[0].Id)
	})

	suite.Run("SquashCommitEmptyChild", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		repo := "repo"
		file := "foo"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))

		commitA, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.PutFile(commitA, file, strings.NewReader("foo")))
		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))

		commitB, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		// squash fails, child is still open
		err = env.PachClient.SquashCommitSet(commitA.Id)
		require.YesError(t, err)
		require.Matches(t, "cannot squash until child commit .* is finished", err.Error())

		require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
		// wait until the commit is completely finished
		_, err = env.PachClient.WaitCommit(pfs.DefaultProjectName, repo, "master", "")
		require.NoError(t, err)

		// now squashing succeeds
		require.NoError(t, env.PachClient.SquashCommitSet(commitA.Id))

		var b bytes.Buffer
		require.NoError(t, env.PachClient.GetFile(commitB, file, &b))
		require.Equal(t, "foo", b.String())
	})

	suite.Run("ListAll", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo1"))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "repo2"))
		commit1 := client.NewCommit(pfs.DefaultProjectName, "repo1", "master", "")
		commit2 := client.NewCommit(pfs.DefaultProjectName, "repo2", "master", "")
		require.NoError(t, env.PachClient.PutFile(commit1, "file1", strings.NewReader("1")))
		require.NoError(t, env.PachClient.PutFile(commit2, "file2", strings.NewReader("2")))
		require.NoError(t, env.PachClient.PutFile(commit1, "file3", strings.NewReader("3")))
		require.NoError(t, env.PachClient.PutFile(commit2, "file4", strings.NewReader("4")))

		cis, err := env.PachClient.ListCommitByRepo(client.NewRepo(pfs.DefaultProjectName, ""))
		require.NoError(t, err)
		require.Equal(t, 4, len(cis))

		bis, err := env.PachClient.ListBranch(pfs.DefaultProjectName, "")
		require.NoError(t, err)
		require.Equal(t, 2, len(bis))
	})

	suite.Run("MonkeyObjectStorage", func(t *testing.T) {
		// This test cannot be done in parallel because the monkey object client
		// modifies global state.
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		seedStr := func(seed int64) string {
			return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
		}
		monkeyRetry := func(t *testing.T, f func() error, errMsg string) {
			backoff.Retry(func() error { //nolint:errcheck
				err := f()
				if err != nil {
					require.True(t, obj.IsMonkeyError(err), "Expected monkey error (%s), %s", err.Error(), errMsg)
				}
				return err
			}, backoff.NewInfiniteBackOff())
		}
		seed := time.Now().UTC().UnixNano()
		obj.InitMonkeyTest(seed)
		iterations := 25
		repo := "input"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo), seedStr(seed))
		filePrefix := "file"
		dataPrefix := "data"
		var commit *pfs.Commit
		var err error
		buf := &bytes.Buffer{}
		obj.EnableMonkeyTest()
		defer obj.DisableMonkeyTest()
		for i := 0; i < iterations; i++ {
			file := filePrefix + strconv.Itoa(i)
			data := dataPrefix + strconv.Itoa(i)
			// Retry start commit until it eventually succeeds.
			monkeyRetry(t, func() error {
				commit, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
				return err
			}, seedStr(seed))
			// Retry put file until it eventually succeeds.
			monkeyRetry(t, func() error {
				if err := env.PachClient.PutFile(commit, file, strings.NewReader(data)); err != nil {
					// Verify that the file does not exist if an error occurred.
					obj.DisableMonkeyTest()
					defer obj.EnableMonkeyTest()
					buf.Reset()
					err := env.PachClient.GetFile(commit, file, buf)
					require.True(t, errutil.IsNotFoundError(err), seedStr(seed))
				}
				return err
			}, seedStr(seed))
			// Retry get file until it eventually succeeds (before commit is finished).
			monkeyRetry(t, func() error {
				buf.Reset()
				if err = env.PachClient.GetFile(commit, file, buf); err != nil {
					return err
				}
				require.Equal(t, data, buf.String(), seedStr(seed))
				return nil
			}, seedStr(seed))
			// Retry finish commit until it eventually succeeds.
			monkeyRetry(t, func() error {
				return finishCommit(env.PachClient, repo, commit.Branch.Name, commit.Id)
			}, seedStr(seed))
			// Retry get file until it eventually succeeds (after commit is finished).
			monkeyRetry(t, func() error {
				buf.Reset()
				if err = env.PachClient.GetFile(commit, file, buf); err != nil {
					return err
				}
				require.Equal(t, data, buf.String(), seedStr(seed))
				return nil
			}, seedStr(seed))
		}
	})

	suite.Run("FsckFix", func(t *testing.T) {
		// TODO(optional 2.0): force-deleting the repo no longer creates dangling references
		t.Skip("this test no longer creates invalid metadata")
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		input := "input"
		output1 := "output1"
		output2 := "output2"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, input))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, output1))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, output2))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, output1, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, input, "master")}))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, output2, "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, output1, "master")}))
		numCommits := 10
		for i := 0; i < numCommits; i++ {
			require.NoError(t, env.PachClient.PutFile(client.NewCommit(pfs.DefaultProjectName, input, "master", ""), "file", strings.NewReader("1")))
		}
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, input, true))
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, input))
		require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, input, "master", "", "", nil))

		// Fsck should fail because ???
		require.YesError(t, env.PachClient.FsckFastExit())

		// Deleting output1 should fail because output2 is provenant on it
		require.YesError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, output1, false))

		// Deleting should now work due to fixing, must delete 2 before 1 though.
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, output2, false))
		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, output1, false))
	})

	suite.Run("PutFileAtomic", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		c := env.PachClient
		test := "test"
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, test))
		testRepo := client.NewRepo(pfs.DefaultProjectName, test)
		commit := testRepo.NewCommit("master", "")

		mfc, err := c.NewModifyFileClient(commit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("file1", strings.NewReader("1")))
		require.NoError(t, mfc.PutFile("file2", strings.NewReader("2")))
		require.NoError(t, mfc.Close())

		cis, err := c.ListCommit(testRepo, commit, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(cis))
		var b bytes.Buffer
		require.NoError(t, c.GetFile(commit, "file1", &b))
		require.Equal(t, "1", b.String())
		b.Reset()
		require.NoError(t, c.GetFile(commit, "file2", &b))
		require.Equal(t, "2", b.String())

		mfc, err = c.NewModifyFileClient(commit)
		require.NoError(t, err)
		require.NoError(t, mfc.PutFile("file3", strings.NewReader("3")))
		require.NoError(t, err)
		require.NoError(t, mfc.DeleteFile("file1"))
		require.NoError(t, mfc.Close())

		cis, err = c.ListCommit(testRepo, commit, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 2, len(cis))
		b.Reset()
		require.NoError(t, c.GetFile(commit, "file3", &b))
		require.Equal(t, "3", b.String())
		b.Reset()
		require.YesError(t, c.GetFile(commit, "file1", &b))

		mfc, err = c.NewModifyFileClient(commit)
		require.NoError(t, err)
		require.NoError(t, mfc.Close())
		cis, err = c.ListCommit(testRepo, commit, nil, 0)
		require.NoError(t, err)
		require.Equal(t, 3, len(cis))
	})

	suite.Run("TestTopologicalSortCommits", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			require.True(t, len(server.TopologicalSort([]*pfs.CommitInfo{})) == 0)
		})
		t.Run("Fuzz", func(t *testing.T) {
			// TODO: update gopls for generics
			union := func(a map[string]struct{}, b map[string]struct{}) map[string]struct{} {
				c := make(map[string]struct{})
				for el := range a {
					c[el] = struct{}{}
				}
				for el := range b {
					c[el] = struct{}{}
				}
				return c
			}
			proj := "foo"
			makeCommit := func(i int) *pfs.Commit {
				return client.NewCommit(proj, fmt.Sprintf("%d", i), "", fmt.Sprintf("%d", i))
			}
			// commit.String() -> number of total transitive provenant commits
			totalProvenance := make(map[string]map[string]struct{})
			var cis []*pfs.CommitInfo
			total := 500
			for i := 0; i < total; i++ {
				totalProv := make(map[string]struct{})
				directProv := make([]*pfs.Commit, 0)
				if i > 0 {
					provCount := rand.Intn(i)
					for j := 0; j < provCount; j++ {
						k := rand.Intn(i)
						totalProv[makeCommit(k).String()] = struct{}{}
						totalProv = union(totalProv, totalProvenance[makeCommit(k).String()])
						directProv = append(directProv, makeCommit(k))
					}
				}
				ci := &pfs.CommitInfo{
					Commit:           makeCommit(i),
					DirectProvenance: directProv,
				}
				totalProvenance[makeCommit(i).String()] = totalProv
				cis = append(cis, ci)
			}
			// shuffle cis
			swaps := total / 2
			for i := 0; i < swaps; i++ {
				j, k := rand.Intn(total), rand.Intn(total)
				cis[j], cis[k] = cis[k], cis[j]
			}
			// assert sort works
			cis = server.TopologicalSort(cis)
			for i, ci := range cis {
				require.True(t, len(totalProvenance[ci.Commit.String()]) <= i)
			}
		})
	})

	const (
		inputRepo          = iota // create a new input repo
		inputBranch               // create a new branch on an existing input repo
		deleteInputBranch         // delete an input branch
		commit                    // commit to an input branch
		squashCommitSet           // squash a commitset from an input branch
		outputRepo                // create a new output repo, with master branch subscribed to random other branches
		outputBranch              // create a new output branch on an existing output repo
		deleteOutputBranch        // delete an output branch
	)

	suite.Run("FuzzProvenance", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		seed := time.Now().UnixNano()
		t.Log("Random seed is", seed)
		r := rand.New(rand.NewSource(seed))

		_, err := env.PachClient.PfsAPIClient.DeleteAll(env.PachClient.Ctx(), &emptypb.Empty{})
		require.NoError(t, err)
		nOps := 300
		opShares := []int{
			1, // inputRepo
			1, // inputBranch
			1, // deleteInputBranch
			5, // commit
			3, // squashCommitSet
			1, // outputRepo
			2, // outputBranch
			1, // deleteOutputBranch
		}
		total := 0
		for _, v := range opShares {
			total += v
		}
		var (
			inputRepos     []string
			inputBranches  []*pfs.Branch
			commits        []*pfs.Commit
			outputBranches []*pfs.Branch
		)
	OpLoop:
		for i := 0; i < nOps; i++ {
			roll := r.Intn(total)
			if i < 0 {
				roll = inputRepo
			}
			var op int
			for _op, v := range opShares {
				roll -= v
				if roll < 0 {
					op = _op
					break
				}
			}
			switch op {
			case inputRepo:
				repo := tu.UniqueString("repo")
				require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
				inputRepos = append(inputRepos, repo)
				require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
				inputBranches = append(inputBranches, client.NewBranch(pfs.DefaultProjectName, repo, "master"))
			case inputBranch:
				if len(inputRepos) == 0 {
					continue OpLoop
				}
				repo := inputRepos[r.Intn(len(inputRepos))]
				branch := tu.UniqueString("branch")
				require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, branch, "", "", nil))
				inputBranches = append(inputBranches, client.NewBranch(pfs.DefaultProjectName, repo, branch))
			case deleteInputBranch:
				if len(inputBranches) == 0 {
					continue OpLoop
				}
				i := r.Intn(len(inputBranches))
				branch := inputBranches[i]
				err = env.PachClient.DeleteBranch(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, false)
				// don't fail if the error was just that it couldn't delete the branch without breaking subvenance
				inputBranches = append(inputBranches[:i], inputBranches[i+1:]...)
				if err != nil && !strings.Contains(err.Error(), "break") {
					require.NoError(t, err)
				}
			case commit:
				if len(inputBranches) == 0 {
					continue OpLoop
				}
				branch := inputBranches[r.Intn(len(inputBranches))]
				commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, branch.Repo.Name, branch.Name)
				require.NoError(t, err)
				require.NoError(t, finishCommit(env.PachClient, branch.Repo.Name, branch.Name, commit.Id))
				// find and finish all commits in output branches, too
				infos, err := env.PachClient.InspectCommitSet(commit.Id)
				require.NoError(t, err)
				for _, info := range infos {
					if info.Commit.Id != commit.Id {
						continue
					}
					require.NoError(t, finishCommit(env.PachClient,
						info.Commit.Branch.Repo.Name, info.Commit.Branch.Name, commit.Id))
				}
				commits = append(commits, commit)
			case squashCommitSet:
				if len(commits) == 0 {
					continue OpLoop
				}
				i := r.Intn(len(commits))
				commit := commits[i]

				err := env.PachClient.SquashCommitSet(commit.Id)
				if pfsserver.IsSquashWithoutChildrenErr(err) {
					err = env.PachClient.DropCommitSet(commit.Id)
					if pfsserver.IsDropWithChildrenErr(err) {
						// The commitset cannot be squashed or dropped as some commits have children and some commits don't
						continue
					}
				} else if pfsserver.IsSquashWithSuvenanceErr(err) {
					// TODO(acohen4): destructure error and successfully squash all the dependent commit sets
					continue
				}
				require.NoError(t, err)
				commits = append(commits[:i], commits[i+1:]...)
				ris, err := env.PachClient.ListRepo()
				require.NoError(t, err)
				for _, ri := range ris {
					bis, err := env.PachClient.ListBranch(pfs.DefaultProjectName, ri.Repo.Name)
					require.NoError(t, err)
					for _, bi := range bis {
						branch := bi.Branch
						info, err := env.PachClient.InspectCommit(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, "")
						require.NoError(t, err)
						if info.Finishing == nil {
							require.NoError(t, finishCommit(env.PachClient, branch.Repo.Name, branch.Name, ""))
						}
					}
				}
			case outputRepo:
				if len(inputBranches) == 0 {
					continue OpLoop
				}
				repo := tu.UniqueString("out")
				require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
				// outputRepos = append(outputRepos, repo)
				var provBranches []*pfs.Branch
				for num, i := range r.Perm(len(inputBranches))[:r.Intn(len(inputBranches))] {
					provBranches = append(provBranches, inputBranches[i])
					if num > 1 {
						break
					}
				}
				err = env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", provBranches)
				if err != nil {
					if pfsserver.IsInvalidBranchStructureErr(err) || strings.Contains(err.Error(), "cannot be in the provenance of its own branch") {
						continue
					}
					require.NoError(t, err)
				} else {
					outputBranches = append(outputBranches, client.NewBranch(pfs.DefaultProjectName, repo, "master"))
					if len(provBranches) > 0 {
						require.NoError(t, finishCommit(env.PachClient, repo, "master", ""))
					}
				}
			case deleteOutputBranch:
				if len(outputBranches) == 0 {
					continue OpLoop
				}
				i := r.Intn(len(outputBranches))
				branch := outputBranches[i]
				err = env.PachClient.DeleteBranch(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, false)
				// don't fail if the error was just that it couldn't delete the branch without breaking subvenance
				outputBranches = append(outputBranches[:i], outputBranches[i+1:]...)
				if err != nil && !strings.Contains(err.Error(), "break") {
					require.NoError(t, err)
				}
			}
			require.NoError(t, env.PachClient.FsckFastExit())
		}
		// make sure we can delete at the end
		_, err = env.PachClient.PfsAPIClient.DeleteAll(env.PachClient.Ctx(), &emptypb.Empty{})
		require.NoError(t, err)
	})

	// TestAtomicHistory repeatedly writes to a file while concurrently reading
	// its history. This checks for a regression where the repo would sometimes
	// lock.
	suite.Run("AtomicHistory", func(t *testing.T) {
		// TODO: There is no notion of file history in V2. We could potentially implement this, but
		// we would need to spend some time thinking about the performance characteristics.
		t.Skip("File history is not implemented in V2")
		//	t.Parallel()
		//  env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
		//
		//	repo := "test"
		//	require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName,repo))
		//	require.NoError(t, env.PachClient.CreateProjectBranch(pfs.DefaultProjectName,repo, "master", "", nil))
		//	aSize := 1 * 1024 * 1024
		//	bSize := aSize + 1024
		//
		//	for i := 0; i < 10; i++ {
		//		// create a file of all A's
		//		a := strings.Repeat("A", aSize)
		//		_, err := env.PachClient.PutFile(repo, "master", "/file", strings.NewReader(a))
		//		require.NoError(t, err)
		//
		//		// sllowwwllly replace it with all B's
		//		ctx, cancel := context.WithCancel(context.Background())
		//		eg, ctx := errgroup.WithContext(ctx)
		//		eg.Go(func() error {
		//			b := strings.Repeat("B", bSize)
		//			r := SlowReader{underlying: strings.NewReader(b)}
		//			_, err := env.PachClient.PutFile(repo, "master", "/file", &r)
		//			cancel()
		//			return err
		//		})
		//
		//		// should pull /file when it's all A's
		//		eg.Go(func() error {
		//			for {
		//				fileInfos, err := env.PachClient.ListFileHistory(repo, "master", "/file", 1)
		//				require.NoError(t, err)
		//				require.Equal(t, len(fileInfos), 1)
		//
		//				// stop once B's have been written
		//				select {
		//				case <-ctx.Done():
		//					return nil
		//				default:
		//					time.Sleep(1 * time.Millisecond)
		//				}
		//			}
		//		})
		//
		//		require.NoError(t, eg.Wait())
		//
		//		// should pull /file when it's all B's
		//		fileInfos, err := env.PachClient.ListFileHistory(repo, "master", "/file", 1)
		//		require.NoError(t, err)
		//		require.Equal(t, 1, len(fileInfos))
		//		require.Equal(t, bSize, int(fileInfos[0].SizeBytes))
		//	}
	})

	// TestTrigger tests branch triggers
	// TODO: This test can be refactored to remove a lot of the boilerplate.
	suite.Run("Trigger", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient

		t.Run("Simple", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "test"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "test", "master", "", "", &pfs.Trigger{
				Branch: "staging",
				Size:   "1B",
			}))
			require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "test", "staging", ""), "file", strings.NewReader("small")))
		})

		t.Run("SizeWithProvenance", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "in"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "in", "trigger", "", "", &pfs.Trigger{
				Branch: "master",
				Size:   "1K",
			}))
			inCommit := client.NewCommit(pfs.DefaultProjectName, "in", "master", "")
			bis, err := c.ListBranch(pfs.DefaultProjectName, "in")
			require.NoError(t, err)
			require.Equal(t, 1, len(bis))
			// Create a downstream branch
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "out"))
			require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "out", "master", "", "", []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "in", "trigger")}))
			require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, "out", "master", ""))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "out", "trigger", "", "", &pfs.Trigger{
				Branch: "master",
				Size:   "1K",
			}))
			// Write a small file, too small to trigger
			require.NoError(t, c.PutFile(inCommit, "file", strings.NewReader("small")))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "in", "master", "")
			require.NoError(t, err)
			bi, err := c.InspectBranch(pfs.DefaultProjectName, "in", "master")
			require.NoError(t, err)
			head := bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "in", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "out", "master")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "out", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)

			require.NoError(t, c.PutFile(inCommit, "file", strings.NewReader(strings.Repeat("a", units.KB))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "in", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "in", "master")
			require.NoError(t, err)
			head = bi.Head.Id

			bi, err = c.InspectBranch(pfs.DefaultProjectName, "in", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			// Output branch should have a commit now
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "out", "master")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)

			resolvedAlias, err := c.InspectCommit(pfs.DefaultProjectName, "in", "", bi.Head.Id)
			require.NoError(t, err)
			require.Equal(t, head, resolvedAlias.Commit.Id)

			// Put a file that will cause the trigger to go off
			require.NoError(t, c.PutFile(client.NewCommit(pfs.DefaultProjectName, "out", "master", ""), "file", strings.NewReader(strings.Repeat("a", units.KB))))
			require.NoError(t, c.FinishCommit(pfs.DefaultProjectName, "out", "master", ""))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "out", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "out", "master")
			require.NoError(t, err)
			head = bi.Head.Id

			// Output trigger should have triggered
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "out", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
		})

		t.Run("Cron", func(t *testing.T) {
			t.Parallel()
			repo := tu.UniqueString("Cron")
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, repo, "trigger", "", "", &pfs.Trigger{
				Branch:   "master",
				CronSpec: "* * * * *", // every minute
			}))
			// Create initial commit.
			commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
			require.NoError(t, c.PutFile(commit, "file1", strings.NewReader("foo")))
			bi, err := c.InspectBranch(pfs.DefaultProjectName, repo, "master")
			require.NoError(t, err)
			head := bi.Head.Id
			// Ensure that the trigger fired after a minute.
			time.Sleep(time.Minute)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			// Ensure that the trigger branch remains unchanged after another minute (no new commmit).
			time.Sleep(time.Minute)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			// Ensure that the trigger still works after another commmit.
			require.NoError(t, c.PutFile(commit, "file1", strings.NewReader("foo")))
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "master")
			require.NoError(t, err)
			head = bi.Head.Id
			time.Sleep(time.Minute)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
		})

		t.Run("CronUpdate", func(t *testing.T) {
			t.Parallel()
			repo := tu.UniqueString("CronUpdate")
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
			// Create the initial trigger for every minute, then update it to every January.
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, repo, "trigger", "", "", &pfs.Trigger{
				Branch:   "master",
				CronSpec: "* * * * *", // every minute
			}))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, repo, "trigger", "", "", &pfs.Trigger{
				Branch:   "master",
				CronSpec: "* * * 1 *", // every January
			}))
			// Create initial commit and ensure that it doesn't fire in a minute.
			commit := client.NewCommit(pfs.DefaultProjectName, repo, "master", "")
			require.NoError(t, c.PutFile(commit, "file1", strings.NewReader("foo")))
			bi, err := c.InspectBranch(pfs.DefaultProjectName, repo, "master")
			require.NoError(t, err)
			head := bi.Head.Id
			time.Sleep(time.Minute)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			// Update the trigger back to one minute and ensure that the trigger fires.
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, repo, "trigger", "", "", &pfs.Trigger{
				Branch:   "master",
				CronSpec: "* * * * *", // every minute
			}))
			time.Sleep(3 * time.Second)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, repo, "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
		})

		t.Run("Count", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "count"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "count", "trigger", "", "", &pfs.Trigger{
				Branch:  "master",
				Commits: 2, // trigger every 2 commits
			}))

			bi, err := c.InspectBranch(pfs.DefaultProjectName, "count", "trigger")
			require.NoError(t, err)
			head := bi.Head

			masterHead := client.NewCommit(pfs.DefaultProjectName, "count", "master", "")
			// The first commit shouldn't trigger
			require.NoError(t, c.PutFile(masterHead, "file1", strings.NewReader("foo")))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "count", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head)

			// Second commit should trigger
			require.NoError(t, c.PutFile(masterHead, "file2", strings.NewReader("bar")))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "count", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)
			head = bi.Head

			// The trigger commit should have the same ID as the master commit
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "master")
			require.NoError(t, err)
			require.Equal(t, head.Id, bi.Head.Id)

			// Third commit shouldn't trigger
			require.NoError(t, c.PutFile(masterHead, "file3", strings.NewReader("fizz")))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "count", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head)

			// Fourth commit should trigger
			require.NoError(t, c.PutFile(masterHead, "file4", strings.NewReader("buzz")))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "count", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)
			head = bi.Head

			// The trigger commit should have the same ID as the master commit
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "count", "master")
			require.NoError(t, err)
			require.Equal(t, head.Id, bi.Head.Id)
		})

		t.Run("Or", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "or"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "or", "trigger", "", "", &pfs.Trigger{
				Branch:        "master",
				RateLimitSpec: "* * * * *",
				Size:          "100",
				Commits:       3,
			}))
			orCommit := client.NewCommit(pfs.DefaultProjectName, "or", "master", "")
			// This triggers, because the cron is satisfied
			require.NoError(t, c.PutFile(orCommit, "file1", strings.NewReader(strings.Repeat("a", 1))))
			_, err := c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err := c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.NotNil(t, bi.Head)
			head := bi.Head.Id
			// This one doesn't because none of them are satisfied
			require.NoError(t, c.PutFile(orCommit, "file2", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			// This one triggers because we hit 100 bytes
			require.NoError(t, c.PutFile(orCommit, "file3", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			head = bi.Head.Id

			// This one doesn't trigger
			require.NoError(t, c.PutFile(orCommit, "file4", strings.NewReader(strings.Repeat("a", 1))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			// This one neither
			require.NoError(t, c.PutFile(orCommit, "file5", strings.NewReader(strings.Repeat("a", 1))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			// This one does, because it's 3 commits
			require.NoError(t, c.PutFile(orCommit, "file6", strings.NewReader(strings.Repeat("a", 1))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			head = bi.Head.Id

			// This one doesn't trigger
			require.NoError(t, c.PutFile(orCommit, "file7", strings.NewReader(strings.Repeat("a", 1))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			time.Sleep(time.Minute)

			require.NoError(t, c.PutFile(orCommit, "file8", strings.NewReader(strings.Repeat("a", 1))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "or", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "or", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
		})

		t.Run("And", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "and"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "and", "trigger", "", "", &pfs.Trigger{
				Branch:        "master",
				All:           true,
				RateLimitSpec: "* * * * *",
				Size:          "100",
				Commits:       3,
			}))
			andCommit := client.NewCommit(pfs.DefaultProjectName, "and", "master", "")
			// Doesn't trigger because all 3 conditions must be met
			require.NoError(t, c.PutFile(andCommit, "file1", strings.NewReader(strings.Repeat("a", 100))))
			_, err := c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err := c.InspectBranch(pfs.DefaultProjectName, "and", "master")
			require.NoError(t, err)
			head := bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)

			// Still doesn't trigger
			require.NoError(t, c.PutFile(andCommit, "file2", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)

			// Finally triggers because we have 3 commits, 100 bytes and Cron
			// Spec (since epoch) is satisfied.
			require.NoError(t, c.PutFile(andCommit, "file3", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.NotNil(t, bi.Head)
			head = bi.Head.Id

			// Doesn't trigger because all 3 conditions must be met
			require.NoError(t, c.PutFile(andCommit, "file4", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			// Still no trigger, not enough time or commits
			require.NoError(t, c.PutFile(andCommit, "file5", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			// Still no trigger, not enough time
			require.NoError(t, c.PutFile(andCommit, "file6", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			time.Sleep(time.Minute)

			// Finally triggers, all triggers have been met
			require.NoError(t, c.PutFile(andCommit, "file7", strings.NewReader(strings.Repeat("a", 100))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "and", "master", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "and", "trigger")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
		})

		t.Run("Chain", func(t *testing.T) {
			t.Parallel()
			// a triggers b which triggers c
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "chain"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "chain", "b", "", "", &pfs.Trigger{
				Branch: "a",
				Size:   "100",
			}))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "chain", "c", "", "", &pfs.Trigger{
				Branch: "b",
				Size:   "200",
			}))
			aCommit := client.NewCommit(pfs.DefaultProjectName, "chain", "a", "")
			// Triggers nothing
			require.NoError(t, c.PutFile(aCommit, "file1", strings.NewReader(strings.Repeat("a", 50))))
			_, err := c.WaitCommit(pfs.DefaultProjectName, "chain", "a", "")
			require.NoError(t, err)
			bi, err := c.InspectBranch(pfs.DefaultProjectName, "chain", "a")
			require.NoError(t, err)
			head := bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "b")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "c")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head)

			// Triggers b, but not c
			require.NoError(t, c.PutFile(aCommit, "file2", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "chain", "a", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "a")
			require.NoError(t, err)
			head = bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "b")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "c")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)

			// Triggers nothing
			require.NoError(t, c.PutFile(aCommit, "file3", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "chain", "a", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "a")
			require.NoError(t, err)
			head = bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "b")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "c")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)

			// Triggers b and c
			require.NoError(t, c.PutFile(aCommit, "file4", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "chain", "a", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "a")
			require.NoError(t, err)
			head = bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "b")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "c")
			require.NoError(t, err)
			require.Equal(t, head, bi.Head.Id)

			// Triggers nothing
			require.NoError(t, c.PutFile(aCommit, "file5", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "chain", "a", "")
			require.NoError(t, err)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "a")
			require.NoError(t, err)
			head = bi.Head.Id
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "b")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "chain", "c")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)
		})

		t.Run("BranchMovement", func(t *testing.T) {
			t.Parallel()
			require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "branch-movement"))
			require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "branch-movement", "c", "", "", &pfs.Trigger{
				Branch: "b",
				Size:   "100",
			}))
			moveCommit := client.NewCommit(pfs.DefaultProjectName, "branch-movement", "a", "")

			require.NoError(t, c.PutFile(moveCommit, "file1", strings.NewReader(strings.Repeat("a", 50))))
			_, err := c.WaitCommit(pfs.DefaultProjectName, "branch-movement", "a", "")
			require.NoError(t, err)
			bi, err := c.InspectBranch(pfs.DefaultProjectName, "branch-movement", "a")
			require.NoError(t, err)
			head := bi.Head.Id
			require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "branch-movement", "b", "a", "", nil))
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "branch-movement", "c")
			require.NoError(t, err)
			require.NotEqual(t, head, bi.Head.Id)

			require.NoError(t, c.PutFile(moveCommit, "file2", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "branch-movement", "a", "")
			require.NoError(t, err)
			require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "branch-movement", "b", "a", "", nil))
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "branch-movement", "c")
			require.NoError(t, err)
			require.NotNil(t, bi.Head)
			cHead := bi.Head.Id

			require.NoError(t, c.PutFile(moveCommit, "file3", strings.NewReader(strings.Repeat("a", 50))))
			_, err = c.WaitCommit(pfs.DefaultProjectName, "branch-movement", "a", "")
			require.NoError(t, err)
			require.NoError(t, c.CreateBranch(pfs.DefaultProjectName, "branch-movement", "b", "a", "", nil))
			bi, err = c.InspectBranch(pfs.DefaultProjectName, "branch-movement", "c")
			require.NoError(t, err)
			require.NotNil(t, bi.Head)
			require.Equal(t, cHead, bi.Head.Id)
		})
	})

	// TriggerValidation tests branch trigger validation
	suite.Run("TriggerValidation", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		c := env.PachClient
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "repo"))
		// Must specify a branch
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "master", "", "", &pfs.Trigger{
			Branch: "",
			Size:   "1K",
		}))
		// Can't trigger a branch on itself
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "master", "", "", &pfs.Trigger{
			Branch: "master",
			Size:   "1K",
		}))
		// Size doesn't parse
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "trigger", "", "", &pfs.Trigger{
			Branch: "master",
			Size:   "this is not a size",
		}))
		// Can't have negative commit count
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "trigger", "", "", &pfs.Trigger{
			Branch:  "master",
			Commits: -1,
		}))

		// a -> b (valid, sets up the next test)
		require.NoError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "b", "", "", &pfs.Trigger{
			Branch: "a",
			Size:   "1K",
		}))
		// Can't have circular triggers
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "a", "", "", &pfs.Trigger{
			Branch: "b",
			Size:   "1K",
		}))
		// CronSpec doesn't parse
		require.YesError(t, c.CreateBranchTrigger(pfs.DefaultProjectName, "repo", "trigger", "", "", &pfs.Trigger{
			Branch:   "master",
			CronSpec: "this is not a cron spec",
		}))
		// Can't use a trigger and provenance together
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, "in"))
		_, err := c.PfsAPIClient.CreateBranch(c.Ctx(),
			&pfs.CreateBranchRequest{
				Branch: client.NewBranch(pfs.DefaultProjectName, "repo", "master"),
				Trigger: &pfs.Trigger{
					Branch: "master",
					Size:   "1K",
				},
				Provenance: []*pfs.Branch{client.NewBranch(pfs.DefaultProjectName, "in", "master")},
			})
		require.YesError(t, err)
	})

	suite.Run("RegressionOrphanedFile", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		fsclient, err := env.PachClient.NewCreateFileSetClient()
		require.NoError(t, err)
		data := []byte("test data")
		spec := fileSetSpec{
			"file1.txt": tarutil.NewMemFile("file1.txt", data),
			"file2.txt": tarutil.NewMemFile("file2.txt", data),
		}
		require.NoError(t, fsclient.PutFileTAR(spec.makeTarStream()))
		resp, err := fsclient.Close()
		require.NoError(t, err)
		t.Logf("tmp fileset id: %s", resp.FileSetId)
		require.NoError(t, env.PachClient.RenewFileSet(resp.FileSetId, 60*time.Second))
		fis, err := env.PachClient.ListFileAll(client.NewCommit(pfs.DefaultProjectName, client.FileSetsRepoName, "", resp.FileSetId), "/")
		require.NoError(t, err)
		require.Equal(t, 2, len(fis))
	})

	suite.Run("Compaction", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, func(config *pachconfig.Configuration) {
			config.StorageCompactionMaxFanIn = 10
		}, dockertestenv.NewTestDBConfig(t))

		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		commit1, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, "master")
		require.NoError(t, err)

		const (
			nFileSets   = 100
			filesPer    = 10
			fileSetSize = 1e3
		)
		for i := 0; i < nFileSets; i++ {
			fsSpec := fileSetSpec{}
			for j := 0; j < filesPer; j++ {
				name := fmt.Sprintf("file%02d", j)
				data, err := io.ReadAll(randomReader(fileSetSize))
				require.NoError(t, err)
				file := tarutil.NewMemFile(name, data)
				hdr, err := file.Header()
				require.NoError(t, err)
				fsSpec[hdr.Name] = file
			}
			require.NoError(t, env.PachClient.PutFileTAR(commit1, fsSpec.makeTarStream()))
			runtime.GC()
		}
		require.NoError(t, finishCommit(env.PachClient, repo, commit1.Branch.Name, commit1.Id))
	})

	suite.Run("ModifyFileGRPC", func(subsuite *testing.T) {
		subsuite.Parallel()

		subsuite.Run("EmptyFile", func(t *testing.T) {
			t.Parallel()
			ctx := pctx.TestContext(t)
			env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
			repo := "test"
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
			c, err := env.PachClient.PfsAPIClient.ModifyFile(context.Background())
			require.NoError(t, err)
			files := []string{"/empty-1", "/empty-2"}
			require.NoError(t, c.Send(&pfs.ModifyFileRequest{
				Body: &pfs.ModifyFileRequest_SetCommit{SetCommit: client.NewCommit(pfs.DefaultProjectName, repo, "master", "")},
			}))
			for _, file := range files {
				require.NoError(t, c.Send(&pfs.ModifyFileRequest{
					Body: &pfs.ModifyFileRequest_AddFile{
						AddFile: &pfs.AddFile{
							Path: file,
							Source: &pfs.AddFile_Raw{
								Raw: &wrapperspb.BytesValue{},
							},
						},
					},
				}))
			}
			_, err = c.CloseAndRecv()
			require.NoError(t, err)
			require.NoError(t, env.PachClient.ListFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), "/", func(fi *pfs.FileInfo) error {
				require.True(t, files[0] == fi.File.Path)
				files = files[1:]
				return nil
			}))
			require.Equal(t, 0, len(files))
		})

		subsuite.Run("SingleMessageFile", func(t *testing.T) {
			t.Parallel()
			ctx := pctx.TestContext(t)
			env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
			repo := "test"
			require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
			filePath := "file"
			fileContent := "foo"
			c, err := env.PachClient.PfsAPIClient.ModifyFile(context.Background())
			require.NoError(t, err)
			require.NoError(t, c.Send(&pfs.ModifyFileRequest{
				Body: &pfs.ModifyFileRequest_SetCommit{SetCommit: client.NewCommit(pfs.DefaultProjectName, repo, "master", "")},
			}))
			require.NoError(t, c.Send(&pfs.ModifyFileRequest{
				Body: &pfs.ModifyFileRequest_AddFile{
					AddFile: &pfs.AddFile{
						Path: filePath,
						Source: &pfs.AddFile_Raw{
							Raw: wrapperspb.Bytes([]byte(fileContent)),
						},
					},
				},
			}))
			_, err = c.CloseAndRecv()
			require.NoError(t, err)
			buf := &bytes.Buffer{}
			require.NoError(t, env.PachClient.GetFile(client.NewCommit(pfs.DefaultProjectName, repo, "master", ""), filePath, buf))
			require.Equal(t, fileContent, buf.String())
		})
	})

	suite.Run("TestPanicOnNilArgs", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		c := env.PachClient
		requireNoPanic := func(err error) {
			t.Helper()
			if err != nil {
				// if a "transport is closing" error happened, pachd abruptly
				// closed the connection. Most likely this is caused by a panic.
				require.False(t, strings.Contains(err.Error(), "transport is closing"), err.Error())
			}
		}
		ctx, cf := pctx.WithCancel(c.Ctx())
		defer cf()
		_, err := c.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.InspectRepo(ctx, &pfs.InspectRepoRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.DeleteRepo(ctx, &pfs.DeleteRepoRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.ListCommit(ctx, &pfs.ListCommitRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.SquashCommitSet(c.Ctx(), &pfs.SquashCommitSetRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.InspectCommitSet(c.Ctx(), &pfs.InspectCommitSetRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.SubscribeCommit(ctx, &pfs.SubscribeCommitRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.InspectBranch(ctx, &pfs.InspectBranchRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.ListBranch(ctx, &pfs.ListBranchRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.DeleteBranch(ctx, &pfs.DeleteBranchRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.GetFileTAR(ctx, &pfs.GetFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.InspectFile(ctx, &pfs.InspectFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.ListFile(ctx, &pfs.ListFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.WalkFile(ctx, &pfs.WalkFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.GlobFile(ctx, &pfs.GlobFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.DiffFile(ctx, &pfs.DiffFileRequest{})
		requireNoPanic(err)
		_, err = c.PfsAPIClient.Fsck(ctx, &pfs.FsckRequest{})
		requireNoPanic(err)
	})

	suite.Run("ErroredCommits", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		repo := "test"
		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, repo))
		checks := func(t *testing.T, branch string) {
			// First commit should contain the first file.
			branchCommit := client.NewCommit(pfs.DefaultProjectName, repo, branch, "^2")
			expected := []string{"/f1"}
			require.NoError(t, env.PachClient.ListFile(branchCommit, "", func(fi *pfs.FileInfo) error {
				require.Equal(t, expected[0], fi.File.Path)
				expected = expected[1:]
				return nil
			}))
			require.Equal(t, 0, len(expected))
			// Second commit (errored commit) should still be readable with its content included.
			branchCommit = client.NewCommit(pfs.DefaultProjectName, repo, branch, "^1")
			expected = []string{"/f1", "/f2"}
			require.NoError(t, env.PachClient.ListFile(branchCommit, "", func(fi *pfs.FileInfo) error {
				require.Equal(t, expected[0], fi.File.Path)
				expected = expected[1:]
				return nil
			}))
			require.Equal(t, 0, len(expected))
			// Third commit should exclude the errored parent commit.
			branchCommit = client.NewCommit(pfs.DefaultProjectName, repo, branch, "")
			expected = []string{"/f1", "/f3"}
			require.NoError(t, env.PachClient.ListFile(branchCommit, "", func(fi *pfs.FileInfo) error {
				require.Equal(t, expected[0], fi.File.Path)
				expected = expected[1:]
				return nil
			}))
			require.Equal(t, 0, len(expected))
		}
		t.Run("FinishedErroredFinished", func(t *testing.T) {
			branch := uuid.New()
			require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, branch, "", "", nil))
			branchCommit := client.NewCommit(pfs.DefaultProjectName, repo, branch, "")
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f1", strings.NewReader("foo\n")))
			commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, branch)
			require.NoError(t, err)
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f2", strings.NewReader("foo\n")))
			_, err = env.PachClient.PfsAPIClient.FinishCommit(context.Background(), &pfs.FinishCommitRequest{
				Commit: commit,
				Error:  "error",
			})
			require.NoError(t, err)
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f3", strings.NewReader("foo\n")))
			checks(t, branch)
		})
		t.Run("FinishedErroredOpen", func(t *testing.T) {
			branch := uuid.New()
			require.NoError(t, env.PachClient.CreateBranch(pfs.DefaultProjectName, repo, branch, "", "", nil))
			branchCommit := client.NewCommit(pfs.DefaultProjectName, repo, branch, "")
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f1", strings.NewReader("foo\n")))
			commit, err := env.PachClient.StartCommit(pfs.DefaultProjectName, repo, branch)
			require.NoError(t, err)
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f2", strings.NewReader("foo\n")))
			_, err = env.PachClient.PfsAPIClient.FinishCommit(context.Background(), &pfs.FinishCommitRequest{
				Commit: commit,
				Error:  "error",
			})
			require.NoError(t, err)
			_, err = env.PachClient.StartCommit(pfs.DefaultProjectName, repo, branch)
			require.NoError(t, err)
			require.NoError(t, env.PachClient.PutFile(branchCommit, "f3", strings.NewReader("foo\n")))
			checks(t, branch)
		})

	})

	suite.Run("SystemRepoDependence", func(t *testing.T) {
		t.Parallel()
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))

		sysRepo := client.NewSystemRepo(pfs.DefaultProjectName, "test", pfs.MetaRepoType)

		// can't create system repo by itself
		_, err := env.PachClient.PfsAPIClient.CreateRepo(env.Context, &pfs.CreateRepoRequest{
			Repo: sysRepo,
		})
		require.YesError(t, err)

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "test"))
		// but now we can
		_, err = env.PachClient.PfsAPIClient.CreateRepo(env.Context, &pfs.CreateRepoRequest{
			Repo: sysRepo,
		})
		require.NoError(t, err)

		require.NoError(t, env.PachClient.DeleteRepo(pfs.DefaultProjectName, "test", false))

		// meta repo should be gone, too
		_, err = env.PachClient.PfsAPIClient.InspectRepo(env.Context, &pfs.InspectRepoRequest{
			Repo: sysRepo,
		})
		require.YesError(t, err)
		require.True(t, errutil.IsNotFoundError(err))
	})
	suite.Run("ErrorMessages", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
		// don't show user .user suffix
		_, err := env.PachClient.InspectRepo(pfs.DefaultProjectName, "test")
		require.YesError(t, err)
		require.True(t, errutil.IsNotFoundError(err))
		require.False(t, strings.Contains(err.Error(), pfs.UserRepoType))

		require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, "test"))

		err = env.PachClient.CreateRepo(pfs.DefaultProjectName, "test")
		require.YesError(t, err)
		require.True(t, errutil.IsAlreadyExistError(err))
		require.False(t, strings.Contains(err.Error(), pfs.UserRepoType))

		_, err = env.PachClient.InspectBranch(pfs.DefaultProjectName, "test", "branch")
		require.YesError(t, err)
		require.True(t, errutil.IsNotFoundError(err))
		require.False(t, strings.Contains(err.Error(), pfs.UserRepoType))

		_, err = env.PachClient.InspectCommit(pfs.DefaultProjectName, "test", "branch", uuid.NewWithoutDashes())
		require.YesError(t, err)
		require.True(t, errutil.IsNotFoundError(err))
		require.False(t, strings.Contains(err.Error(), pfs.UserRepoType))
	})

	suite.Run("EgressToPostgres", func(_suite *testing.T) {
		os.Setenv("PACHYDERM_SQL_PASSWORD", tu.DefaultPostgresPassword)

		type Schema struct {
			Id    int    `column:"ID" dtype:"INT"`
			A     string `column:"A" dtype:"VARCHAR(100)"`
			Datum string `column:"DATUM" dtype:"INT"`
		}

		type File struct {
			data string
			path string
		}

		tests := []struct {
			name           string
			files          []File
			options        *pfs.SQLDatabaseEgress
			tables         []string
			expectedCounts map[string]int64
		}{
			{
				name: "CSV",
				files: []File{
					{"1,Foo,101\n2,Bar,102", "/test_table/0000"},
					{"3,Hello,103\n4,World,104", "/test_table/subdir/0001"},
					{"1,this is in test_table2,201", "/test_table2/0000"},
					{"", "/empty_table/0000"},
				},
				options: &pfs.SQLDatabaseEgress{
					FileFormat: &pfs.SQLDatabaseEgress_FileFormat{
						Type: pfs.SQLDatabaseEgress_FileFormat_CSV,
					},
				},
				tables:         []string{"test_table", "test_table2", "empty_table"},
				expectedCounts: map[string]int64{"test_table": 4, "test_table2": 1, "empty_table": 0},
			},
			{
				name: "JSON",
				files: []File{
					{`{"ID":1,"A":"Foo","DATUM":101}
					  {"ID":2,"A":"Bar","DATUM":102}`, "/test_table/0000"},
					{`{"ID":3,"A":"Hello","DATUM":103}
					  {"ID":4,"A":"World","DATUM":104}`, "/test_table/subdir/0001"},
					{`{"ID":1,"A":"Foo","DATUM":201}`, "/test_table2/0000"},
					{"", "/empty_table/0000"},
				},
				options: &pfs.SQLDatabaseEgress{
					FileFormat: &pfs.SQLDatabaseEgress_FileFormat{
						Type:    pfs.SQLDatabaseEgress_FileFormat_JSON,
						Columns: []string{"ID", "A", "DATUM"},
					},
				},
				tables:         []string{"test_table", "test_table2", "empty_table"},
				expectedCounts: map[string]int64{"test_table": 4, "test_table2": 1, "empty_table": 0},
			},
			{
				name: "HEADER_CSV",
				files: []File{
					{"ID,A,DATUM\n1,Foo,101\n2,Bar,102", "/test_table/0000"},
					{"ID,A,DATUM\n3,Hello,103\n4,World,104", "/test_table/subdir/0001"},
					{"ID,A,DATUM\n1,this is in test_table2,201", "/test_table2/0000"},
					{"ID,A,DATUM", "/empty_table/0000"},
				},
				options: &pfs.SQLDatabaseEgress{
					FileFormat: &pfs.SQLDatabaseEgress_FileFormat{
						Type:    pfs.SQLDatabaseEgress_FileFormat_CSV,
						Columns: []string{"ID", "A", "DATUM"},
					},
				},
				tables:         []string{"test_table", "test_table2", "empty_table"},
				expectedCounts: map[string]int64{"test_table": 4, "test_table2": 1, "empty_table": 0},
			},
			{
				name: "HEADER_CSV_JUMBLED",
				files: []File{
					{"A,ID,DATUM\nFoo,1,101\nBar,2,102", "/test_table/0000"},
					{"A,DATUM,ID\nHello,103,3\nWorld,104,4", "/test_table/subdir/0001"},
					{"DATUM,A,ID\n201,this is in test_table2,1", "/test_table2/0000"},
					{"DATUM,ID,A", "/empty_table/0000"},
				},
				options: &pfs.SQLDatabaseEgress{
					FileFormat: &pfs.SQLDatabaseEgress_FileFormat{
						Type:    pfs.SQLDatabaseEgress_FileFormat_CSV,
						Columns: []string{"ID", "A", "DATUM"},
					},
				},
				tables:         []string{"test_table", "test_table2", "empty_table"},
				expectedCounts: map[string]int64{"test_table": 4, "test_table2": 1, "empty_table": 0},
			},
		}
		for _, test := range tests {
			_suite.Run(test.name, func(t *testing.T) {
				ctx := pctx.TestContext(t)
				env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
				// setup target database
				dbName := tu.GenerateEphemeralDBName(t)
				tu.CreateEphemeralDB(t, sqlx.NewDb(env.ServiceEnv.GetDBClient().DB, "postgres"), dbName)
				db := tu.OpenDB(t,
					dbutil.WithMaxOpenConns(1),
					dbutil.WithUserPassword(tu.DefaultPostgresUser, tu.DefaultPostgresPassword),
					dbutil.WithHostPort(dockertestenv.PGBouncerHost(), dockertestenv.PGBouncerPort),
					dbutil.WithDBName(dbName),
				)
				for _, tableName := range test.tables {
					require.NoError(t, pachsql.CreateTestTable(db, tableName, Schema{}))
				}

				// setup source repo based on target database, and generate fake data
				require.NoError(t, env.PachClient.CreateRepo(pfs.DefaultProjectName, dbName))
				commit := client.NewCommit(pfs.DefaultProjectName, dbName, "master", "")
				for _, f := range test.files {
					require.NoError(t, env.PachClient.PutFile(
						commit,
						f.path,
						strings.NewReader(f.data)))
				}

				// run Egress to copy data from source commit to target database
				test.options.Secret = &pfs.SQLDatabaseEgress_Secret{Name: "does not matter", Key: "does not matter"}
				test.options.Url = fmt.Sprintf("postgres://%s@%s:%d/%s", tu.DefaultPostgresUser, dockertestenv.PGBouncerHost(), dockertestenv.PGBouncerPort, dbName)
				resp, err := env.PachClient.Egress(env.PachClient.Ctx(),
					&pfs.EgressRequest{
						Commit: commit,
						Target: &pfs.EgressRequest_SqlDatabase{
							SqlDatabase: test.options,
						},
					})
				require.NoError(t, err)
				require.Equal(t, test.expectedCounts, resp.GetSqlDatabase().GetRowsWritten())

				// verify that actual rows got written to db
				var count int64
				for table, expected := range test.expectedCounts {
					require.NoError(t, db.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&count))
					require.Equal(t, expected, count)
				}
			})
		}
	})
}

var (
	randSeed = int64(0)
	randMu   sync.Mutex
)

type SlowReader struct {
	underlying io.Reader
	delay      time.Duration
}

func (r *SlowReader) Read(p []byte) (n int, err error) {
	n, err = r.underlying.Read(p)
	if r.delay == 0 {
		time.Sleep(1 * time.Millisecond)
	} else {
		time.Sleep(r.delay)
	}
	return
}

func getRand() *rand.Rand {
	randMu.Lock()
	seed := randSeed
	randSeed++
	randMu.Unlock()
	return rand.New(rand.NewSource(seed))
}

func randomReader(n int) io.Reader {
	return io.LimitReader(getRand(), int64(n))
}
