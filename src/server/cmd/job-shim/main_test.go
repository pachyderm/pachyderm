package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
)

func TestDownloadInput(t *testing.T) {
	c, err := client.NewFromAddress("127.0.0.1:30650")
	require.NoError(t, err)

	commitMounts := []*fuse.CommitMount{
		{
			Commit: client.NewCommit("input", "master/0"),
		},
		{
			Commit: client.NewCommit("output", "master/0"),
			Alias:  "out",
		},
	}

	require.NoError(t, downloadInput(c, commitMounts))
	require.NoError(t, ioutil.WriteFile(filepath.Join(PFSOutputPrefix, "file"), []byte("hello"), 0777))
	require.NoError(t, uploadOutput(c, commitMounts[1]))
}
