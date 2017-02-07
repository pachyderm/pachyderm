package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"runtime"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// br parses a string as a BlockRef
func br(s ...string) []*pfs.BlockRef {
	result := make([]*pfs.BlockRef, len(s))
	for i, ss := range s {
		result[i] = &pfs.BlockRef{}
		proto.UnmarshalText(ss, result[i])
		if result[i].Range == nil {
			result[i].Range = &pfs.ByteRange{
				Lower: 0,
				Upper: 1, // Makes sure test files have non-zero size
			}
		}
	}
	return result
}

// Convenience function to convert a list of strings to []interface{} for
// EqualOneOf
func i(ss ...string) []interface{} {
	result := make([]interface{}, len(ss))
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

func finish(t *testing.T, h OpenHashTree) *HashTreeProto {
	h2, err := h.Finish()
	require.NoError(t, err)
	return h2.(*HashTreeProto)
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
	preop, err := h.(*hashtree).clone()
	if err != nil {
		t.Fatalf("could not clone 'h' in requireOperationInvariant: %s", err)
	}
	// perform operation on 'h'
	op()
	// Make sure 'h' is still the same
	_, file, line, _ := runtime.Caller(1)
	require.True(t, equals(preop, h),
		fmt.Sprintf("%s %s:%d\n%s  %s\n%s %s\n",
			"requireOperationInvariant called at", file, line,
			"pre-op HashTree:\n", tostring(preop),
			"post-op HashTree:\n", tostring(h)))
}

func TestPutFileBasic(t *testing.T) {
	// Put a file
	h := NewHashTree().(*hashtree)
	h.PutFile("/foo", br(`block{hash:"20c27"}`))
	require.Equal(t, int64(1), h.fs["/foo"].SubtreeSize)
	require.Equal(t, int64(1), h.fs[""].SubtreeSize)

	// Put a file under a directory and make sure changes are propagated upwards
	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	require.Equal(t, int64(1), h.fs["/dir/bar"].SubtreeSize)
	require.Equal(t, int64(1), h.fs["/dir"].SubtreeSize)
	require.Equal(t, int64(2), h.fs[""].SubtreeSize)
	h.PutFile("/dir/buzz", br(`block{hash:"8e02c"}`))
	require.Equal(t, int64(1), h.fs["/dir/buzz"].SubtreeSize)
	require.Equal(t, int64(2), h.fs["/dir"].SubtreeSize)
	require.Equal(t, int64(3), h.fs[""].SubtreeSize)

	// inspect h
	h1 := finish(t, h)
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
	require.Equal(t, int64(1), h1.Fs["/foo"].SubtreeSize)

	// Make sure subsequent PutFile calls append
	h.PutFile("/foo", br(`block{hash:"413e7"}`))
	h2 := finish(t, h)
	require.NotEqual(t, h1.Fs["/foo"].Hash, h2.Fs["/foo"].Hash)
	require.Equal(t, int64(2), h2.Fs["/foo"].SubtreeSize)
}

func TestPutDirBasic(t *testing.T) {
	h := NewHashTree().(*hashtree)
	emptySha := sha256.Sum256([]byte{})

	// put a directory
	h.PutDir("/dir")
	require.Equal(t, len(h.fs), 2) // "/dir" and "/"
	require.Equal(t, []string(nil), h.fs["/dir"].DirNode.Children)
	h1 := finish(t, h)
	require.Equal(t, []string(nil), h1.Fs["/dir"].DirNode.Children)
	require.Equal(t, emptySha[:], h1.Fs["/dir"].Hash)
	require.Equal(t, len(h1.Fs), 2)

	// put a directory under another directory
	h.PutDir("/dir/foo")
	require.NotEqual(t, []string{}, h.fs["/dir"].DirNode.Children)
	h2 := finish(t, h)
	require.NotEqual(t, []string{}, h2.Fs["/dir"].DirNode.Children)
	nodes, err := h2.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	require.NotEqual(t, emptySha[:], h2.Fs["/dir"].Hash)

	// delete the directory
	h.DeleteFile("/dir/foo")
	require.Equal(t, []string{}, h.fs["/dir"].DirNode.Children)
	h3 := finish(t, h)
	require.Equal(t, []string{}, h3.Fs["/dir"].DirNode.Children)
	nodes, err = h3.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], h3.Fs["/dir"].Hash)

	// Make sure that deleting a dir also deletes files under the dir
	h.PutFile("/dir/foo/bar", br(`block{hash:"20c27"}`))
	h.DeleteFile("/dir/foo")
	require.Equal(t, []string{}, h.fs["/dir"].DirNode.Children)
	require.Equal(t, len(h.fs), 2)
	h4 := finish(t, h)
	require.NoError(t, err)
	require.Equal(t, []string{}, h4.Fs["/dir"].DirNode.Children)
	nodes, err = h4.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], h4.Fs["/dir"].Hash)
	require.Equal(t, len(h4.Fs), 2)
}

func TestPutError(t *testing.T) {
	h := NewHashTree()
	err := h.PutFile("/foo", br(`block{hash:"20c27"}`))
	require.NoError(t, err)

	// PutFile fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo/bar", br(`block{hash:"8e02c"}`))
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

	// Merge fails if src and dest disagree about whether a node is a file or
	// directory, and h is unchanged
	srcOpen := NewHashTree()
	srcOpen.PutFile("/buzz", br(`block{hash:"9d432"}`))
	srcOpen.PutFile("/foo/bar", br(`block{hash:"ebc57"}`))
	src, err := srcOpen.Finish()
	require.NoError(t, err)
	requireOperationInvariant(t, h, func() {
		err := h.Merge([]HashTree{src})
		require.YesError(t, err, tostring(h))
		require.Equal(t, PathConflict, Code(err))
	})

	// PutFile fails if a directory already exists (put /foo when /foo/bar exists)
	err = h.DeleteFile("/foo")
	require.NoError(t, err)
	err = h.PutFile("/foo/bar", br(`block{hash:"ebc57"}`))
	require.NoError(t, err)
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo", br(`block{hash:"8e02c"}`))
		require.YesError(t, err)
		require.Equal(t, PathConflict, Code(err))
	})
}

func TestDeleteDirError(t *testing.T) {
	// Put root dir
	h := NewHashTree().(*hashtree)
	h.PutDir("/")
	require.Equal(t, 1, len(h.fs))

	err := h.DeleteFile("/does/not/exist")
	require.YesError(t, err)
	require.Equal(t, PathNotFound, Code(err))
	require.Equal(t, 1, len(h.fs))
}

// Given a directory D, test that adding and then deleting a file/directory to
// D does not change D.
func TestAddDeleteReverts(t *testing.T) {
	h := NewHashTree()
	addDeleteFile := func() {
		h.PutFile("/dir/__NEW_FILE__", br(`block{hash:"8e02c"}`))
		h.DeleteFile("/dir/__NEW_FILE__")
	}
	addDeleteDir := func() {
		h.PutDir("/dir/__NEW_DIR__")
		h.DeleteFile("/dir/__NEW_DIR__")
	}
	addDeleteSubFile := func() {
		h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", br(`block{hash:"8e02c"}`))
		h.DeleteFile("/dir/__NEW_DIR__")
	}

	h.PutDir("/dir")
	requireOperationInvariant(t, h, addDeleteFile)
	requireOperationInvariant(t, h, addDeleteDir)
	requireOperationInvariant(t, h, addDeleteSubFile)
	// Add some files to make sure the test still passes when D already has files
	// in it.
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))
	h.PutFile("/dir/bar", br(`block{hash:"20c27"}`))
	requireOperationInvariant(t, h, addDeleteFile)
	requireOperationInvariant(t, h, addDeleteDir)
	requireOperationInvariant(t, h, addDeleteSubFile)
}

// Given a directory D, test that deleting and then adding a file/directory to
// D does not change D.
func TestDeleteAddReverts(t *testing.T) {
	h := NewHashTree()
	deleteAddFile := func() {
		h.DeleteFile("/dir/__NEW_FILE__")
		h.PutFile("/dir/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	}
	deleteAddDir := func() {
		h.DeleteFile("/dir/__NEW_DIR__")
		h.PutDir("/dir/__NEW_DIR__")
	}
	deleteAddSubFile := func() {
		h.DeleteFile("/dir/__NEW_DIR__")
		h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	}

	h.PutFile("/dir/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	requireOperationInvariant(t, h, deleteAddFile)
	h.PutDir("/dir/__NEW_DIR__")
	requireOperationInvariant(t, h, deleteAddDir)
	h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	requireOperationInvariant(t, h, deleteAddSubFile)

	// Make sure test still passes when trees are nonempty
	h = NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))
	h.PutFile("/dir/bar", br(`block{hash:"20c27"}`))

	h.PutFile("/dir/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	requireOperationInvariant(t, h, deleteAddFile)
	h.PutDir("/dir/__NEW_DIR__")
	requireOperationInvariant(t, h, deleteAddDir)
	h.PutFile("/dir/__NEW_DIR__/__NEW_FILE__", br(`block{hash:"8e02c"}`))
	requireOperationInvariant(t, h, deleteAddSubFile)
}

// The hash of a directory doesn't change no matter what order files are added
// to it.
func TestPutFileCommutative(t *testing.T) {
	h := NewHashTree()
	h2 := NewHashTree()
	// Puts files into h in the order [A, B] and into h2 in the order [B, A]
	comparePutFiles := func() {
		h.PutFile("/dir/__NEW_FILE_A__", br(`block{hash:"ebc57"}`))
		h.PutFile("/dir/__NEW_FILE_B__", br(`block{hash:"20c27"}`))

		// Get state of both /dir and /, to make sure changes are preserved upwards
		// through the file hierarchy
		dirNodePtr, err := h.GetOpen("/dir")
		require.NoError(t, err)
		rootNodePtr, err := h.GetOpen("/")
		require.NoError(t, err)

		h2.PutFile("/dir/__NEW_FILE_B__", br(`block{hash:"20c27"}`))
		h2.PutFile("/dir/__NEW_FILE_A__", br(`block{hash:"ebc57"}`))

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
	h, h2 = NewHashTree(), NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"8e02c"}`))
	h2.PutFile("/dir/foo", br(`block{hash:"8e02c"}`))
	h.PutFile("/dir/bar", br(`block{hash:"9d432"}`))
	h2.PutFile("/dir/bar", br(`block{hash:"9d432"}`))
	comparePutFiles()
}

// Given a directory D, renaming (removing and re-adding under a different name)
// a file or directory under D changes the hash of D, even if the contents are
// identical.
func TestRenameChangesHash(t *testing.T) {
	// Write a file, and then get the hash of every node from the file to the root
	h := NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))

	h1 := finish(t, h)
	dirPre, err := h1.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h1.Get("/")
	require.NoError(t, err)

	// rename /dir/foo to /dir/bar, and make sure that changes the hash
	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))

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
	h.PutFile("/dir2/foo", br(`block{hash:"ebc57"}`))

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
	h := NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))

	h1 := finish(t, h)
	dirPre, err := h1.Get("/dir")
	require.NoError(t, err)
	rootPre, err := h1.Get("/")
	require.NoError(t, err)

	// Change
	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/foo", br(`block{hash:"8e02c"}`))

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

func TestGlobFile(t *testing.T) {
	hTmp := NewHashTree()
	hTmp.PutFile("/foo", br(`block{hash:"20c27"}`))
	hTmp.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	hTmp.PutFile("/dir/buzz", br(`block{hash:"8e02c"}`))
	h, err := hTmp.Finish()
	require.NoError(t, err)

	// Patterns that match the whole repo ("/")
	for _, pattern := range []string{"", "/"} {
		nodes, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 1, len(nodes))
		for _, node := range nodes {
			require.EqualOneOf(t, i(""), node.Name)
		}
	}

	// patterns that match top-level dirs/files
	for _, pattern := range []string{"*", "/*"} {
		nodes, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 2, len(nodes))
		for _, node := range nodes {
			require.EqualOneOf(t, i("foo", "dir"), node.Name)
		}
	}

	// Patterns that match second-level dirs/files
	for _, pattern := range []string{"dir/*", "/dir/*", "*/*", "/*/*"} {
		nodes, err := h.Glob(pattern)
		require.NoError(t, err)
		require.Equal(t, 2, len(nodes))
		for _, node := range nodes {
			require.EqualOneOf(t, i("bar", "buzz"), node.Name)
		}
	}
}

func TestMerge(t *testing.T) {
	lTmp, rTmp := NewHashTree(), NewHashTree()
	lTmp.PutFile("/foo-left", br(`block{hash:"20c27"}`))
	lTmp.PutFile("/dir-left/bar-left", br(`block{hash:"ebc57"}`))
	lTmp.PutFile("/dir-shared/buzz-left", br(`block{hash:"8e02c"}`))
	lTmp.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	rTmp.PutFile("/foo-right", br(`block{hash:"20c27"}`))
	rTmp.PutFile("/dir-right/bar-right", br(`block{hash:"ebc57"}`))
	rTmp.PutFile("/dir-shared/buzz-right", br(`block{hash:"8e02c"}`))
	rTmp.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	l, r := finish(t, lTmp), finish(t, rTmp)

	expectedTmp := NewHashTree()
	expectedTmp.PutFile("/foo-left", br(`block{hash:"20c27"}`))
	expectedTmp.PutFile("/dir-left/bar-left", br(`block{hash:"ebc57"}`))
	expectedTmp.PutFile("/dir-shared/buzz-left", br(`block{hash:"8e02c"}`))
	expectedTmp.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	expectedTmp.PutFile("/foo-right", br(`block{hash:"20c27"}`))
	expectedTmp.PutFile("/dir-right/bar-right", br(`block{hash:"ebc57"}`))
	expectedTmp.PutFile("/dir-shared/buzz-right", br(`block{hash:"8e02c"}`))
	expectedTmp.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	expected, err := expectedTmp.Finish()
	require.NoError(t, err)

	h := l.Open()
	h.Merge([]HashTree{r})
	requireSame(t, expected, finish(t, h))

	h = r.Open()
	h.Merge([]HashTree{l})
	requireSame(t, expected, finish(t, h))

	h = NewHashTree()
	h.Merge([]HashTree{l, r})
	requireSame(t, expected, finish(t, h))
}

// Test that Merge() works with empty hash trees
func TestMergeEmpty(t *testing.T) {
	expectedTmp := NewHashTree()
	expectedTmp.PutFile("/foo", br(`block{hash:"20c27"}`))
	expectedTmp.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	expected, err := expectedTmp.Finish()
	require.NoError(t, err)

	// Merge empty tree into full tree
	l := expected.Open()
	r := NewHashTree()
	l.Merge([]HashTree{finish(t, r)})
	requireSame(t, expected, finish(t, l))

	// Merge full tree into empty tree
	r.Merge([]HashTree{finish(t, l)})
	requireSame(t, expected, finish(t, r))
}

// Test that HashTree methods return the right error codes
func TestErrorCode(t *testing.T) {
	require.Equal(t, OK, Code(nil))
	require.Equal(t, Unknown, Code(fmt.Errorf("external error")))

	h := NewHashTree()
	hdone := finish(t, NewHashTree())
	_, err := hdone.Get("/path")
	require.Equal(t, PathNotFound, Code(err))

	h.PutFile("/foo", br(`block{hash:"20c27"}`))
	err = h.PutFile("/foo/bar", br(`block{hash:"9d432"}`))
	require.Equal(t, PathConflict, Code(err))
	h.PutFile("/bar/foo", br(`block{hash:"9d432"}`))
	err = h.PutFile("/bar", br(`block{hash:"20c27"}`))
	require.Equal(t, PathConflict, Code(err))

	_, err = finish(t, h).Glob("/*\\")
	require.Equal(t, MalformedGlob, Code(err))
}

func TestSerialize(t *testing.T) {
	hTmp := NewHashTree()
	require.NoError(t, hTmp.PutFile("/foo", br(`block{hash:"20c27"}`)))
	require.NoError(t, hTmp.PutFile("/bar/buzz", br(`block{hash:"9d432"}`)))
	h := finish(t, hTmp)

	// Serialize and Deserialize 'h'
	bts, err := Serialize(h)
	require.NoError(t, err)
	h2, err := Deserialize(bts)
	require.NoError(t, err)
	requireSame(t, h, h2)

	// Modify 'h', and Serialize and Deserialize it again
	require.NoError(t, hTmp.PutFile("/bar/buzz2", br(`block{hash:"8e02c"}`)))
	h = finish(t, hTmp)
	bts, err = Serialize(h)
	require.NoError(t, err)
	h3, err := Deserialize(bts)
	require.NoError(t, err)
	requireSame(t, h, h3)

	// Make sure 'h2' does not equal 'h' or 'h3'
	require.False(t, proto.Equal(h, h2.(*HashTreeProto)))
	require.False(t, proto.Equal(h2.(*HashTreeProto), h3.(*HashTreeProto)))
}

func TestSerializeError(t *testing.T) {
	// Test version
	bts, err := proto.Marshal(&HashTreeProto{Version: -1})
	require.NoError(t, err)
	_, err = Deserialize(bts)
	require.YesError(t, err)
	require.Equal(t, Unsupported, Code(err))
}
