package worker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestDatumIterators(t *testing.T) {
	c := getPachClient(t)
	defer require.NoError(t, c.DeleteAll())
	require.NoError(t, activateEnterprise(c))

	dataRepo := tu.UniqueString("TestDatumIteratorPFS_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// put files in structured in a way so that there are many ways to glob it
	commit, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	for j := 0; j < 50; j++ {
		_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("foo%v", j), strings.NewReader("bar"))
		require.NoError(t, err)
	}
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	in1 := client.NewPFSInput(dataRepo, "/foo?1")
	in1.Pfs.Commit = commit.ID
	pfs1, err := NewDatumIterator(c, in1)
	require.NoError(t, err)

	in2 := client.NewPFSInput(dataRepo, "/foo*2")
	in2.Pfs.Commit = commit.ID
	pfs2, err := NewDatumIterator(c, in2)
	require.NoError(t, err)

	// iterate through pfs1 and verify it is as we expect
	validateDI(t, "pfs1", pfs1)
	validateDI(t, "pfs2", pfs2)

	in3 := client.NewUnionInput(in1, in2)
	union1, err := NewDatumIterator(c, in3)
	require.NoError(t, err)
	validateDI(t, "union1", union1)

	in4 := client.NewCrossInput(in1, in2)
	cross1, err := NewDatumIterator(c, in4)
	require.NoError(t, err)
	validateDI(t, "cross1", cross1)

	in5 := client.NewCrossInput(in3, in4)
	cross2, err := NewDatumIterator(c, in5)
	require.NoError(t, err)
	validateDI(t, "cross2", cross2)

}

func validateDI(t *testing.T, name string, di DatumIterator) {
	datums := make(map[string]bool, di.Len())
	len := 0
	for di.Next() {
		key := ""
		for _, file := range di.Datum() {
			key += file.FileInfo.File.Path
		}

		if _, ok := datums[key]; ok {
			t.Errorf("Duplicate datum in %v detected: %v", name, key)
		}
		datums[key] = true
		len++
	}
	require.Equal(t, di.Len(), len)
}
