// +build linux

package fuse_test

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

	testFuse(t, func(c client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "", "")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commit.ID, "file")
		file, err := os.Create(path)
		OpenCommitSyscallSpec.NoError(t, err, "open")
		_, err = file.Write([]byte("foobarbaz"))

		OpenCommitSyscallSpec.NoError(t, err, "write")
		OpenCommitSyscallSpec.NoError(t, file.Close())
		require.NoError(t, c.FinishCommit(repo, commit.ID))

		fmt.Printf("==== Finished commit\n")

		file, err = os.Open(path)
		defer file.Close()
		ClosedCommitSyscallSpec.NoError(t, err, "open")

		word1 := make([]byte, 3)
		n1, err := file.Read(word1)
		ClosedCommitSyscallSpec.NoError(t, err, "read")
		require.Equal(t, 3, n1)
		require.Equal(t, "foo", string(word1))

		fmt.Printf("==== %v - Read word len %v : %v\n", time.Now(), n1, string(word1))

		offset, err := file.Seek(6, 0)
		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)

		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		ClosedCommitSyscallSpec.YesError(t, err)
		require.Equal(t, int64(0), offset)

		fmt.Printf("==== Seeked to %v\n", offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		word2 := make([]byte, 3)
		n2, err := file.Read(word2)
		require.NoError(t, err)
		require.Equal(t, 3, n2)
		require.Equal(t, "baz", string(word2))

		fmt.Printf("==== Read word len %v : %v\n", n2, string(word2)) */

	})
}

func TestSeekWriteGap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	testFuse(t, func(c client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "", "")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commit.ID, "file")
		file, err := os.Create(path)
		OpenCommitSyscallSpec.NoError(t, err, "open")
		defer func() {
			OpenCommitSyscallSpec.NoError(t, file.Close(), "close")
		}()

		_, err = file.Write([]byte("foo"))
		OpenCommitSyscallSpec.NoError(t, err, "write")

		err = file.Sync()
		OpenCommitSyscallSpec.NoError(t, err, "sync")

		offset, err := file.Seek(6, 0)

		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)
		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		OpenCommitSyscallSpec.YesError(t, err, "lseek")
		require.Equal(t, int64(0), offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		err = file.Sync()
		require.NoError(t, err)

		n1, err := file.Write([]byte("baz"))
		require.YesError(t, err)
		require.Equal(t, 3, n1)

		require.NoError(t, c.FinishCommit(repo, commit.ID)) */
	})
}

func TestSeekWriteBackwards(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	testFuse(t, func(c client.APIClient, mountpoint string) {
		repo := "test"
		require.NoError(t, c.CreateRepo(repo))
		commit, err := c.StartCommit(repo, "", "")
		require.NoError(t, err)
		path := filepath.Join(mountpoint, repo, commit.ID, "file")
		file, err := os.Create(path)
		OpenCommitSyscallSpec.NoError(t, err, "open")
		defer func() {
			OpenCommitSyscallSpec.NoError(t, file.Close(), "close")
		}()

		_, err = file.Write([]byte("foofoofoo"))
		OpenCommitSyscallSpec.NoError(t, err, "write")

		err = file.Sync()
		OpenCommitSyscallSpec.NoError(t, err, "sync")

		offset, err := file.Seek(3, 0)

		fmt.Printf("==== %v - err (%v)\n", time.Now(), err)
		fmt.Printf("==== %v - offset (%v)\n", time.Now(), offset)
		OpenCommitSyscallSpec.YesError(t, err, "lseek")
		require.Equal(t, int64(0), offset)

		/* Leaving in place so the test's intention is clear / for repro'ing manually for mac

		err = file.Sync()
		require.NoError(t, err)

		n1, err := file.Write([]byte("bar"))
		require.YesError(t, err)
		require.Equal(t, 3, n1)

		fmt.Printf("==== %v - write word len %v\n", time.Now(), n1)

		require.NoError(t, c.FinishCommit(repo, commit.ID)) */
	})
}
