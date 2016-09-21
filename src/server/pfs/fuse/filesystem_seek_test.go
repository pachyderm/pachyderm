// +build linux

package fuse

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestSeekRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	testFuse(t, func(c *client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commitIDToPath(commit.ID), "file")
		file, err := os.Create(path)
		require.NoError(t, err)
		_, err = file.Write([]byte("foobarbaz"))

		require.NoError(t, err)
		require.NoError(t, file.Close())
		require.NoError(t, c.FinishCommit(repo, commit.ID))

		fmt.Printf("==== Finished commit\n")

		file, err = os.Open(path)
		defer file.Close()
		require.NoError(t, err)

		word1 := make([]byte, 3)
		n1, err := file.Read(word1)
		require.NoError(t, err)
		require.Equal(t, 3, n1)
		require.Equal(t, "foo", string(word1))

		fmt.Printf("==== %v - Read word len %v : %v\n", time.Now(), n1, string(word1))

		offset, err := file.Seek(6, 0)
		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)

		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		require.YesError(t, err)
		require.Equal(t, int64(0), offset)

		fmt.Printf("==== Seeked to %v\n", offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		word2 := make([]byte, 3)
		n2, err := file.Read(word2)
		require.NoError(t, err)
		require.Equal(t, 3, n2)
		require.Equal(t, "baz", string(word2))

		fmt.Printf("==== Read word len %v : %v\n", n2, string(word2)) */

	}, false)
}

func TestSeekWriteGap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	testFuse(t, func(c *client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commitIDToPath(commit.ID), "file")
		file, err := os.Create(path)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, file.Close())
		}()

		_, err = file.Write([]byte("foo"))
		require.NoError(t, err)

		err = file.Sync()
		require.NoError(t, err)

		offset, err := file.Seek(6, 0)

		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)
		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		require.YesError(t, err)
		require.Equal(t, int64(0), offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		err = file.Sync()
		require.NoError(t, err)

		n1, err := file.Write([]byte("baz"))
		require.YesError(t, err)
		require.Equal(t, 3, n1)

		require.NoError(t, c.FinishCommit(repo, commit.ID)) */
	}, false)
}

func TestSeekWriteBackwards(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	testFuse(t, func(c *client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "master")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commitIDToPath(commit.ID), "file")
		file, err := os.Create(path)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, file.Close())
		}()

		_, err = file.Write([]byte("foofoofoo"))
		require.NoError(t, err)

		err = file.Sync()
		require.NoError(t, err)

		offset, err := file.Seek(3, 0)

		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)
		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		require.YesError(t, err)
		require.Equal(t, int64(0), offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		err = file.Sync()
		require.NoError(t, err)

		n1, err := file.Write([]byte("bar"))
		require.YesError(t, err)
		require.Equal(t, 3, n1)

		fmt.Printf("==== %v - write word len %v\n", time.Now(), n1)

		require.NoError(t, c.FinishCommit(repo, commit.ID)) */
	}, false)
}
