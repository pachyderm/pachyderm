package testing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/spf13/cobra"
)

var mu sync.Mutex

func RunCmd(cmd *cobra.Command, args []string, stdin []byte, t *testing.T) {
	// we lock a global mutex because this func modifies global state so 2
	// copies of it can't run concurrently
	mu.Lock()
	defer mu.Unlock()
	osArgs := os.Args
	defer func() { os.Args = osArgs }()
	os.Args = args
	if stdin != nil {
		osStdin := os.Stdin
		defer func() { os.Stdin = osStdin }()
		fauxStdin, err := ioutil.TempFile("", "RunCmd_stdin")
		require.NoError(t, err)
		_, err = fauxStdin.Write(stdin)
		require.NoError(t, err)
		_, err = fauxStdin.Seek(0, 0)
		require.NoError(t, err)
		os.Stdin = fauxStdin
	}
	require.NoError(t, cmd.Execute())
}

func TestCmd(cmd *cobra.Command, args []string, stdin []byte, expectedState *State, c *client.APIClient, t *testing.T) {
	RunCmd(cmd, args, stdin, t)
	MatchState(expectedState, c, t)
}

// State describes the state of a Pachyderm cluster. It's used to specify what
// a cluster should look like for the purposes of automating tests.
type State struct {
	Repos []*RepoState
}

func (s *State) Repo(name string) *RepoState {
	repoState := &RepoState{Info: &pfs.RepoInfo{Repo: client.NewRepo(name)}}
	s.Repos = append(s.Repos, repoState)
	return repoState
}

// RepoState describes state for a repo.
type RepoState struct {
	Info    *pfs.RepoInfo
	Commits []*CommitState
}

func (s *RepoState) Commit(id string) *CommitState {
	commitState := &CommitState{Info: &pfs.CommitInfo{Commit: client.NewCommit(s.Info.Repo.Name, id)}}
	s.Commits = append(s.Commits, commitState)
	return commitState
}

// CommitState describes state for a commit.
type CommitState struct {
	Info  *pfs.CommitInfo
	Files []*FileState
}

func (s *CommitState) File(path string) *FileState {
	fileState := &FileState{Info: &pfs.FileInfo{File: client.NewFile(s.Info.Commit.Repo.Name, s.Info.Commit.ID, path)}}
	s.Files = append(s.Files, fileState)
	return fileState
}

// FileState describes state for a file.
type FileState struct {
	Info    *pfs.FileInfo
	Content []byte
}

// MatchState attempts to match the state specified by `state` with the state
// it can access through `c`. It returns nil if the state matches otherwise it
// returns an error which describes what it didn't find. It also may return
// other client related errors, for example if there are connection problems.
// MatchState won't error if `c` contains state beyond what's specified in
// `state`.
func MatchState(state *State, c *client.APIClient, t *testing.T) {
	for _, repoState := range state.Repos {
		repoInfo, err := c.InspectRepo(repoState.Info.Repo.Name)
		require.NoError(t, err)
		matchRepoState(repoState, repoInfo, t)
		commitInfos, err := c.ListCommit([]*pfs.Commit{client.NewCommit(repoState.Info.Repo.Name, "")},
			nil, client.CommitTypeNone, client.CommitStatusAll, false)
		require.NoError(t, err)
		require.Equal(t, len(repoState.Commits), len(commitInfos))
		for i, commitState := range repoState.Commits {
			matchCommitState(commitState, commitInfos[i], t)
			for _, fileState := range commitState.Files {
				fileInfo, err := c.InspectFile(fileState.Info.File.Commit.Repo.Name,
					fileState.Info.File.Commit.ID, fileState.Info.File.Path, "", false, nil)
				require.NoError(t, err)
				var buffer bytes.Buffer
				require.NoError(t, c.GetFile(fileState.Info.File.Commit.Repo.Name,
					fileState.Info.File.Commit.ID, fileState.Info.File.Path, 0, 0, "", false, nil, &buffer))
				matchFileState(fileState, fileInfo, buffer.Bytes(), t)
			}
		}
	}
}

func matchRepoState(repoState *RepoState, repoInfo *pfs.RepoInfo, t *testing.T) {
	require.Equal(t, repoState.Info.Repo, repoInfo.Repo)
	if repoState.Info.Created != nil {
		require.Equal(t, repoState.Info.Created, repoInfo.Created)
	}
	if repoState.Info.SizeBytes != 0 {
		require.Equal(t, repoState.Info.SizeBytes, repoInfo.SizeBytes)
	}
	if repoState.Info.Provenance != nil {
		matchRepoProvenance(repoState.Info.Provenance, repoInfo.Provenance, t)
	}
}

func matchRepoProvenance(x []*pfs.Repo, y []*pfs.Repo, t *testing.T) {
	var xs []string
	var ys []string
	for _, repo := range x {
		xs = append(xs, repo.Name)
	}
	for _, repo := range y {
		ys = append(ys, repo.Name)
	}
	sort.Strings(xs)
	sort.Strings(ys)
	require.Equal(t, xs, ys)
}

func matchCommitState(commitState *CommitState, commitInfo *pfs.CommitInfo, t *testing.T) {
	require.Equal(t, commitState.Info.Commit, commitInfo.Commit)
	if commitState.Info.Branch != "" {
		require.Equal(t, commitState.Info.Branch, commitInfo.Commit)
	}
	if commitState.Info.CommitType != client.CommitTypeNone {
		require.Equal(t, commitState.Info.CommitType, commitInfo.CommitType)
	}
	if commitState.Info.ParentCommit != nil {
		require.Equal(t, commitState.Info.ParentCommit, commitInfo.ParentCommit)
	}
	if commitState.Info.Started != nil {
		require.Equal(t, commitState.Info.Started, commitInfo.Started)
	}
	if commitState.Info.Finished != nil {
		require.Equal(t, commitState.Info.Finished, commitInfo.Finished)
	}
	if commitState.Info.SizeBytes != 0 {
		require.Equal(t, commitState.Info.SizeBytes, commitInfo.SizeBytes)
	}
	if commitState.Info.Provenance != nil {
		matchCommitProvenance(commitState.Info.Provenance, commitInfo.Provenance, t)
	}
}

func matchCommitProvenance(x []*pfs.Commit, y []*pfs.Commit, t *testing.T) {
	var xs []string
	var ys []string
	for _, commit := range x {
		xs = append(xs, fmt.Sprintf("%s/%s", commit.Repo.Name, commit.ID))
	}
	for _, commit := range y {
		ys = append(ys, fmt.Sprintf("%s/%s", commit.Repo.Name, commit.ID))
	}
	sort.Strings(xs)
	sort.Strings(ys)
	require.Equal(t, xs, ys)
}

func matchFileState(fileState *FileState, fileInfo *pfs.FileInfo, content []byte, t *testing.T) {
	require.Equal(t, fileState.Info.File, fileInfo.File)
	if fileState.Info.FileType != pfs.FileType_FILE_TYPE_NONE {
		require.Equal(t, fileState.Info.FileType, fileInfo.FileType)
	}
	if fileState.Info.SizeBytes != 0 {
		require.Equal(t, fileState.Info.SizeBytes, fileInfo.SizeBytes)
	}
	if fileState.Info.Modified != nil {
		require.Equal(t, fileState.Info.Modified, fileInfo.Modified)
	}
	if fileState.Info.Children != nil {
		matchFiles(fileState.Info.Children, fileInfo.Children, t)
	}
}

func matchFiles(x []*pfs.File, y []*pfs.File, t *testing.T) {
	var xs []string
	var ys []string
	for _, file := range x {
		xs = append(xs, fmt.Sprintf("%s/%s/%s", file.Commit.Repo.Name, file.Commit.ID, file.Path))
	}
	for _, file := range y {
		ys = append(ys, fmt.Sprintf("%s/%s/%s", file.Commit.Repo.Name, file.Commit.ID, file.Path))
	}
	sort.Strings(xs)
	sort.Strings(ys)
	require.Equal(t, xs, ys)
}
