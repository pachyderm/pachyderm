package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"runtime"
	"testing"

	bolt "github.com/coreos/bbolt"
	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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
	for i, v := range ss {
		result[i] = v
	}
	return result
}

func tostring(hTmp OpenHashTree) string {
	h := hTmp.(*hashtree)
	bufsize := len(h.fs) * 25
	buf := bytes.NewBuffer(make([]byte, 0, bufsize))
	for k, v := range h.fs {
		buf.WriteString(fmt.Sprintf("\"%s\": %+v\n", k, v))
	}
	return buf.String()
}

func equals(lTmp, rTmp OpenHashTree) bool {
	l, r := lTmp.(*hashtree), rTmp.(*hashtree)
	if len(l.fs) != len(r.fs) {
		return false
	}
	for path, lv := range l.fs {
		rv, ok := r.fs[path]
		if !ok {
			return false
		}
		// Don't compare hash, since that's not meaningful for OpenHashTrees
		if lv.Name != rv.Name {
			return false
		}
		if !proto.Equal(lv.DirNode, rv.DirNode) ||
			!proto.Equal(lv.FileNode, rv.FileNode) {
			return false
		}
	}
	return true
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
	case *hashtree:
		return len(h.fs)
	case *HashTreeProto:
		return len(h.Fs)
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

func finish(t *testing.T, h OpenHashTree) HashTree {
	h2, err := h.Finish()
	require.NoError(t, err)
	return h2
}

// requireSame compares 'h' to another hash tree (e.g. to make sure that it
// hasn't changed)
func requireSame(t *testing.T, lTmp, rTmp HashTree) {
	l, r := lTmp.(*HashTreeProto), rTmp.(*HashTreeProto)
	// Make sure 'h' is still the same
	_, file, line, _ := runtime.Caller(1)
	require.True(t, proto.Equal(l, r),
		fmt.Sprintf("%s %s:%d\n%s %s\n%s  %s\n",
			"requireSame called at", file, line,
			"expected:\n", proto.MarshalTextString(l),
			"but got:\n", proto.MarshalTextString(r)))
}

// requireOperationInvariant makes sure that h isn't affected by calling 'op'.
// Good for checking that adding and deleting a file does nothing persistent,
// etc. This is separate from 'requireSame()' because often we want to test that
// an operation is invariant on several slightly different trees, and with this
// we only have to define 'op' once.
func requireOperationInvariant(t *testing.T, h OpenHashTree, op func()) {
	t.Helper()
	preop, err := h.Finish()
	require.NoError(t, err)
	// perform operation on 'h'
	op()
	// Make sure 'h' is still the same
	postop, err := h.Finish()
	require.NoError(t, err)
	preRoot, err := preop.Get("")
	require.NoError(t, err)
	postRoot, err := postop.Get("")
	require.NoError(t, err)
	require.True(t, bytes.Equal(preRoot.Hash, postRoot.Hash))
}

func newHashTree(tb testing.TB) OpenHashTree {
	result, err := NewDBHashTree()
	require.NoError(tb, err)
	return result
}

func TestPutFileBasic(t *testing.T) {
	// Put a file
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))
	hTmp := finish(t, h)
	require.Equal(t, int64(1), getT(t, hTmp, "/foo").SubtreeSize)
	require.Equal(t, int64(1), getT(t, hTmp, "").SubtreeSize)

	// Put a file under a directory and make sure changes are propagated upwards
	require.NoError(t, h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1))
	hTmp = finish(t, h)
	require.Equal(t, int64(1), getT(t, hTmp, "/dir/bar").SubtreeSize)
	require.Equal(t, int64(1), getT(t, hTmp, "/dir").SubtreeSize)
	require.Equal(t, int64(2), getT(t, hTmp, "").SubtreeSize)
	require.NoError(t, h.PutFile("/dir/buzz", obj(`hash:"8e02c"`), 1))

	// inspect h
	h1 := finish(t, h)
	require.Equal(t, int64(1), getT(t, h1, "/dir/buzz").SubtreeSize)
	require.Equal(t, int64(2), getT(t, h1, "/dir").SubtreeSize)
	require.Equal(t, int64(3), getT(t, h1, "").SubtreeSize)
	nodes, err := h1.List("/")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("foo", "dir"), node.Name)
	}

	nodes, err = h1.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("bar", "buzz"), node.Name)
	}
	require.Equal(t, int64(1), getT(t, h1, "/foo").SubtreeSize)

	// Make sure subsequent PutFile calls append
	require.NoError(t, h.PutFile("/foo", obj(`hash:"413e7"`), 1))
	h2 := finish(t, h)
	fmt.Printf("h1: %v\n", getT(t, h1, "/foo"))
	fmt.Printf("h2: %v\n", getT(t, h2, "/foo"))
	require.NotEqual(t, getT(t, h1, "/foo").Hash, getT(t, h2, "/foo").Hash)
	require.Equal(t, int64(2), getT(t, h2, "/foo").SubtreeSize)
}

func TestPutDirBasic(t *testing.T) {
	h := newHashTree(t)
	emptySha := sha256.Sum256([]byte{})

	// put a directory
	require.NoError(t, h.PutDir("/dir"))
	require.Equal(t, lenT(t, h), 2) // "/dir" and "/"
	require.Equal(t, []string(nil), getT(t, h, "/dir").DirNode.Children)
	h1 := finish(t, h)
	require.Equal(t, []string(nil), getT(t, h1, "/dir").DirNode.Children)
	require.Equal(t, emptySha[:], getT(t, h1, "/dir").Hash)
	require.Equal(t, lenT(t, h1), 2)

	// put a directory under another directory
	require.NoError(t, h.PutDir("/dir/foo"))
	require.NotEqual(t, []string{}, getT(t, h, "/dir").DirNode.Children)
	h2 := finish(t, h)
	require.NotEqual(t, []string{}, getT(t, h2, "/dir").DirNode.Children)
	nodes, err := h2.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	require.NotEqual(t, emptySha[:], getT(t, h2, "/dir").Hash)

	// delete the directory
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	h3 := finish(t, h)
	require.Equal(t, 0, len(getT(t, h3, "/dir").DirNode.Children))
	nodes, err = h3.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], getT(t, h3, "/dir").Hash)

	// Make sure that deleting a dir also deletes files under the dir
	require.NoError(t, h.PutFile("/dir/foo/bar", obj(`hash:"20c27"`), 1))
	require.NoError(t, h.DeleteFile("/dir/foo"))
	require.Equal(t, 0, len(getT(t, h, "/dir").DirNode.Children))
	require.Equal(t, lenT(t, h), 2)
	h4 := finish(t, h)
	require.NoError(t, err)
	require.Equal(t, 0, len(getT(t, h4, "/dir").DirNode.Children))
	nodes, err = h4.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], getT(t, h4, "/dir").Hash)
	require.Equal(t, lenT(t, h4), 2)
}

func TestPutError(t *testing.T) {
	h := newHashTree(t)
	require.NoError(t, h.PutFile("/foo", obj(`hash:"20c27"`), 1))

	// PutFile fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo/bar", obj(`hash:"8e02c"`), 1)
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
		node, err := h.GetOpen("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathNotFound, Code(err))
		require.Nil(t, node)
	})

	// PutDir fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutDir("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
		node, err := h.GetOpen("/foo/bar")
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
		h.DeleteFile("/dir/__NEW_FILE__")
		h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	}
	deleteAddDir := func() {
		h.DeleteFile("/dir/__NEW_DIR__")
		h.PutDir("/dir/__NEW_DIR__")
	}
	deleteAddSubFile := func() {
		h.DeleteFile("/dir/__NEW_DIR__")
		h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	}

	h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	requireOperationInvariant(t, h, deleteAddFile)
	h.PutDir("/dir/__NEW_DIR__")
	requireOperationInvariant(t, h, deleteAddDir)
	h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	requireOperationInvariant(t, h, deleteAddSubFile)

	// Make sure test still passes when trees are nonempty
	h = newHashTree(t)
	h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1)
	h.PutFile("/dir/bar", obj(`hash:"20c27"`), 1)

	h.PutFile("/dir/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	requireOperationInvariant(t, h, deleteAddFile)
	h.PutDir("/dir/__NEW_DIR__")
	requireOperationInvariant(t, h, deleteAddDir)
	h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", obj(`hash:"8e02c"`), 1)
	requireOperationInvariant(t, h, deleteAddSubFile)
}

// The hash of a directory doesn't change no matter what order files are added
// to it.
func TestPutFileCommutative(t *testing.T) {
	h := newHashTree(t)
	h2 := newHashTree(t)
	// Puts files into h in the order [A, B] and into h2 in the order [B, A]
	comparePutFiles := func() {
		h.PutFile("/dir/__NEW_FILE_A__", obj(`hash:"ebc57"`), 1)
		h.PutFile("/dir/__NEW_FILE_B__", obj(`hash:"20c27"`), 1)

		// Get state of both /dir and /, to make sure changes are preserved upwards
		// through the file hierarchy
		dirNodePtr, err := h.GetOpen("/dir")
		require.NoError(t, err)
		rootNodePtr, err := h.GetOpen("/")
		require.NoError(t, err)

		h2.PutFile("/dir/__NEW_FILE_B__", obj(`hash:"20c27"`), 1)
		h2.PutFile("/dir/__NEW_FILE_A__", obj(`hash:"ebc57"`), 1)

		dirNodePtr2, err := h2.GetOpen("/dir")
		require.NoError(t, err)
		rootNodePtr2, err := h2.GetOpen("/")
		require.NoError(t, err)
		require.Equal(t, *dirNodePtr, *dirNodePtr2)
		require.Equal(t, *rootNodePtr, *rootNodePtr2)
	}

	// (1) Run the test on empty trees
	comparePutFiles()
	// (2) Add some files & check that test still passes when trees are nonempty
	h, h2 = newHashTree(t), newHashTree(t)
	h.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1)
	h2.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1)
	h.PutFile("/dir/bar", obj(`hash:"9d432"`), 1)
	h2.PutFile("/dir/bar", obj(`hash:"9d432"`), 1)
	comparePutFiles()
}

// Given a directory D, renaming (removing and re-adding under a different name)
// a file or directory under D changes the hash of D, even if the contents are
// identical.
func TestRenameChangesHash(t *testing.T) {
	// Write a file, and then get the hash of every node from the file to the root
	h := newHashTree(t)
	h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1)

	h1 := finish(t, h)
	dirPre, err := h1.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h1.Get("/")
	require.NoError(t, err)

	// rename /dir/foo to /dir/bar, and make sure that changes the hash
	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1)

	h2 := finish(t, h)
	dirPost, err := h2.Get("/dir")
	require.NoError(t, err)
	rootPost, err := h2.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, dirPre.Hash, dirPost.Hash)
	require.NotEqual(t, rootPre.Hash, rootPost.Hash)
	require.Equal(t, dirPre.SubtreeSize, dirPost.SubtreeSize)
	require.Equal(t, rootPre.SubtreeSize, rootPost.SubtreeSize)

	// rename /dir to /dir2, and make sure that changes the hash
	h.DeleteFile("/dir")
	h.PutFile("/dir2/foo", obj(`hash:"ebc57"`), 1)

	h3 := finish(t, h)
	dirPost, err = h3.Get("/dir2")
	require.NoError(t, err)
	rootPost, err = h3.Get("/")
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
	h.PutFile("/dir/foo", obj(`hash:"ebc57"`), 1)

	h1 := finish(t, h)
	dirPre, err := h1.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h1.Get("/")
	require.NoError(t, err)

	// Change
	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/foo", obj(`hash:"8e02c"`), 1)

	h2 := finish(t, h)
	dirPost, err := h2.Get("/dir")
	require.NoError(t, err)
	rootPost, err := h2.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, dirPre.Hash, dirPost.Hash)
	require.NotEqual(t, rootPre.Hash, rootPost.Hash)
	require.Equal(t, dirPre.SubtreeSize, dirPost.SubtreeSize)
	require.Equal(t, rootPre.SubtreeSize, rootPost.SubtreeSize)
}

func TestIsGlob(t *testing.T) {
	require.True(t, isGlob(`*`))
	require.True(t, isGlob(`path/to*/file`))
	require.True(t, isGlob(`path/**/file`))
	require.True(t, isGlob(`path/to/f?le`))
	require.True(t, isGlob(`pa!h/to/file`))
	require.True(t, isGlob(`pa[th]/to/file`))
	require.True(t, isGlob(`pa{th}/to/file`))
	require.True(t, isGlob(`*/*`))
	require.False(t, isGlob(`path`))
	require.False(t, isGlob(`path/to/file1.txt`))
	require.False(t, isGlob(`path/to_test-a/file.txt`))
}

func TestGlobFile(t *testing.T) {
	hTmp := newHashTree(t)
	hTmp.PutFile("/foo", obj(`hash:"20c27"`), 1)
	hTmp.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1)
	hTmp.PutFile("/dir/buzz", obj(`hash:"8e02c"`), 1)
	h, err := hTmp.Finish()
	require.NoError(t, err)

	// patterns that match the whole repo ("/")
	for _, pattern := range []string{"", "/"} {
		paths, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 1, len(paths))
		for path := range paths {
			require.EqualOneOf(t, i(""), path)
		}
	}

	// patterns that match top-level dirs/files
	for _, pattern := range []string{"*", "/*"} {
		paths, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 2, len(paths))
		for path := range paths {
			require.EqualOneOf(t, i("/foo", "/dir"), path)
		}
	}

	// patterns that match second-level dirs/files
	for _, pattern := range []string{"dir/*", "/dir/*", "*/*", "/*/*"} {
		paths, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 2, len(paths))
		for path := range paths {
			require.EqualOneOf(t, i("/dir/bar", "/dir/buzz"), path)
		}
	}
}

func TestMerge(t *testing.T) {
	lTmp, rTmp := newHashTree(t), newHashTree(t)
	lTmp.PutFile("/foo-left", obj(`hash:"20c27"`), 1)
	lTmp.PutFile("/dir-left/bar-left", obj(`hash:"ebc57"`), 1)
	lTmp.PutFile("/dir-shared/buzz-left", obj(`hash:"8e02c"`), 1)
	lTmp.PutFile("/dir-shared/file-shared", obj(`hash:"9d432"`), 1)
	rTmp.PutFile("/foo-right", obj(`hash:"20c27"`), 1)
	rTmp.PutFile("/dir-right/bar-right", obj(`hash:"ebc57"`), 1)
	rTmp.PutFile("/dir-shared/buzz-right", obj(`hash:"8e02c"`), 1)
	rTmp.PutFile("/dir-shared/file-shared", obj(`hash:"9d432"`), 1)
	l, r := finish(t, lTmp), finish(t, rTmp)

	expectedTmp := newHashTree(t)
	expectedTmp.PutFile("/foo-left", obj(`hash:"20c27"`), 1)
	expectedTmp.PutFile("/dir-left/bar-left", obj(`hash:"ebc57"`), 1)
	expectedTmp.PutFile("/dir-shared/buzz-left", obj(`hash:"8e02c"`), 1)
	expectedTmp.PutFile("/dir-shared/file-shared", obj(`hash:"9d432"`), 1)
	expectedTmp.PutFile("/foo-right", obj(`hash:"20c27"`), 1)
	expectedTmp.PutFile("/dir-right/bar-right", obj(`hash:"ebc57"`), 1)
	expectedTmp.PutFile("/dir-shared/buzz-right", obj(`hash:"8e02c"`), 1)
	expectedTmp.PutFile("/dir-shared/file-shared", obj(`hash:"9d432"`), 1)
	expected, err := expectedTmp.Finish()
	require.NoError(t, err)

	h := l.Open()
	err = h.Merge(r)
	require.NoError(t, err)
	requireSame(t, expected, finish(t, h))

	h = r.Open()
	err = h.Merge(l)
	require.NoError(t, err)
	requireSame(t, expected, finish(t, h))

	h = newHashTree(t)
	err = h.Merge(l, r)
	require.NoError(t, err)
	requireSame(t, expected, finish(t, h))
}

// Test that Merge() works with empty hash trees
func TestMergeEmpty(t *testing.T) {
	expectedTmp := newHashTree(t)
	expectedTmp.PutFile("/foo", obj(`hash:"20c27"`), 1)
	expectedTmp.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1)
	expected, err := expectedTmp.Finish()
	require.NoError(t, err)

	// Merge empty tree into full tree
	l := expected.Open()
	r := newHashTree(t)
	require.NoError(t, l.Merge(finish(t, r)))
	requireSame(t, expected, finish(t, l))

	// Merge full tree into empty tree
	require.NoError(t, r.Merge(finish(t, l)))
	requireSame(t, expected, finish(t, r))
}

// Test that Walk() works
func TestWalk(t *testing.T) {
	tmp := newHashTree(t)
	tmp.PutFile("/foo", obj(`hash:"20c27"`), 1)
	tmp.PutFile("/dir/bar", obj(`hash:"ebc57"`), 1)
	tree, err := tmp.Finish()
	require.NoError(t, err)

	expectedPaths := map[string]bool{
		"/":        true,
		"/foo":     true,
		"/dir":     true,
		"/dir/bar": true,
	}
	require.NoError(t, tree.Walk("/", func(path string, node *NodeProto) error {
		require.True(t, expectedPaths[path])
		delete(expectedPaths, path)
		return nil
	}))
	require.Equal(t, 0, len(expectedPaths))
}

// Test that HashTree methods return the right error codes
func TestErrorCode(t *testing.T) {
	require.Equal(t, OK, Code(nil))
	require.Equal(t, Unknown, Code(fmt.Errorf("external error")))

	h := newHashTree(t)
	hdone := finish(t, newHashTree(t))
	_, err := hdone.Get("/path")
	require.Equal(t, PathNotFound, Code(err))

	h.PutFile("/foo", obj(`hash:"20c27"`), 1)
	err = h.PutFile("/foo/bar", obj(`hash:"9d432"`), 1)
	require.Equal(t, PathConflict, Code(err))
	h.PutFile("/bar/foo", obj(`hash:"9d432"`), 1)
	err = h.PutFile("/bar", obj(`hash:"20c27"`), 1)
	require.Equal(t, PathConflict, Code(err))
}

func TestSerialize(t *testing.T) {
	hTmp := newHashTree(t)
	require.NoError(t, hTmp.PutFile("/foo", obj(`hash:"20c27"`), 1))
	require.NoError(t, hTmp.PutFile("/bar/buzz", obj(`hash:"9d432"`), 1))
	h := finish(t, hTmp)

	// Serialize and Deserialize 'h'
	bts, err := Serialize(h)
	require.NoError(t, err)
	h2, err := Deserialize(bts)
	require.NoError(t, err)
	requireSame(t, h, h2)

	// Modify 'h', and Serialize and Deserialize it again
	require.NoError(t, hTmp.PutFile("/bar/buzz2", obj(`hash:"8e02c"`), 1))
	h = finish(t, hTmp)
	bts, err = Serialize(h)
	require.NoError(t, err)
	h3, err := Deserialize(bts)
	require.NoError(t, err)
	requireSame(t, h, h3)

	// Make sure 'h2' does not equal 'h' or 'h3'
	// require.False(t, proto.Equal(h, h2.(*HashTreeProto)))
	// require.False(t, proto.Equal(h2.(*HashTreeProto), h3.(*HashTreeProto)))
}

func TestSerializeError(t *testing.T) {
	// Test version
	h := &HashTreeProto{Version: -1}
	bts, err := h.Marshal()
	require.NoError(t, err)
	_, err = Deserialize(bts)
	require.YesError(t, err)
	require.Equal(t, Unsupported, Code(err))
}

func TestListEmpty(t *testing.T) {
	tree := newHashTree(t)
	_, err := tree.List("/")
	require.NoError(t, err)
	_, err = tree.Glob("*")
	require.NoError(t, err)
	_, err = tree.Glob("/*")
	require.NoError(t, err)
}
