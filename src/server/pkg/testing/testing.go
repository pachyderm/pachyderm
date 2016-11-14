package testing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/spf13/cobra"
)

var mu sync.Mutex

// RunCmd runs a cobra command setting os.Args and os.Stdin appropriately.
func RunCmd(cmd *cobra.Command, args []string, stdin string, t *testing.T) {
	// we lock a global mutex because this func modifies global state so 2
	// copies of it can't run concurrently
	mu.Lock()
	defer mu.Unlock()
	osArgs := os.Args
	defer func() { os.Args = osArgs }()
	os.Args = append([]string{"pachctl"}, args...)
	if stdin != "" {
		osStdin := os.Stdin
		defer func() { os.Stdin = osStdin }()
		fauxStdin, err := ioutil.TempFile("", "RunCmd_stdin")
		require.NoError(t, err)
		_, err = fauxStdin.Write([]byte(stdin))
		require.NoError(t, err)
		_, err = fauxStdin.Seek(0, 0)
		require.NoError(t, err)
		os.Stdin = fauxStdin
	}
	require.NoError(t, cmd.Execute())
}

// TestCmd runs a command using RunCmd and then checks that the cluster state matches expectedState.
func TestCmd(cmd *cobra.Command, args []string, stdin string, expectedState *State, c *client.APIClient, t *testing.T) {
	RunCmd(cmd, args, stdin, t)
	MatchState(expectedState, c, t)
}

// State describes the state of a Pachyderm cluster. It's used to specify what
// a cluster should look like for the purposes of automating tests.
type State struct {
	Repos     []*RepoState
	Pipelines []*PipelineState
}

// Repo adds an expected repo to a State object.
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

// Commit adds an expected Commit to a State object.
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

// File adds an expected file to a State object.
func (s *CommitState) File(path string) *FileState {
	fileState := &FileState{Info: &pfs.FileInfo{File: client.NewFile(s.Info.Commit.Repo.Name, s.Info.Commit.ID, path)}}
	s.Files = append(s.Files, fileState)
	return fileState
}

// FileState describes state for a file.
type FileState struct {
	Info    *pfs.FileInfo
	Content string
}

// Pipeline adds an expected Pipeline to a State object.
func (s *State) Pipeline(name string) (*PipelineState, *RepoState) {
	pipelineState := &PipelineState{Info: &pps.PipelineInfo{Pipeline: client.NewPipeline(name)}}
	s.Pipelines = append(s.Pipelines, pipelineState)
	return pipelineState, s.Repo(name)
}

// PipelineState describes the state for a Pipeline.
type PipelineState struct {
	Info *pps.PipelineInfo
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
		commitInfos, err := c.ListCommit([]string{repoState.Info.Repo.Name},
			nil, client.CommitTypeNone, client.CommitStatusAll, false)
		require.NoError(t, err)
		for i, commitState := range repoState.Commits {
			var commitInfo *pfs.CommitInfo
			if len(commitState.Info.Provenance) == 0 {
				require.True(t, i < len(commitInfos))
				commitInfo = commitInfos[i]
			} else {
				for _, rangeCommitInfo := range commitInfos {
					if equalCommitProvenance(commitState.Info.Provenance, rangeCommitInfo.Provenance) {
						commitInfo = rangeCommitInfo
						break
					}
				}
			}
			require.NotNil(t, commitInfo, "failed to find a commit with provenance %+v", commitState.Info.Provenance)
			matchCommitState(commitState, commitInfo, t)
			for _, fileState := range commitState.Files {
				fileInfo, err := c.InspectFile(commitInfo.Commit.Repo.Name,
					commitInfo.Commit.ID, fileState.Info.File.Path, "", false, nil)
				require.NoError(t, err)
				var buffer bytes.Buffer
				require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name,
					commitInfo.Commit.ID, fileState.Info.File.Path, 0, 0, "", false, nil, &buffer))
				matchFileState(fileState, fileInfo, buffer.String(), t)
			}
		}
	}
	for _, pipelineState := range state.Pipelines {
		pipelineInfo, err := c.InspectPipeline(pipelineState.Info.Pipeline.Name)
		require.NoError(t, err)
		matchPipelineState(pipelineState, pipelineInfo, t)
	}
}

func matchRepoState(repoState *RepoState, repoInfo *pfs.RepoInfo, t *testing.T) {
	matchRepo(repoState.Info.Repo, repoInfo.Repo, t)
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

func matchRepo(x *pfs.Repo, y *pfs.Repo, t *testing.T) {
	if x.Name != "" {
		require.Equal(t, x.Name, y.Name)
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
	matchCommit(commitState.Info.Commit, commitInfo.Commit, t)
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

func matchCommit(x *pfs.Commit, y *pfs.Commit, t *testing.T) {
	matchRepo(x.Repo, y.Repo, t)
	if x.ID != "" {
		require.Equal(t, x.ID, y.ID)
	}
}

func equalCommitProvenance(x []*pfs.Commit, y []*pfs.Commit) bool {
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
	return reflect.DeepEqual(xs, ys)
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

func matchFileState(fileState *FileState, fileInfo *pfs.FileInfo, content string, t *testing.T) {
	matchFile(fileState.Info.File, fileInfo.File, t)
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
	require.Equal(t, fileState.Content, content)
}

func matchFile(x *pfs.File, y *pfs.File, t *testing.T) {
	matchCommit(x.Commit, y.Commit, t)
	if x.Path != "" {
		require.Equal(t, x.Path, y.Path)
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

func matchPipelineState(pipelineState *PipelineState, pipelineInfo *pps.PipelineInfo, t *testing.T) {
	require.Equal(t, pipelineState.Info.Pipeline, pipelineInfo.Pipeline)
	if pipelineState.Info.Transform != nil {
		require.Equal(t, pipelineState.Info.Transform, pipelineInfo.Transform)
	}
	if pipelineState.Info.ParallelismSpec != nil {
		require.Equal(t, pipelineState.Info.ParallelismSpec, pipelineInfo.ParallelismSpec)
	}
	//TODO this isn't matching everything
}

// UniqueString string creates a unique string with a prefix.
func UniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}
