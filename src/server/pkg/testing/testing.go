package testing

import (
	"sort"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// State describes the state of a Pachyderm cluster. It's used to specify what
// a cluster should look like for the purposes of automating tests.
type State struct {
	Repos []*RepoState
}

type RepoState struct {
	Info    *pfs.RepoInfo
	Commits []*CommitState
}

type CommitState struct {
	Info  *pfs.CommitInfo
	Files []*FileState
}

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
	require.Equal(t, len(x), len(y))
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
