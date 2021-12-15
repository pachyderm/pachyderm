package pfs

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestErrorMatching(t *testing.T) {
	c := client.NewCommit("foo", "bar", "")
	require.True(t, IsCommitNotFoundErr(ErrCommitNotFound{Commit: c}))
	require.False(t, IsCommitNotFoundErr(ErrCommitDeleted{Commit: c}))
	require.False(t, IsCommitNotFoundErr(ErrCommitFinished{Commit: c}))

	require.False(t, IsCommitDeletedErr(ErrCommitNotFound{Commit: c}))
	require.True(t, IsCommitDeletedErr(ErrCommitDeleted{Commit: c}))
	require.False(t, IsCommitDeletedErr(ErrCommitFinished{Commit: c}))

	require.False(t, IsCommitFinishedErr(ErrCommitNotFound{Commit: c}))
	require.False(t, IsCommitFinishedErr(ErrCommitDeleted{Commit: c}))
	require.True(t, IsCommitFinishedErr(ErrCommitFinished{Commit: c}))
}
