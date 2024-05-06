package fuse

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestGrandparentHelper(t *testing.T) {
	// Test case 1: path1 is a grandparent of path2
	path1 := "/path/to/grandparent/parent/"
	path2 := "/path/to/grandparent/parent/child/file.txt"
	isGrandparent, intermediate := isGrandparentOf(path1, path2)
	require.True(t, isGrandparent)
	require.Equal(t, "/child/", intermediate)

	// Test case 2: path1 is not a grandparent of path2
	path1 = "/path/to/grandparent/"
	path2 = "/path/to/other/parent/child/file.txt"
	isGrandparent, intermediate = isGrandparentOf(path1, path2)
	require.False(t, isGrandparent)
	require.Equal(t, "", intermediate)

	// Test case 3: path1 and path2 are the same
	path1 = "/path/to/grandparent/"
	path2 = "/path/to/grandparent/"
	isGrandparent, intermediate = isGrandparentOf(path1, path2)
	require.False(t, isGrandparent)
	require.Equal(t, "", intermediate)

	// Test case 4: path1 is a great grandparent of path2
	path1 = "/path/to/grandparent/"
	path2 = "/path/to/grandparent/parent/child/file.txt"
	isGrandparent, intermediate = isGrandparentOf(path1, path2)
	require.True(t, isGrandparent)
	require.Equal(t, "/parent/child/", intermediate)

	// Test case 5: path1 is a parent of path2
	path1 = "/path/to/grandparent/parent"
	path2 = "/path/to/grandparent/parent/child/"
	isGrandparent, intermediate = isGrandparentOf(path1, path2)
	require.False(t, isGrandparent)
	require.Equal(t, "", intermediate)

	// Test case 6: root path
	path1 = ""
	path2 = "/path/to/grandparent/parent/child"
	isGrandparent, intermediate = isGrandparentOf(path1, path2)
	require.True(t, isGrandparent)
	require.Equal(t, "/path/to/grandparent/parent/", intermediate)
}

func TestParentHelper(t *testing.T) {
	require.True(t, isParentOf("", "/dir"))
	require.True(t, isParentOf("/", "dir"))
	require.True(t, isParentOf("/", "/dir/"))
	require.True(t, isParentOf("dir", "/dir/sdir"))
	require.True(t, isParentOf("/dir", "/dir/sdir"))
	require.True(t, isParentOf("/dir/", "/dir/sdir/"))
	require.True(t, isParentOf("/dir/sdir", "/dir/sdir/file"))

	require.False(t, isParentOf("", ""))
	require.False(t, isParentOf("/", ""))
	require.False(t, isParentOf("", "/"))
	require.False(t, isParentOf("dir", "/dir"))
	require.False(t, isParentOf("/dir", "/dir"))
	require.False(t, isParentOf("/dir", "/dir/sdir/file"))
	require.False(t, isParentOf("/dir/sdir1", "/dir/sdir2"))
}

func TestDescendantHelper(t *testing.T) {
	require.True(t, isDescendantOf("/", "/"))
	require.True(t, isDescendantOf("dir", ""))
	require.True(t, isDescendantOf("dir/", ""))
	require.True(t, isDescendantOf("dir", "/dir"))
	require.True(t, isDescendantOf("dir", "/dir/"))
	require.True(t, isDescendantOf("/dir/", "/dir"))
	require.True(t, isDescendantOf("/dir/", "dir/"))
	require.True(t, isDescendantOf("dir/sdir/file", "/dir/sdir/"))

	require.False(t, isDescendantOf("/", "dir"))
	require.False(t, isDescendantOf("", "dir"))
	require.False(t, isDescendantOf("dir", "/dir2"))
	require.False(t, isDescendantOf("dir2", "/dir"))
	require.False(t, isDescendantOf("dir", "/dir/sdir"))
	require.False(t, isDescendantOf("dir/sdir", "/dir2"))
	require.False(t, isDescendantOf("dir/sdir", "/dir/sdir2"))
}
