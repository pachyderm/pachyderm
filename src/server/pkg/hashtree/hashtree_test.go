package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"

	bolt "github.com/coreos/bbolt"
)

// obj parses a string as an Object
func obj(s ...string) []*pfs.Object {
	result := make([]*pfs.Object, len(s))
	for i, ss := range s {
		result[i] = &pfs.Object{}
		err := proto.UnmarshalText(ss, result[i])
		if err != nil {
			panic(err)
		}
	}
	return result
}

// Convenience function to convert a list of strings to []interface{} for
// EqualOneOf
func i(ss ...string) []string {
	result := make([]string, len(ss))
	copy(result, ss)
	return result
}

func TestFmtNodeType(t *testing.T) {
	require.Equal(t, "directory", directory.String())
	require.Equal(t, "file", file.String())
	require.Equal(t, "none", none.String())
	require.Equal(t, "unknown", unrecognized.String())
}

func getT(t *testing.T, h HashTree, path string) *NodeProto {
	t.Helper()
	node, err := h.Get(path)
	require.NoError(t, err)
	return node
}

func lenT(t *testing.T, h HashTree) int {
	switch h := h.(type) {
	default:
		panic(fmt.Sprintf("unrecognized hashtree type: %v", t))
	case *dbHashTree:
		result := 0
		require.NoError(t, h.View(func(tx *bolt.Tx) error {
			return fs(tx).ForEach(func(_, _ []byte) error {
				result++
				return nil
			})
		}))
		return result
	}
}

// requireSame compares 'h' to another hash tree (e.g. to make sure that it
// hasn't changed)
func requireSame(t *testing.T, l, r HashTree) {
	t.Helper()
	lRoot, err := l.Get("")
	require.NoError(t, err)
	rRoot, err := r.Get("")
	require.NoError(t, err)
	require.True(t, bytes.Equal(lRoot.Hash, rRoot.Hash))
}

// requireOperationInvariant makes sure that h isn't affected by calling 'op'.
// Good for checking that adding and deleting a file does nothing persistent,
// etc. This is separate from 'requireSame()' because often we want to test that
// an operation is invariant on several slightly different trees, and with this
// we only have to define 'op' once.
func requireOperationInvariant(t *testing.T, h HashTree, op func()) {
	t.Helper()
	preop, err := h.Copy()
	require.NoError(t, err)
	// perform operation on 'h'
	op()
	// Make sure 'h' is still the same
	require.NoError(t, h.Hash())
	requireSame(t, preop, h)
}

func newHashTree(tb testing.TB) HashTree {
	result, err := NewDBHashTree("")
	require.NoError(tb, err)
	return result
}

func TestPutFileBasic(t *testing.T) {
	// Put a file
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.Hash())
	require.Equal(t, int64(1), getT(t, h, "/foo").SubtreeSize)
	require.Equal(t, int64(1), getT(t, h, "").SubtreeSize)

	// Put a file under a directory and make sure changes are propagated upwards
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.Hash())
	require.Equal(t, int64(1), getT(t, h, "/dir/bar").SubtreeSize)
	require.Equal(t, int64(1), getT(t, h, "/dir").SubtreeSize)
	require.Equal(t, int64(2), getT(t, h, "").SubtreeSize)

	// Put another file
	require.NoError(t, h.PutFile("/dir/buzz", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h.PutFile("/dir.buzz", obj(`hash:"4ab7d"`), 1))
	require.NoError(t, h.Hash())
	// inspect h
	require.Equal(t, int64(1), getT(t, h, "/dir/buzz").SubtreeSize)
	require.Equal(t, int64(2), getT(t, h, "/dir").SubtreeSize)
	require.Equal(t, int64(1), getT(t, h, "/dir.buzz").SubtreeSize)
	require.Equal(t, int64(4), getT(t, h, "").SubtreeSize)
	nodes, err := h.ListAll("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("foo", "dir", "dir.buzz"), node.Name)
	}

	nodes, err = h.ListAll("/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("bar", "buzz"), node.Name)
	}
	require.Equal(t, int64(1), getT(t, h, "/foo").SubtreeSize)

	// Make sure subsequent PutFile calls append
	h2, err := h.Copy()
	require.NoError(t, err)
	require.NoError(t, h2.PutFile("/foo", obj(`hash:"413e7"`), 1))
	require.NoError(t, h2.Hash())
	require.NotEqual(t, getT(t, h, "/foo").Hash, getT(t, h2, "/foo").Hash)
	require.Equal(t, int64(2), getT(t, h2, "/foo").SubtreeSize)
}

func TestPutDirBasic(t *testing.T) {
	h := newHashTree(t)
	emptySha := sha256.Sum256([]byte{})

	// put a directory
	require.NoError(t, h.PutDir("/dir"))
	require.Equal(t, lenT(t, h), 2) // "/dir" and "/"
	require.Equal(t, []string(nil), getT(t, h, "/dir").DirNode.Children)
	require.NoError(t, h.Hash())
	require.Equal(t, []string(nil), getT(t, h, "/dir").DirNode.Children)
	require.Equal(t, emptySha[:], getT(t, h, "/dir").Hash)
	require.Equal(t, lenT(t, h), 2)

	// put a directory under another directory
	require.NoError(t, h.PutDir("/dir/foo"))
	require.NotEqual(t, []string{}, getT(t, h, "/dir").DirNode.Children)
	require.NoError(t, h.Hash())
	require.NotEqual(t, []string{}, getT(t, h, "/dir").DirNode.Children)
	nodes, err := h.ListAll("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	require.NotEqual(t, emptySha[:], getT(t, h, "/dir").Hash)

	// delete the directory
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	require.NoError(t, h.Hash())
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	nodes, err = h.ListAll("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], getT(t, h, "/dir").Hash)

	// Make sure that deleting a dir also deletes files under the dir
	require.NoError(t, h.PutFile("/dir/foo/bar", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	require.Equal(t, lenT(t, h), 2)
	require.NoError(t, h.Hash())
	require.NoError(t, err)
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	nodes, err = h.ListAll("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], getT(t, h, "/dir").Hash)
	require.Equal(t, lenT(t, h), 2)
}

func TestPutError(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))

	// PutFile fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo/bar", obj(`hash:"8e02c"`), 1)
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
		node, err := h.Get("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathNotFound, Code(err))
		require.Nil(t, node)
	})

	// PutDir fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutDir("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
		node, err := h.Get("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathNotFound, Code(err))
		require.Nil(t, node)
	})

	// PutFile fails if a directory already exists (put /foo when /foo/bar exists)
	require.NoError(t, h.DeleteFile("/foo"))
	require.NoError(t, h.PutFile("/foo/bar", obj(`hash:"ebc57"`), 1))
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo", obj(`hash:"8e02c"`), 1)
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
	})
}

// Given a directory D, test that adding and then deleting a file/directory to
// D does not change D.
func TestAddDeleteReverts(t *testing.T) {
	h := newHashTree(t)
	addDeleteFile := func() {
		require.NoError(t, h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
		require.NoError(t, h.DeleteFile("/dir/__NEW_FILE__"))
	}
	addDeleteDir := func() {
		require.NoError(t, h.PutDir("/dir/__NEW_DIR__"))
		require.NoError(t, h.DeleteFile("/dir/__NEW_DIR__"))
	}
	addDeleteSubFile := func() {
		require.NoError(t, h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
		require.NoError(t, h.DeleteFile("/dir/__NEW_DIR__"))
	}

	require.NoError(t, h.PutDir("/dir"))
	requireOperationInvariant(t, h, addDeleteFile)
	requireOperationInvariant(t, h, addDeleteDir)
	requireOperationInvariant(t, h, addDeleteSubFile)
	// Add some files to make sure the test still passes when D already has files
	// in it.
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"20c27"`), 1))
	requireOperationInvariant(t, h, addDeleteFile)
	requireOperationInvariant(t, h, addDeleteDir)
	requireOperationInvariant(t, h, addDeleteSubFile)
}

// Given a directory D, test that deleting and then adding a file/directory to
// D does not change D.
func TestDeleteAddReverts(t *testing.T) {
	h := newHashTree(t)
	deleteAddFile := func() {
		require.NoError(t, h.DeleteFile("/dir/__NEW_FILE__"))
		require.NoError(t, h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	}
	deleteAddDir := func() {
		require.NoError(t, h.DeleteFile("/dir/__NEW_DIR__"))
		require.NoError(t, h.PutDir("/dir/__NEW_DIR__"))
	}
	deleteAddSubFile := func() {
		require.NoError(t, h.DeleteFile("/dir/__NEW_DIR__"))
		require.NoError(t, h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	}

	require.NoError(t, h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	requireOperationInvariant(t, h, deleteAddFile)
	require.NoError(t, h.PutDir("/dir/__NEW_DIR__"))
	requireOperationInvariant(t, h, deleteAddDir)
	require.NoError(t, h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	requireOperationInvariant(t, h, deleteAddSubFile)

	// Make sure test still passes when trees are nonempty
	h = newHashTree(t)
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"20c27"`), 1))

	require.NoError(t, h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	requireOperationInvariant(t, h, deleteAddFile)
	require.NoError(t, h.PutDir("/dir/__NEW_DIR__"))
	requireOperationInvariant(t, h, deleteAddDir)
	require.NoError(t, h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1))
	requireOperationInvariant(t, h, deleteAddSubFile)
}

// The hash of a directory doesn't change no matter what order files are added
// to it.
func TestPutFileCommutative(t *testing.T) {
	h := newHashTree(t)
	h2 := newHashTree(t)
	// Puts files into h in the order [A, B] and into h2 in the order [B, A]
	comparePutFiles := func() {
		require.NoError(t, h.PutFile("/dir/__NEW_FILE_A__", obj(`hash:"ebc57"`), 1))
		require.NoError(t, h.PutFile("/dir/__NEW_FILE_B__", obj(`hash:"20c27"`), 1))

		// Get state of both /dir and /, to make sure changes are preserved upwards
		// through the file hierarchy
		dirNodePtr, err := h.Get("/dir")
		require.NoError(t, err)
		rootNodePtr, err := h.Get("/")
		require.NoError(t, err)

		require.NoError(t, h2.PutFile("/dir/__NEW_FILE_B__", obj(`hash:"20c27"`), 1))
		require.NoError(t, h2.PutFile("/dir/__NEW_FILE_A__", obj(`hash:"ebc57"`), 1))

		dirNodePtr2, err := h2.Get("/dir")
		require.NoError(t, err)
		rootNodePtr2, err := h2.Get("/")
		require.NoError(t, err)
		require.Equal(t, *dirNodePtr, *dirNodePtr2)
		require.Equal(t, *rootNodePtr, *rootNodePtr2)
	}

	// (1) Run the test on empty trees
	comparePutFiles()
	// (2) Add some files & check that test still passes when trees are nonempty
	h, h2 = newHashTree(t), newHashTree(t)
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h2.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"9d432"`), 1))
	require.NoError(t, h2.PutFile("/dir/bar", obj(`hash:"9d432"`), 1))
	comparePutFiles()
}

// Given a directory D, renaming (removing and re-adding under a different name)
// a file or directory under D changes the hash of D, even if the contents are
// identical.
func TestRenameChangesHash(t *testing.T) {
	// Write a file, and then get the hash of every node from the file to the root
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.Hash())
	dirPre, err := h.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h.Get("/")
	require.NoError(t, err)

	// rename /dir/foo to /dir/bar, and make sure that changes the hash
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.Hash())

	dirPost, err := h.Get("/dir")
	require.NoError(t, err)
	rootPost, err := h.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, dirPre.Hash, dirPost.Hash)
	require.NotEqual(t, rootPre.Hash, rootPost.Hash)
	require.Equal(t, dirPre.SubtreeSize, dirPost.SubtreeSize)
	require.Equal(t, rootPre.SubtreeSize, rootPost.SubtreeSize)

	// rename /dir to /dir2, and make sure that changes the hash
	require.NoError(t, h.DeleteFile("/dir"))
	require.NoError(t, h.PutFile("/dir2/foo", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.Hash())

	dirPost, err = h.Get("/dir2")
	require.NoError(t, err)
	rootPost, err = h.Get("/")
	require.NoError(t, err)

	require.Equal(t, dirPre.Hash, dirPost.Hash) // dir == dir2
	require.NotEqual(t, rootPre.Hash, rootPost.Hash)
	require.Equal(t, dirPre.SubtreeSize, dirPost.SubtreeSize)
	require.Equal(t, rootPre.SubtreeSize, rootPost.SubtreeSize)
}

// Given a directory D, rewriting (removing and re-adding a different file
// under the same name) a file or directory under D changes the hash of D, even
// if the contents are identical.
func TestRewriteChangesHash(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.Hash())

	dirPre, err := h.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h.Get("/")
	require.NoError(t, err)

	// Change
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.NoError(t, h.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h.Hash())

	dirPost, err := h.Get("/dir")
	require.NoError(t, err)
	rootPost, err := h.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, dirPre.Hash, dirPost.Hash)
	require.NotEqual(t, rootPre.Hash, rootPost.Hash)
	require.Equal(t, dirPre.SubtreeSize, dirPost.SubtreeSize)
	require.Equal(t, rootPre.SubtreeSize, rootPost.SubtreeSize)
}

func TestIsGlob(t *testing.T) {
	require.True(t, IsGlob(`*`))
	require.True(t, IsGlob(`path/to*/file`))
	require.True(t, IsGlob(`path/**/file`))
	require.True(t, IsGlob(`path/to/f?le`))
	require.True(t, IsGlob(`pa!h/to/file`))
	require.True(t, IsGlob(`pa[th]/to/file`))
	require.True(t, IsGlob(`pa{th}/to/file`))
	require.True(t, IsGlob(`*/*`))
	require.False(t, IsGlob(`path`))
	require.False(t, IsGlob(`path/to/file1.txt`))
	require.False(t, IsGlob(`path/to_test-a/file.txt`))
}

func TestGlobPrefix(t *testing.T) {
	require.Equal(t, `/`, GlobLiteralPrefix(`*`))
	require.Equal(t, `/`, GlobLiteralPrefix(`**`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/*`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/**`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/(*)`))
	require.Equal(t, `/di`, GlobLiteralPrefix(`di?/(*)`))
	require.Equal(t, `/di`, GlobLiteralPrefix(`di?[rg]`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/@(a)`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/+(a)`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/{foo,bar}`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/(a|b)`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/^[a-z]`))
	require.Equal(t, `/dir/`, GlobLiteralPrefix(`dir/[!abc]`))
}

func TestGlobFile(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.PutFile("/dir/buzz", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h.Hash())

	// patterns that match the whole repo ("/")
	addTo := func(ss *[]string) func(string, *NodeProto) error {
		return func(s string, _ *NodeProto) error {
			*ss = append(*ss, s)
			return nil
		}
	}
	for _, pattern := range []string{"", "/"} {
		var paths []string
		require.NoError(t, h.Glob(pattern, addTo(&paths)))
		require.ElementsEqual(t, i("/"), paths)
	}

	// patterns that match top-level dirs/files
	for _, pattern := range []string{"*", "/*"} {
		var paths []string
		require.NoError(t, h.Glob(pattern, addTo(&paths)))
		require.ElementsEqual(t, i("/foo", "/dir"), paths)
	}

	// patterns that match second-level dirs/files
	for _, pattern := range []string{"dir/*", "/dir/*", "*/*", "/*/*"} {
		var paths []string
		require.NoError(t, h.Glob(pattern, addTo(&paths)))
		require.ElementsEqual(t, i("/dir/bar", "/dir/buzz"), paths)
	}
}

// Test that Walk() works
func TestWalk(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1))
	require.NoError(t, h.PutFile("/dir2/buzz", obj(`hash:"fa347"`), 1))
	require.NoError(t, h.PutFile("/dir.bar", obj(`hash:"3ead7"`), 1))
	require.NoError(t, h.Hash())

	expectedPaths := []string{"/", "/dir", "/dir/bar", "/dir.bar", "/dir2", "/dir2/buzz", "/foo"}
	i := 0
	require.NoError(t, h.Walk("/", func(path string, node *NodeProto) error {
		require.Equal(t, expectedPaths[i], path)
		i++
		return nil
	}))
	require.Equal(t, len(expectedPaths), i)

	expectedPaths = []string{"/dir", "/dir/bar"}
	i = 0
	require.NoError(t, h.Walk("/dir", func(path string, node *NodeProto) error {
		require.Equal(t, expectedPaths[i], path)
		i++
		return nil
	}))
	require.Equal(t, len(expectedPaths), i)
}

// Test that HashTree methods return the right error codes
func TestErrorCode(t *testing.T) {
	require.Equal(t, OK, Code(nil))
	require.Equal(t, Unknown, Code(errors.Errorf("external error")))

	h := newHashTree(t)
	_, err := h.Get("/path")
	require.Equal(t, PathNotFound, Code(err))

	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	err = h.PutFile("/foo/bar", obj(`hash:"9d432"`), 1)
	require.Equal(t, PathConflict, Code(err))
	require.NoError(t, h.PutFile("/bar/foo", obj(`hash:"9d432"`), 1))
	err = h.PutFile("/bar", obj(`hash:"20c27"`), 1)
	require.Equal(t, PathConflict, Code(err))
}

func TestSerialize(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.PutFile("/bar/buzz", obj(`hash:"9d432"`), 1))
	require.NoError(t, h.Hash())

	// Serialize and Deserialize 'h'
	var buf bytes.Buffer
	require.NoError(t, h.Serialize(&buf))
	h2, err := DeserializeDBHashTree("", &buf)
	require.NoError(t, err)
	requireSame(t, h, h2)

	// Modify 'h', and Serialize and Deserialize it again
	require.NoError(t, h.PutFile("/bar/buzz2", obj(`hash:"8e02c"`), 1))
	require.NoError(t, h.Hash())
	buf.Reset()
	require.NoError(t, h.Serialize(&buf))
	h3, err := DeserializeDBHashTree("", &buf)
	require.NoError(t, err)
	requireSame(t, h, h3)

	// Make sure 'h2' does not equal 'h' or 'h3'
	// require.False(t, proto.Equal(h, h2.(*HashTreeProto)))
	// require.False(t, proto.Equal(h2.(*HashTreeProto), h3.(*HashTreeProto)))
}

// func TestSerializeError(t *testing.T) {
// 	// Test version
// 	h := &HashTreeProto{Version: -1}
// 	bts, err := h.Marshal()
// 	require.NoError(t, err)
// 	h3 := &dbHashTree{}
// 	_, err = Deserialize(bts)
// 	require.YesError(t, err)
// 	require.Equal(t, Unsupported, Code(err))
// }

func TestListEmpty(t *testing.T) {
	tree := newHashTree(t)
	_, err := tree.ListAll("/")
	require.NoError(t, err)
	nop := func(string, *NodeProto) error { return nil }
	require.NoError(t, tree.Glob("*", nop))
	require.NoError(t, tree.Glob("/*", nop))

	require.NoError(t, tree.DeleteFile("/"))
	require.NoError(t, tree.DeleteFile(""))

	_, err = tree.ListAll("/")
	require.NoError(t, err)
	require.NoError(t, tree.Glob("*", nop))
	require.NoError(t, tree.Glob("/*", nop))
}

func diffTrees(t *testing.T, new, old HashTree, path string) ([]string, []string) {
	var newFiles []string
	var oldFiles []string
	collect := func(path string, node *NodeProto, new bool) error {
		if new {
			newFiles = append(newFiles, path)
		} else {
			oldFiles = append(oldFiles, path)
		}
		return nil
	}
	require.NoError(t, new.Diff(old, path, path, -1, collect))
	return newFiles, oldFiles
}

func TestDiff(t *testing.T) {
	old := newHashTree(t)
	require.NoError(t, old.PutFile("/foo", obj(`hash:"4a2e9"`), 1))
	require.NoError(t, old.PutFile("/dir/bar", obj(`hash:"10ead"`), 1))
	require.NoError(t, old.Hash())
	new, err := old.Copy()
	require.NoError(t, err)
	newFiles, oldFiles := diffTrees(t, new, old, "")
	require.Equal(t, 0, len(newFiles))
	require.Equal(t, 0, len(oldFiles))
	require.NoError(t, new.PutFile("/buzz", obj(`hash:"20afd"`), 1))
	require.NoError(t, new.Hash())
	newFiles, oldFiles = diffTrees(t, new, old, "")
	require.Equal(t, 1, len(newFiles))
	require.Equal(t, 0, len(oldFiles))
}

func TestChildIterator(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("a/1", obj(`hash:"23ea6"`), 1))
	require.NoError(t, h.PutFile("b/2", obj(`hash:"92fbc"`), 1))
	require.NoError(t, h.Hash())
	h.(*dbHashTree).View(func(tx *bolt.Tx) error {
		c := NewChildCursor(tx, "")
		require.Equal(t, "/a", s(c.K()))
		c.Next()
		require.Equal(t, "/b", s(c.K()))
		return nil
	})

	h.(*dbHashTree).View(func(tx *bolt.Tx) error {
		c := NewChildCursor(tx, "a")
		require.Equal(t, "/a/1", s(c.K()))
		c.Next()
		require.Nil(t, c.K())
		return nil
	})

	h.(*dbHashTree).View(func(tx *bolt.Tx) error {
		c := NewChildCursor(tx, "b")
		require.Equal(t, "/b/2", s(c.K()))
		c.Next()
		require.Nil(t, c.K())
		return nil
	})
}

func TestCache(t *testing.T) {
	// Cache with size 2
	c, err := NewCache(2)
	require.NoError(t, err)

	// Set eviction to be synchronous to avoid polling later
	*c.syncEvict = true

	h := newHashTree(t)
	require.NoError(t, h.PutFile("foo", obj(`hash:"1d4a7"`), 1))
	require.NoError(t, h.Hash())

	// Put a hashtree into the cache
	h1, err := c.GetOrAdd(1, func() (HashTree, error) { return h, nil })
	require.NoError(t, err)
	require.Equal(t, h, h1)
	require.Equal(t, 1, c.lruCache.Len())

	// Get the hashtree from the same key
	h1, err = c.GetOrAdd(1, func() (HashTree, error) { return h, nil })
	require.NoError(t, err)
	require.Equal(t, h, h1)
	require.Equal(t, 1, c.lruCache.Len())

	// Fail to instantiate a hashtree for a new key
	_, err = c.GetOrAdd(4, func() (HashTree, error) { return nil, errors.Errorf("error") })
	require.YesError(t, err)
	require.Equal(t, 1, c.lruCache.Len())

	// After adding a second hashtree, the first hashtree should still be in the cache
	h2, err := c.GetOrAdd(2, func() (HashTree, error) { return newHashTree(t), nil })
	require.NoError(t, err)
	require.Equal(t, 2, c.lruCache.Len())
	_, err = h.Get("foo")
	require.NoError(t, err)

	// But after adding a third, the first one should be evicted
	h3, err := c.GetOrAdd(3, func() (HashTree, error) { return newHashTree(t), nil })
	require.NoError(t, err)
	require.Equal(t, 2, c.lruCache.Len())

	_, err = h1.ListAll("")
	require.YesError(t, err)
	_, err = h2.ListAll("")
	require.NoError(t, err)
	_, err = h3.ListAll("")
	require.NoError(t, err)

	c.Close()
	_, err = h2.ListAll("")
	require.YesError(t, err)
	_, err = h3.ListAll("")
	require.YesError(t, err)
	require.Equal(t, 0, c.lruCache.Len())
}

func blocks(ss ...string) []*pfs.BlockRef {
	var blockRefs []*pfs.BlockRef
	for i, s := range ss {
		blockRefs = append(blockRefs, &pfs.BlockRef{})
		err := proto.UnmarshalText(s, blockRefs[i])
		if err != nil {
			panic(err)
		}
	}
	return blockRefs
}

func writeMergeNode(w *Writer, path, hash string, size int64, blockRefs ...*pfs.BlockRef) {
	path = clean(path)
	n := mergeNode(path, hash, size)
	if len(blockRefs) == 0 {
		n.nodeProto.DirNode = &DirectoryNodeProto{}
	} else {
		n.nodeProto.FileNode = &FileNodeProto{BlockRefs: blockRefs}
	}
	w.Write(n)
}

func mergeNode(path, hash string, size int64) *MergeNode {
	h, err := pfs.DecodeHash(hash)
	if err != nil {
		panic(err)
	}
	return &MergeNode{
		k: b(path),
		nodeProto: &NodeProto{
			Name:        base(path),
			Hash:        h,
			SubtreeSize: size,
		},
	}
}

func TestMergeFiles(t *testing.T) {
	c, err := NewMergeCache("merge-cache-test")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, c.Close())
	}()

	l, r := NewUnordered(""), NewUnordered("")
	l.PutFile("/foo-left", []byte("l0"), 1, blocks(``)...)
	l.PutFile("/dir-left/bar-left", []byte("l1"), 1, blocks(``)...)
	l.PutFile("/dir-shared/buzz-left", []byte("l2"), 1, blocks(``)...)
	l.PutFile("/dir-shared/file-shared", []byte("l3"), 1, blocks(``)...)
	r.PutFile("/foo-right", []byte("r0"), 1, blocks(``)...)
	r.PutFile("/dir-right/bar-right", []byte("r1"), 1, blocks(``)...)
	r.PutFile("/dir-shared/buzz-right", []byte("r2"), 1, blocks(``)...)
	r.PutFile("/dir-shared/file-shared", []byte("r3"), 1, blocks(``)...)
	lBuf, rBuf, resultBuf := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	require.NoError(t, l.Ordered().Serialize(lBuf))
	require.NoError(t, r.Ordered().Serialize(rBuf))
	require.NoError(t, c.Put("0", lBuf))
	require.NoError(t, c.Put("1", rBuf))
	require.NoError(t, c.Merge(NewWriter(resultBuf), nil, nil))

	expectedBuf := &bytes.Buffer{}
	w := NewWriter(expectedBuf)
	writeMergeNode(w, "/", "5d5e6bf6978265596cc1302f0bc368be893a4c14ce742b76dc20e983eac3446c7b022672c53ed6d80314b0c743fd22f50c07a09a79775165ce063ce984a91946", 8)
	writeMergeNode(w, "/dir-left", "9cf1225943e6797220168992de346051cdc762b9c1ba99bd6bdcc12ec3f1fea7", 1)
	writeMergeNode(w, "/dir-left/bar-left", "6c31", 1, blocks(``)...)
	writeMergeNode(w, "/dir-right", "7dfa4a4878ffe3acb6574123fb775c7758c97270742e3801a093c0b1ea3cb9fc", 1)
	writeMergeNode(w, "/dir-right/bar-right", "7231", 1, blocks(``)...)
	writeMergeNode(w, "/dir-shared", "4302fc8eaa65188e71b5d298cc4c40f95bb91cad8dd1fc9db29529514daad51dacfff5ae58d0b8ac4c65beb4926236261b3eff994e2a3b22b0caed3434e17060", 4)
	writeMergeNode(w, "/dir-shared/buzz-left", "6c32", 1, blocks(``)...)
	writeMergeNode(w, "/dir-shared/buzz-right", "7232", 1, blocks(``)...)
	writeMergeNode(w, "/dir-shared/file-shared", "ddc7def93be72db4ed4467edb815395d2bd191c231d0c5f8705dbedb465787202f377e58eb3ff7cd3d63e6c6bfd6ab079328937ed3dc12795f69810d25ed1afa", 2, blocks(``, ``)...)
	writeMergeNode(w, "/foo-left", "6c30", 1, blocks(``)...)
	writeMergeNode(w, "/foo-right", "7230", 1, blocks(``)...)

	require.Equal(t, expectedBuf, resultBuf)
}
