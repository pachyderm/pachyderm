package hashtree

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// Parse a string as a BlockRef
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

// Render s as a hex string -- pure convenience/abbreviation function
func hex(s []byte) string {
	return fmt.Sprintf("%x", s)
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

func (h *HashTree) debug() {
	fmt.Println("################################")
	for k, v := range h.Fs {
		fmt.Printf("%s: %s\n", k, proto.CompactTextString(v))
	}
}

func TestPutFileBasic(t *testing.T) {
	h := HashTree{}
	h.PutFile("/foo", br(`block{hash:"20c27"}`))
	require.Equal(t, int64(1), h.Fs["/foo"].Size)
	require.Equal(t, int64(1), h.Fs[""].Size)

	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	require.Equal(t, int64(1), h.Fs["/dir/bar"].Size)
	require.Equal(t, int64(1), h.Fs["/dir"].Size)
	require.Equal(t, int64(2), h.Fs[""].Size)
	h.PutFile("/dir/buzz", br(`block{hash:"8e02c"}`))
	require.Equal(t, int64(1), h.Fs["/dir/buzz"].Size)
	require.Equal(t, int64(2), h.Fs["/dir"].Size)
	require.Equal(t, int64(3), h.Fs[""].Size)

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
	oldSha := make([]byte, len(h.Fs["/foo"].Hash))
	copy(oldSha, h.Fs["/foo"].Hash)
	require.Equal(t, int64(1), h.Fs["/foo"].Size)
	h.PutFile("/foo", br(`block{hash:"413e7"}`))
	require.NotEqual(t, oldSha, h.Fs["/foo"].Hash)
	require.Equal(t, int64(2), h.Fs["/foo"].Size)
}

// func TestPutDirError(t *testing.T) {
// 	// Put root dir
// 	h := &HashTree{}
// 	h.PutDir("/")
// 	require.Equal(t, 1, len(h.Fs))
//
// 	err := h.DeleteDir("/does/not/exist")
// 	require.YesError(t, err)
// 	require.Equal(t, 1, len(h.Fs))
// }

func TestDeleteFile(t *testing.T) {
	h := HashTree{}
	emptySha := sha256.Sum256([]byte{})

	// Create '/dir'
	h.PutDir("/dir")
	require.Equal(t, len(h.Fs), 2)
	require.Equal(t, emptySha[:], h.Fs["/dir"].Hash)
	require.Equal(t, int64(0), h.Fs["/dir"].Size)
	require.Equal(t, int64(0), h.Fs[""].Size)
	require.Equal(t, 1, len(h.Fs[""].DirNode.Children))

	rootHash := make([]byte, len(h.Fs[""].Hash))
	copy(rootHash, h.Fs[""].Hash)

	nodes, err := h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))

	// Add a child of '/dir'
	h.PutFile("/dir/foo", br(`block{hash:"20c27"}`))
	require.Equal(t, len(h.Fs), 3)
	require.NotEqual(t, emptySha[:], h.Fs["/dir"].Hash)
	require.NotEqual(t, int64(0), h.Fs["/dir"].Size)

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))

	// Delete the child and make sure '/dir' and '/' go back to old state
	h.DeleteFile("/dir/foo")
	require.Equal(t, len(h.Fs), 2)
	require.Equal(t, emptySha[:], h.Fs["/dir"].Hash)
	require.Equal(t, int64(0), h.Fs["/dir"].Size)
	require.Equal(t, rootHash, h.Fs[""].Hash)
	require.Equal(t, int64(0), h.Fs[""].Size)

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 0, len(nodes))

	// Create /dir/foo again, and copy sha to check against later
	h.PutFile("/dir/foo", br(`block{hash:"20c27"}`))
	require.Equal(t, int64(1), h.Fs["/dir"].Size)
	require.Equal(t, 1, len(h.Fs[""].DirNode.Children))
	require.Equal(t, 1, len(h.Fs["/dir"].DirNode.Children))

	dirHash := make([]byte, len(h.Fs["/dir"].Hash))
	copy(dirHash, h.Fs["/dir"].Hash)
	copy(rootHash, h.Fs[""].Hash)

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))

	// Add a second child of '/dir'
	h.PutFile("/dir/bar", br(`block{hash:"ebc57"}`))
	require.Equal(t, len(h.Fs), 4)
	require.NotEqual(t, dirHash, h.Fs["/dir"].Hash)
	require.Equal(t, int64(2), h.Fs["/dir"].Size)

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 2, len(nodes))

	// Delete the child and make sure '/dir' and '/' go back to old state
	h.DeleteFile("/dir/bar")
	require.Equal(t, len(h.Fs), 3)
	require.Equal(t, dirHash, h.Fs["/dir"].Hash)
	require.Equal(t, int64(1), h.Fs["/dir"].Size)
	require.Equal(t, rootHash, h.Fs[""].Hash)
	require.Equal(t, int64(1), h.Fs[""].Size)
	require.Equal(t, 1, len(h.Fs[""].DirNode.Children))

	nodes, err = h.List("/dir")
	require.NoError(t, err)
	require.Equal(t, 1, len(nodes))
}

// func TestDeleteDir(t *testing.T) {
// 	h := HashTree{}
// 	emptySha := sha256.Sum256([]byte{})
// 	h.PutDir("/dir")
// 	require.Equal(t, emptySha[:], h.Fs["/dir"].Hash)
// 	require.Equal(t, []string(nil), h.Fs["/dir"].DirNode.Children)
//
// 	h.PutDir("/dir/foo")
// 	nodes, err := h.List("/dir")
// 	require.NoError(t, err)
// 	require.Equal(t, 1, len(nodes))
// 	require.NotEqual(t, emptySha[:], h.Fs["/dir"].Hash)
// 	require.NotEqual(t, []string(nil), h.Fs["/dir"].DirNode.Children)
//
// 	h.DeleteDir("/dir/foo")
// 	nodes, err = h.List("/dir")
// 	fmt.Printf("%s\n", nodes)
// 	require.NoError(t, err)
// 	require.Equal(t, 0, len(nodes))
// 	require.Equal(t, emptySha[:], h.Fs["/dir"].Hash)
// 	require.Equal(t, []string(nil), h.Fs["/dir"].DirNode.Children)
//
// 	h.PutFile("/dir/foo/bar", br(`block{hash:"20c27"}`))
// 	h.DeleteDir("/dir/foo")
// 	nodes, err = h.List("/dir")
// 	fmt.Printf("%s\n", nodes)
// 	require.NoError(t, err)
// 	require.Equal(t, 0, len(nodes))
// 	require.Equal(t, emptySha[:], h.Fs["/dir"].Hash)
// 	require.Equal(t, []string(nil), h.Fs["/dir"].DirNode.Children)
// 	require.Equal(t, len(h.Fs), 1)
// }

func TestGlobFile(t *testing.T) {
	h := HashTree{}
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

func TestInsertStr(t *testing.T) {
	// Don't expand slice
	x := make([]string, 3, 4)
	copy(x, []string{"a", "b", "d"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.Equal(t, 4, cap(x))

	// Expand slice by constant amount
	x = make([]string, 3, 3)
	copy(x, []string{"a", "b", "d"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)

	// Expand slice by factor (may fail if constant grows)
	x = make([]string, 25, 25)
	copy(x, []string{"a", "b", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y",
		"z"}, x)
	require.Equal(t, 26, len(x))
	require.True(t, cap(x) >= 26)

	// insert at beginning
	x = make([]string, 3, 3)
	copy(x, []string{"b", "c", "d"})
	insertStr(&x, "a")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)

	// insert at end
	x = make([]string, 3, 3)
	copy(x, []string{"a", "b", "c"})
	insertStr(&x, "d")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)
}

// To Test
// - Replace file (DeleteFile + PutFile of same file, same name) does not change hash
// - renaming a file (DeleteFile + PutFile same file, different name) changes hash
// - updating a file in-place (PutFile with same name, new contents) changes hash
// - PutFile is commutative (putting the same files yields the same hash, no
//     matter what order the files are put)
