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

func clone(h HashTree) HashTree {
	bb, err := h.Marshal()
	if err != nil {
		panic("could not clone HashTree: " + err.Error())
	}
	h2, err := Unmarshal(bb)
	if err != nil {
		panic("could not clone HashTree: " + err.Error())
	}
	return h2
}

func tostring(htmp HashTree) string {
	h := htmp.(*hashtree)
	bufsize := len(h.fs) * 25
	buf := bytes.NewBuffer(make([]byte, 0, bufsize))
	for k, v := range h.fs {
		buf.WriteString(fmt.Sprintf("\"%s\": %+v\n", k, v))
	}
	return buf.String()
}

func equals(ltmp, rtmp HashTree) bool {
	l, r := ltmp.(*hashtree), rtmp.(*hashtree)
	if len(l.fs) != len(r.fs) {
		return false
	}
	for k, v := range l.fs {
		if ov, ok := r.fs[k]; !ok || !proto.Equal(v, ov) {
			return false
		}
	}
	return true
}

// requireSame compares 'h' to another hash tree (e.g. to make sure that it
// hasn't changed)
func requireSame(t *testing.T, ltmp, rtmp HashTree) {
	l, r := ltmp.(*hashtree), rtmp.(*hashtree)
	// Make sure 'h' is still the same
	_, file, line, _ := runtime.Caller(1)
	require.True(t, equals(l, r),
		fmt.Sprintf("%s %s:%d\n%s %s\n%s  %s\n",
			"requireSame called at", file, line,
			"expected:\n", tostring(l),
			"but got:\n", tostring(r)))
}

// requireOperationInvariant makes sure that h isn't affected by calling 'op'.
// Good for checking that adding and deleting a file does nothing persistent,
// etc. This is separate from 'requireSame()' because often we want to test that
// an operation is invariant on several slightly different trees, and with this
// we only have to define 'op' once.
func requireOperationInvariant(t *testing.T, h HashTree, op func()) {
	preop := clone(h)
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
	nodes, err := h.List("/")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("foo", "dir"), node.Name)
	}

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))
	for _, node := range nodes {
		require.EqualOneOf(t, i("bar", "buzz"), node.Name)
	}

	// Make sure subsequent PutFile calls append
	oldSha := make([]byte, len(h.fs["/foo"].Hash))
	copy(oldSha, h.fs["/foo"].Hash)
	require.Equal(t, int64(1), h.fs["/foo"].SubtreeSize)

	h.PutFile("/foo", br(`block{hash:"413e7"}`))
	require.NotEqual(t, oldSha, h.fs["/foo"].Hash)
	require.Equal(t, int64(2), h.fs["/foo"].SubtreeSize)
}

func TestPutDirBasic(t *testing.T) {
	h := NewHashTree().(*hashtree)
	emptySha := sha256.Sum256([]byte{})

	// put a directory
	h.PutDir("/dir")
	require.Equal(t, emptySha[:], h.fs["/dir"].Hash)
	require.Equal(t, []string(nil), h.fs["/dir"].DirNode.Children)
	require.Equal(t, len(h.fs), 2)

	// put a directory under another directory
	h.PutDir("/dir/foo")
	nodes, err := h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
	require.NotEqual(t, emptySha[:], h.fs["/dir"].Hash)
	require.NotEqual(t, []string{}, h.fs["/dir"].DirNode.Children)

	// delete the directory
	h.DeleteFile("/dir/foo")
	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], h.fs["/dir"].Hash)
	require.Equal(t, []string{}, h.fs["/dir"].DirNode.Children)

	// Make sure that deleting a dir also deletes files under the dir
	h.PutFile("/dir/foo/bar", br(`block{hash:"20c27"}`))
	h.DeleteFile("/dir/foo")
	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))
	require.Equal(t, emptySha[:], h.fs["/dir"].Hash)
	require.Equal(t, []string{}, h.fs["/dir"].DirNode.Children)
	require.Equal(t, len(h.fs), 2)
}

func TestPutError(t *testing.T) {
	h := NewHashTree()
	err := h.PutFile("/foo", br(`block{hash:"20c27"}`))
	require.NoError(t, err)

	// PutFile fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutFile("/foo/bar", br(`block{hash:"8e02c"}`))
		require.YesError(t, err)
		node, err := h.Get("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathNotFound, Code(err))
		require.Nil(t, node)
	})

	// PutDir fails if the parent is a file, and h is unchanged
	requireOperationInvariant(t, h, func() {
		err := h.PutDir("/foo/bar")
		require.YesError(t, err)
		node, err := h.Get("/foo/bar")
		require.YesError(t, err)
		require.Equal(t, PathNotFound, Code(err))
		require.Nil(t, node)
	})

	// Merge fails if src and dest disagree about whether a node is a file or
	// directory, and h is unchanged
	src := NewHashTree()
	src.PutFile("/buzz", br(`block{hash:"9d432"}`))
	src.PutFile("/foo/bar", br(`block{hash:"ebc57"}`))
	requireOperationInvariant(t, h, func() {
		err := h.Merge([]HashTree{src})
		require.YesError(t, err, tostring(h))
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

	// Add some files to make sure the test still passes when D already has files
	// in it.
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
	comparePutFiles := func() {
		h.PutFile("/dir/__NEW_FILE_A__", br(`block{hash:"ebc57"}`))
		h.PutFile("/dir/__NEW_FILE_B__", br(`block{hash:"20c27"}`))

		// Get state of both /dir and /, to make sure changes are preserved upwards
		// through the file hierarchy
		dirNodePtr, err := h.Get("/dir")
		require.NoError(t, err)
		rootNodePtr, err := h.Get("/")
		require.NoError(t, err)

		h2.PutFile("/dir/__NEW_FILE_B__", br(`block{hash:"20c27"}`))
		h2.PutFile("/dir/__NEW_FILE_A__", br(`block{hash:"ebc57"}`))

		dirNodePtr2, err := h2.Get("/dir")
		require.NoError(t, err)
		rootNodePtr2, err := h2.Get("/")
		require.NoError(t, err)
		require.Equal(t, *dirNodePtr, *dirNodePtr2)
		require.Equal(t, *rootNodePtr, *rootNodePtr2)

		// Revert 'h' before the next call to deleteAddInspect()
		h.DeleteFile("/dir/__nEw_FiLe__")
	}

	comparePutFiles()
	// Add some files to make sure the test still passes when D already has files
	// in it.
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
	h := NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))

	dirPtr, err := h.Get("/dir")
	require.NoError(t, err)
	rootPtr, err := h.Get("/")
	require.NoError(t, err)
	dirPre, rootPre := proto.Clone(dirPtr).(*NodeProto), proto.Clone(rootPtr).(*NodeProto)

	// rename /dir/foo to /dir/bar
	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))

	dirPtr, err = h.Get("/dir")
	require.NoError(t, err)
	rootPtr, err = h.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, (*dirPre).Hash, (*dirPtr).Hash)
	require.NotEqual(t, (*rootPre).Hash, (*rootPtr).Hash)
	require.Equal(t, (*dirPre).SubtreeSize, (*dirPtr).SubtreeSize)
	require.Equal(t, (*rootPre).SubtreeSize, (*rootPtr).SubtreeSize)

	// rename /dir to /dir2
	h.DeleteFile("/dir")
	h.PutFile("/dir2/foo", br(`block{hash:"ebc57"}`))

	dirPtr, err = h.Get("/dir2")
	require.NoError(t, err)
	rootPtr, err = h.Get("/")
	require.NoError(t, err)

	require.Equal(t, dirPre.Hash, (*dirPtr).Hash) // dir == dir2
	require.NotEqual(t, rootPre.Hash, (*rootPtr).Hash)
	require.Equal(t, (*dirPre).SubtreeSize, (*dirPtr).SubtreeSize)
	require.Equal(t, (*rootPre).SubtreeSize, (*rootPtr).SubtreeSize)
}

// Given a directory D, rewriting (removing and re-adding a different file
// under the same name) a file or directory under D changes the hash of D, even
// if the contents are identical.
func TestRewriteChangesHash(t *testing.T) {
	h := NewHashTree()
	h.PutFile("/dir/foo", br(`block{hash:"ebc57"}`))

	dirPtr, err := h.Get("/dir")
	require.NoError(t, err)
	rootPtr, err := h.Get("/")
	require.NoError(t, err)
	dirPre, rootPre := proto.Clone(dirPtr).(*NodeProto), proto.Clone(rootPtr).(*NodeProto)

	h.DeleteFile("/dir/foo")
	h.PutFile("/dir/foo", br(`block{hash:"8e02c"}`))

	dirPtr, err = h.Get("/dir")
	require.NoError(t, err)
	rootPtr, err = h.Get("/")
	require.NoError(t, err)

	require.NotEqual(t, dirPre.Hash, (*dirPtr).Hash)
	require.NotEqual(t, rootPre.Hash, (*rootPtr).Hash)
	require.Equal(t, (*dirPre).SubtreeSize, (*dirPtr).SubtreeSize)
	require.Equal(t, (*rootPre).SubtreeSize, (*rootPtr).SubtreeSize)
}

func TestGlobFile(t *testing.T) {
	h := NewHashTree()
	h.PutFile("/foo", br(`block{hash:"20c27"}`))
	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	h.PutFile("/dir/buzz", br(`block{hash:"8e02c"}`))

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
	l, r := NewHashTree(), NewHashTree()
	l.PutFile("/foo-left", br(`block{hash:"20c27"}`))
	l.PutFile("/dir-left/bar-left", br(`block{hash:"ebc57"}`))
	l.PutFile("/dir-shared/buzz-left", br(`block{hash:"8e02c"}`))
	l.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	r.PutFile("/foo-right", br(`block{hash:"20c27"}`))
	r.PutFile("/dir-right/bar-right", br(`block{hash:"ebc57"}`))
	r.PutFile("/dir-shared/buzz-right", br(`block{hash:"8e02c"}`))
	r.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))

	expected := NewHashTree()
	expected.PutFile("/foo-left", br(`block{hash:"20c27"}`))
	expected.PutFile("/dir-left/bar-left", br(`block{hash:"ebc57"}`))
	expected.PutFile("/dir-shared/buzz-left", br(`block{hash:"8e02c"}`))
	expected.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))
	expected.PutFile("/foo-right", br(`block{hash:"20c27"}`))
	expected.PutFile("/dir-right/bar-right", br(`block{hash:"ebc57"}`))
	expected.PutFile("/dir-shared/buzz-right", br(`block{hash:"8e02c"}`))
	expected.PutFile("/dir-shared/file-shared", br(`block{hash:"9d432"}`))

	h := clone(l)
	h.Merge([]HashTree{r})
	requireSame(t, expected, h)

	h = clone(r)
	h.Merge([]HashTree{l})
	requireSame(t, expected, h)

	h = NewHashTree()
	h.Merge([]HashTree{l, r})
	requireSame(t, expected, h)
}

// Test that Merge() works with empty hash trees
func TestMergeEmpty(t *testing.T) {
	l, r, expected := NewHashTree(), NewHashTree(), NewHashTree()
	expected.PutFile("/foo", br(`block{hash:"20c27"}`))
	expected.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))

	b, _ := expected.Marshal()

	// Merge empty tree into full tree
	l, err := Unmarshal(b)
	require.NoError(t, err)
	l.Merge([]HashTree{r})
	requireSame(t, expected, l)

	// Merge full tree into empty tree
	r.Merge([]HashTree{l})
	requireSame(t, expected, r)
}

// Test that HashTree methods return the right error codes
func TestErrorCode(t *testing.T) {
	require.Equal(t, OK, Code(nil))
	require.Equal(t, Unknown, Code(fmt.Errorf("external error")))

	h := NewHashTree()
	_, err := h.Get("/path")
	require.Equal(t, PathNotFound, Code(err))

	h.PutFile("/foo", br(`block{hash:"20c27"}`))
	err = h.PutFile("/foo/bar", br(`block{hash:"9d432"}`))
	require.Equal(t, PathConflict, Code(err))

	_, err = h.Glob("/*\\")
	require.Equal(t, MalformedGlob, Code(err))
}

func TestMarshal(t *testing.T) {
	h := NewHashTree()
	require.NoError(t, h.PutFile("/foo", br(`block{hash:"20c27"}`)))
	require.NoError(t, h.PutFile("/bar/buzz", br(`block{hash:"9d432"}`)))

	// Marshal and Unmarshal 'h'
	bts, err := h.Marshal()
	require.NoError(t, err)
	h2, err := Unmarshal(bts)
	require.NoError(t, err)
	require.True(t, equals(h, h2))

	// Modify 'h', and Marshal and Unmarshal it again
	require.NoError(t, h.PutFile("/bar/buzz2", br(`block{hash:"8e02c"}`)))
	bts, err = h.Marshal()
	require.NoError(t, err)
	h3, err := Unmarshal(bts)
	require.NoError(t, err)
	require.True(t, equals(h, h3))

	// Make sure 'h2' does not equal 'h' or 'h3'
	require.False(t, equals(h, h2))
	require.False(t, equals(h2, h3))
}
