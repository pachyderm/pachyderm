package datum

import (
	"archive/tar"
	"bytes"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	pfstesting "github.com/pachyderm/pachyderm/src/server/pfs/server/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

func TestIteratorsV2(t *testing.T) {
	config := pfstesting.NewPachdConfig()
	config.StorageV2 = true
	require.NoError(t, testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		c := env.PachClient
		dataRepo := tu.UniqueString(t.Name() + "_data")
		require.NoError(t, c.CreateRepo(dataRepo))
		// Put files in structured in a way so that there are many ways to glob it.
		commit, err := c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		buf := &bytes.Buffer{}
		tw := tar.NewWriter(buf)
		for i := 0; i < 50; i++ {
			require.NoError(t, writeFile(tw, &testFile{
				name: fmt.Sprintf("/foo%v", i),
				data: []byte("input"),
			}))
		}
		require.NoError(t, c.PutTarV2(dataRepo, commit.ID, buf))
		require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
		// Zero datums.
		in0 := client.NewPFSInput(dataRepo, "!(**)")
		in0.Pfs.Commit = commit.ID
		t.Run("ZeroDatums", func(t *testing.T) {
			pfs0, err := NewIteratorV2(c, in0)
			require.NoError(t, err)
			validateDIV2(t, pfs0)
		})
		// Basic PFS inputs
		in1 := client.NewPFSInput(dataRepo, "/foo?1")
		in1.Pfs.Commit = commit.ID
		in2 := client.NewPFSInput(dataRepo, "/foo*2")
		in2.Pfs.Commit = commit.ID
		t.Run("Basic", func(t *testing.T) {
			pfs1, err := NewIteratorV2(c, in1)
			require.NoError(t, err)
			pfs2, err := NewIteratorV2(c, in2)
			require.NoError(t, err)
			validateDIV2(t, pfs1, "/foo11", "/foo21", "/foo31", "/foo41")
			validateDIV2(t, pfs2, "/foo12", "/foo2", "/foo22", "/foo32", "/foo42")
		})
		// Union input.
		in3 := client.NewUnionInput(in1, in2)
		t.Run("Union", func(t *testing.T) {
			union1, err := NewIteratorV2(c, in3)
			require.NoError(t, err)
			validateDIV2(t, union1, "/foo11", "/foo21", "/foo31", "/foo41",
				"/foo12", "/foo2", "/foo22", "/foo32", "/foo42")
		})
		// Cross input.
		in4 := client.NewCrossInput(in1, in2)
		t.Run("Cross", func(t *testing.T) {
			cross1, err := NewIteratorV2(c, in4)
			require.NoError(t, err)
			validateDIV2(t, cross1,
				"/foo11/foo12", "/foo11/foo2", "/foo11/foo22", "/foo11/foo32", "/foo11/foo42",
				"/foo21/foo12", "/foo21/foo2", "/foo21/foo22", "/foo21/foo32", "/foo21/foo42",
				"/foo31/foo12", "/foo31/foo2", "/foo31/foo22", "/foo31/foo32", "/foo31/foo42",
				"/foo41/foo12", "/foo41/foo2", "/foo41/foo22", "/foo41/foo32", "/foo41/foo42",
			)
		})
		// Empty cross.
		in6 := client.NewCrossInput(in3, in0, in2, in4)
		t.Run("EmptyCross", func(t *testing.T) {
			cross3, err := NewIteratorV2(c, in6)
			require.NoError(t, err)
			validateDIV2(t, cross3)
		})
		// Nested empty cross.
		in7 := client.NewCrossInput(in6, in1)
		t.Run("NestedEmptyCross", func(t *testing.T) {
			cross4, err := NewIteratorV2(c, in7)
			require.NoError(t, err)
			validateDIV2(t, cross4)
		})
		return nil
	}, config))

	//      TODO: Convert these tests when join and s3 inputs are supported.
	//	// in[8-9] are elements of in10, which is a join input
	//	in8 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)", "$1$2", false)
	//	in8.Pfs.Commit = commit.ID
	//	in9 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)", "$2$1", false)
	//	in9.Pfs.Commit = commit.ID
	//	in10 := client.NewJoinInput(in8, in9)
	//	t.Run("Join", func(t *testing.T) {
	//		join1, err := NewIterator(c, in10)
	//		require.NoError(t, err)
	//		validateDI(t, join1,
	//			"/foo11/foo11",
	//			"/foo12/foo21",
	//			"/foo13/foo31",
	//			"/foo14/foo41",
	//			"/foo21/foo12",
	//			"/foo22/foo22",
	//			"/foo23/foo32",
	//			"/foo24/foo42",
	//			"/foo31/foo13",
	//			"/foo32/foo23",
	//			"/foo33/foo33",
	//			"/foo34/foo43",
	//			"/foo41/foo14",
	//			"/foo42/foo24",
	//			"/foo43/foo34",
	//			"/foo44/foo44")
	//	})
	//
	//	// in11 is an S3 input
	//	in11 := client.NewS3PFSInput("", dataRepo, "")
	//	in11.Pfs.Commit = commit.ID
	//	t.Run("PlainS3", func(t *testing.T) {
	//		s3itr, err := NewIterator(c, in11)
	//		require.NoError(t, err)
	//		validateDI(t, s3itr, "/")
	//
	//		// Check that every datum has an S3 input
	//		s3itr, _ = NewIterator(c, in11)
	//		var checked, s3Count int
	//		for s3itr.Next() {
	//			checked++
	//			require.Equal(t, 1, len(s3itr.Datum()))
	//			if s3itr.Datum()[0].S3 {
	//				s3Count++
	//				break
	//			}
	//		}
	//		require.True(t, checked > 0 && checked == s3Count,
	//			"checked: %v, s3Count: %v", checked, s3Count)
	//	})
	//
	//	// in12 is a cross that contains an S3 input and two non-s3 inputs
	//	in12 := client.NewCrossInput(in1, in2, in11)
	//	t.Run("S3MixedCross", func(t *testing.T) {
	//		s3CrossItr, err := NewIterator(c, in12)
	//		require.NoError(t, err)
	//		validateDI(t, s3CrossItr,
	//			"/foo11/foo12/", "/foo21/foo12/", "/foo31/foo12/", "/foo41/foo12/",
	//			"/foo11/foo2/", "/foo21/foo2/", "/foo31/foo2/", "/foo41/foo2/",
	//			"/foo11/foo22/", "/foo21/foo22/", "/foo31/foo22/", "/foo41/foo22/",
	//			"/foo11/foo32/", "/foo21/foo32/", "/foo31/foo32/", "/foo41/foo32/",
	//			"/foo11/foo42/", "/foo21/foo42/", "/foo31/foo42/", "/foo41/foo42/",
	//		)
	//
	//		s3CrossItr, _ = NewIterator(c, in12)
	//		var checked, s3Count int
	//		for s3CrossItr.Next() {
	//			checked++
	//			for _, d := range s3CrossItr.Datum() {
	//				if d.S3 {
	//					s3Count++
	//				}
	//			}
	//		}
	//		require.True(t, checked > 0 && checked == s3Count,
	//			"checked: %v, s3Count: %v", checked, s3Count)
	//	})
	//
	//	// in13 is a cross consisting of exclusively S3 inputs
	//	in13 := client.NewCrossInput(in11, in11, in11)
	//	t.Run("S3OnlyCrossUnionJoin", func(t *testing.T) {
	//		s3CrossItr, err := NewIterator(c, in13)
	//		require.NoError(t, err)
	//		validateDI(t, s3CrossItr, "///")
	//
	//		s3CrossItr, _ = NewIterator(c, in13)
	//		var checked, s3Count int
	//		for s3CrossItr.Next() {
	//			checked++
	//			for _, d := range s3CrossItr.Datum() {
	//				if d.S3 {
	//					s3Count++
	//				}
	//			}
	//		}
	//		require.True(t, checked > 0 && 3*checked == s3Count,
	//			"checked: %v, s3Count: %v", checked, s3Count)
	//	})
}

func validateDIV2(t testing.TB, di IteratorV2, datums ...string) {
	t.Helper()
	require.NoError(t, di.Iterate(func(inputs []*common.InputV2) error {
		var key string
		for _, input := range inputs {
			key += input.FileInfo.File.Path
		}
		require.Equal(t, datums[0], key)
		datums = datums[1:]
		return nil
	}))
	require.Equal(t, 0, len(datums))
}
