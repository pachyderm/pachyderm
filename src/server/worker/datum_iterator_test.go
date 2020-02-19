package worker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestDatumIterators(t *testing.T) {
	c := tu.GetPachClient(t)
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

	// make one with zero datums for testing edge cases
	in0 := client.NewPFSInput(dataRepo, "!(**)")
	in0.Pfs.Commit = commit.ID
	pfs0, err := NewDatumIterator(c, in0)
	require.NoError(t, err)

	in1 := client.NewPFSInput(dataRepo, "/foo?1")
	in1.Pfs.Commit = commit.ID
	pfs1, err := NewDatumIterator(c, in1)
	require.NoError(t, err)

	in2 := client.NewPFSInput(dataRepo, "/foo*2")
	in2.Pfs.Commit = commit.ID
	pfs2, err := NewDatumIterator(c, in2)
	require.NoError(t, err)

	// iterate through pfs0, pfs1 and pfs2 and verify they are as we expect
	validateDI(t, pfs0)
	validateDI(t, pfs1, "/foo11", "/foo21", "/foo31", "/foo41")
	validateDI(t, pfs2, "/foo12", "/foo2", "/foo22", "/foo32", "/foo42")

	in3 := client.NewUnionInput(in1, in2)
	union1, err := NewDatumIterator(c, in3)
	require.NoError(t, err)
	validateDI(t, union1, "/foo11", "/foo21", "/foo31", "/foo41",
		"/foo12", "/foo2", "/foo22", "/foo32", "/foo42")

	in4 := client.NewCrossInput(in1, in2)
	cross1, err := NewDatumIterator(c, in4)
	require.NoError(t, err)
	validateDI(t, cross1,
		"/foo11/foo12", "/foo21/foo12", "/foo31/foo12", "/foo41/foo12",
		"/foo11/foo2", "/foo21/foo2", "/foo31/foo2", "/foo41/foo2",
		"/foo11/foo22", "/foo21/foo22", "/foo31/foo22", "/foo41/foo22",
		"/foo11/foo32", "/foo21/foo32", "/foo31/foo32", "/foo41/foo32",
		"/foo11/foo42", "/foo21/foo42", "/foo31/foo42", "/foo41/foo42",
	)

	in5 := client.NewCrossInput(in3, in4)
	cross2, err := NewDatumIterator(c, in5)
	require.NoError(t, err)
	validateDI(t, cross2,
		"/foo11/foo11/foo12", "/foo21/foo11/foo12", "/foo31/foo11/foo12", "/foo41/foo11/foo12", "/foo12/foo11/foo12", "/foo2/foo11/foo12", "/foo22/foo11/foo12", "/foo32/foo11/foo12", "/foo42/foo11/foo12",
		"/foo11/foo21/foo12", "/foo21/foo21/foo12", "/foo31/foo21/foo12", "/foo41/foo21/foo12", "/foo12/foo21/foo12", "/foo2/foo21/foo12", "/foo22/foo21/foo12", "/foo32/foo21/foo12", "/foo42/foo21/foo12",
		"/foo11/foo31/foo12", "/foo21/foo31/foo12", "/foo31/foo31/foo12", "/foo41/foo31/foo12", "/foo12/foo31/foo12", "/foo2/foo31/foo12", "/foo22/foo31/foo12", "/foo32/foo31/foo12", "/foo42/foo31/foo12",
		"/foo11/foo41/foo12", "/foo21/foo41/foo12", "/foo31/foo41/foo12", "/foo41/foo41/foo12", "/foo12/foo41/foo12", "/foo2/foo41/foo12", "/foo22/foo41/foo12", "/foo32/foo41/foo12", "/foo42/foo41/foo12",
		"/foo11/foo11/foo2", "/foo21/foo11/foo2", "/foo31/foo11/foo2", "/foo41/foo11/foo2", "/foo12/foo11/foo2", "/foo2/foo11/foo2", "/foo22/foo11/foo2", "/foo32/foo11/foo2", "/foo42/foo11/foo2",
		"/foo11/foo21/foo2", "/foo21/foo21/foo2", "/foo31/foo21/foo2", "/foo41/foo21/foo2", "/foo12/foo21/foo2", "/foo2/foo21/foo2", "/foo22/foo21/foo2", "/foo32/foo21/foo2", "/foo42/foo21/foo2",
		"/foo11/foo31/foo2", "/foo21/foo31/foo2", "/foo31/foo31/foo2", "/foo41/foo31/foo2", "/foo12/foo31/foo2", "/foo2/foo31/foo2", "/foo22/foo31/foo2", "/foo32/foo31/foo2", "/foo42/foo31/foo2",
		"/foo11/foo41/foo2", "/foo21/foo41/foo2", "/foo31/foo41/foo2", "/foo41/foo41/foo2", "/foo12/foo41/foo2", "/foo2/foo41/foo2", "/foo22/foo41/foo2", "/foo32/foo41/foo2", "/foo42/foo41/foo2",
		"/foo11/foo11/foo22", "/foo21/foo11/foo22", "/foo31/foo11/foo22", "/foo41/foo11/foo22", "/foo12/foo11/foo22", "/foo2/foo11/foo22", "/foo22/foo11/foo22", "/foo32/foo11/foo22", "/foo42/foo11/foo22",
		"/foo11/foo21/foo22", "/foo21/foo21/foo22", "/foo31/foo21/foo22", "/foo41/foo21/foo22", "/foo12/foo21/foo22", "/foo2/foo21/foo22", "/foo22/foo21/foo22", "/foo32/foo21/foo22", "/foo42/foo21/foo22",
		"/foo11/foo31/foo22", "/foo21/foo31/foo22", "/foo31/foo31/foo22", "/foo41/foo31/foo22", "/foo12/foo31/foo22", "/foo2/foo31/foo22", "/foo22/foo31/foo22", "/foo32/foo31/foo22", "/foo42/foo31/foo22",
		"/foo11/foo41/foo22", "/foo21/foo41/foo22", "/foo31/foo41/foo22", "/foo41/foo41/foo22", "/foo12/foo41/foo22", "/foo2/foo41/foo22", "/foo22/foo41/foo22", "/foo32/foo41/foo22", "/foo42/foo41/foo22",
		"/foo11/foo11/foo32", "/foo21/foo11/foo32", "/foo31/foo11/foo32", "/foo41/foo11/foo32", "/foo12/foo11/foo32", "/foo2/foo11/foo32", "/foo22/foo11/foo32", "/foo32/foo11/foo32", "/foo42/foo11/foo32",
		"/foo11/foo21/foo32", "/foo21/foo21/foo32", "/foo31/foo21/foo32", "/foo41/foo21/foo32", "/foo12/foo21/foo32", "/foo2/foo21/foo32", "/foo22/foo21/foo32", "/foo32/foo21/foo32", "/foo42/foo21/foo32",
		"/foo11/foo31/foo32", "/foo21/foo31/foo32", "/foo31/foo31/foo32", "/foo41/foo31/foo32", "/foo12/foo31/foo32", "/foo2/foo31/foo32", "/foo22/foo31/foo32", "/foo32/foo31/foo32", "/foo42/foo31/foo32",
		"/foo11/foo41/foo32", "/foo21/foo41/foo32", "/foo31/foo41/foo32", "/foo41/foo41/foo32", "/foo12/foo41/foo32", "/foo2/foo41/foo32", "/foo22/foo41/foo32", "/foo32/foo41/foo32", "/foo42/foo41/foo32",
		"/foo11/foo11/foo42", "/foo21/foo11/foo42", "/foo31/foo11/foo42", "/foo41/foo11/foo42", "/foo12/foo11/foo42", "/foo2/foo11/foo42", "/foo22/foo11/foo42", "/foo32/foo11/foo42", "/foo42/foo11/foo42",
		"/foo11/foo21/foo42", "/foo21/foo21/foo42", "/foo31/foo21/foo42", "/foo41/foo21/foo42", "/foo12/foo21/foo42", "/foo2/foo21/foo42", "/foo22/foo21/foo42", "/foo32/foo21/foo42", "/foo42/foo21/foo42",
		"/foo11/foo31/foo42", "/foo21/foo31/foo42", "/foo31/foo31/foo42", "/foo41/foo31/foo42", "/foo12/foo31/foo42", "/foo2/foo31/foo42", "/foo22/foo31/foo42", "/foo32/foo31/foo42", "/foo42/foo31/foo42",
		"/foo11/foo41/foo42", "/foo21/foo41/foo42", "/foo31/foo41/foo42", "/foo41/foo41/foo42", "/foo12/foo41/foo42", "/foo2/foo41/foo42", "/foo22/foo41/foo42", "/foo32/foo41/foo42", "/foo42/foo41/foo42")

	// cross with a zero datum input should also be zero
	in6 := client.NewCrossInput(in3, in0, in2, in4)
	cross3, err := NewDatumIterator(c, in6)
	require.NoError(t, err)
	validateDI(t, cross3)

	// zero cross inside a cross should also be zero
	in7 := client.NewCrossInput(in6, in1)
	cross4, err := NewDatumIterator(c, in7)
	require.NoError(t, err)
	validateDI(t, cross4)

	in8 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)", "$1$2", false)
	in8.Pfs.Commit = commit.ID
	in9 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)", "$2$1", false)
	in9.Pfs.Commit = commit.ID

	join1, err := newJoinDatumIterator(c, []*pps.Input{in8, in9})
	require.NoError(t, err)
	validateDI(t, join1,
		"/foo11/foo11",
		"/foo12/foo21",
		"/foo13/foo31",
		"/foo14/foo41",
		"/foo21/foo12",
		"/foo22/foo22",
		"/foo23/foo32",
		"/foo24/foo42",
		"/foo31/foo13",
		"/foo32/foo23",
		"/foo33/foo33",
		"/foo34/foo43",
		"/foo41/foo14",
		"/foo42/foo24",
		"/foo43/foo34",
		"/foo44/foo44")
}

func benchmarkDatumIterators(j int, b *testing.B) {
	c := tu.GetPachClient(b)
	defer require.NoError(b, c.DeleteAll())
	require.NoError(b, activateEnterprise(c))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dataRepo := tu.UniqueString("TestDatumIteratorPFS_data")
		require.NoError(b, c.CreateRepo(dataRepo))

		// put files in structured in a way so that there are many ways to glob it
		commit, err := c.StartCommit(dataRepo, "master")
		require.NoError(b, err)
		for i := 0; i < 100*j; i++ {
			_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("foo%v", i), strings.NewReader("bar"))
			require.NoError(b, err)
		}

		require.NoError(b, err)
		require.NoError(b, c.FinishCommit(dataRepo, commit.ID))

		// make one with zero datums for testing edge cases
		in0 := client.NewPFSInput(dataRepo, "!(**)")
		in0.Pfs.Commit = commit.ID
		pfs0, err := NewDatumIterator(c, in0)
		require.NoError(b, err)

		in1 := client.NewPFSInput(dataRepo, "/foo?1*")
		in1.Pfs.Commit = commit.ID
		pfs1, err := NewDatumIterator(c, in1)
		require.NoError(b, err)

		in2 := client.NewPFSInput(dataRepo, "/foo*2")
		in2.Pfs.Commit = commit.ID
		pfs2, err := NewDatumIterator(c, in2)
		require.NoError(b, err)

		validateDI(b, pfs0)
		validateDI(b, pfs1)
		validateDI(b, pfs2)

		b.Run("union", func(b *testing.B) {
			in3 := client.NewUnionInput(in1, in2)
			union1, err := NewDatumIterator(c, in3)
			require.NoError(b, err)
			validateDI(b, union1)
		})

		b.Run("cross", func(b *testing.B) {
			in4 := client.NewCrossInput(in1, in2)
			cross1, err := NewDatumIterator(c, in4)
			require.NoError(b, err)
			validateDI(b, cross1)
		})

		b.Run("join", func(b *testing.B) {
			in8 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)*", "$1$2", false)
			in8.Pfs.Commit = commit.ID
			in9 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)*", "$2$1", false)
			in9.Pfs.Commit = commit.ID
			join1, err := newJoinDatumIterator(c, []*pps.Input{in8, in9})
			require.NoError(b, err)
			validateDI(b, join1)
		})

		b.Run("iterated", func(b *testing.B) {
			in3 := client.NewUnionInput(in1, in2)
			in4 := client.NewCrossInput(in1, in2)

			in5 := client.NewCrossInput(in3, in4)
			cross2, err := NewDatumIterator(c, in5)
			require.NoError(b, err)

			// cross with a zero datum input should also be zero
			in6 := client.NewCrossInput(in3, in0, in2, in4)
			cross3, err := NewDatumIterator(c, in6)
			require.NoError(b, err)

			// zero cross inside a cross should also be zero
			in7 := client.NewCrossInput(in6, in1)
			cross4, err := NewDatumIterator(c, in7)
			require.NoError(b, err)

			validateDI(b, cross2)
			validateDI(b, cross3)
			validateDI(b, cross4)

		})
	}
}

func BenchmarkDI1(b *testing.B)  { benchmarkDatumIterators(1, b) }
func BenchmarkDI2(b *testing.B)  { benchmarkDatumIterators(2, b) }
func BenchmarkDI4(b *testing.B)  { benchmarkDatumIterators(4, b) }
func BenchmarkDI8(b *testing.B)  { benchmarkDatumIterators(8, b) }
func BenchmarkDI16(b *testing.B) { benchmarkDatumIterators(16, b) }
func BenchmarkDI32(b *testing.B) { benchmarkDatumIterators(32, b) }

func validateDI(t testing.TB, di DatumIterator, datums ...string) {
	i := 0
	clone := di
	for di.Next() {
		key := ""
		for _, file := range di.Datum() {
			key += file.FileInfo.File.Path
		}

		key2 := ""
		clone.DatumN(0)
		for _, file := range clone.DatumN(i) {
			key2 += file.FileInfo.File.Path
		}

		if len(datums) > 0 {
			require.Equal(t, datums[i], key)
		}
		require.Equal(t, key2, key)
		i++
	}
	if len(datums) > 0 {
		require.Equal(t, di.Len(), len(datums))
	}
	require.Equal(t, di.Len(), i)
}
