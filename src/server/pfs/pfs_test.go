package pfs

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestErrorMatching(t *testing.T) {
	c := client.NewCommit("foo", "bar")
	require.True(t, IsCommitNotFoundErr(ErrCommitNotFound{c}))
	require.False(t, IsCommitNotFoundErr(ErrCommitDeleted{c}))
	require.False(t, IsCommitNotFoundErr(ErrCommitFinished{c}))

	require.False(t, IsCommitDeletedErr(ErrCommitNotFound{c}))
	require.True(t, IsCommitDeletedErr(ErrCommitDeleted{c}))
	require.False(t, IsCommitDeletedErr(ErrCommitFinished{c}))

	require.False(t, IsCommitFinishedErr(ErrCommitNotFound{c}))
	require.False(t, IsCommitFinishedErr(ErrCommitDeleted{c}))
	require.True(t, IsCommitFinishedErr(ErrCommitFinished{c}))
}
