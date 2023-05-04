//go:build unit_test

package pfs

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestErrorMatching(t *testing.T) {
	c := client.NewCommit(pfs.DefaultProjectName, "foo", "bar", "")
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
