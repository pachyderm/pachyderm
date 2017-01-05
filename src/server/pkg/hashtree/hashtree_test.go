package hashtree

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestGlobFile1(t *testing.T) {
	h := &HashTree{}
	fooNode = &Node{
		FileNode: &FileNode{
			Name: "foo",
		},
	}
	dirNode = &Node{
		DirNode: &DirNode{
			Name: "dir",
		},
	}
	dirBarNode = &Node{
		FileNode: &FileNode{
			Name: "bar",
		},
	}
	dirBuzzNode = &Node{
		FileNode: &FileNode{
			Name: "buzz",
		},
	}
	h.Fs = make(map[string]*Node)
	h.Fs["/foo"] = fooNode
	h.Fs["/dir"] = dirNode
	h.Fs["/dir/bar"] = dirBarNode
	h.Fs["/dir/buzz"] = dirBuzzNode

	for _, pattern := range []string{"*", "/*"} {
		nodes, err := h.GlobFile("*")
		require.NoError(t, err)
		require.Equal(t, 2, len(nodes))
		for _, node := range nodes {
			require.EqualOneOf(t, []*Node{fooNode, dirNode}, node)
		}
	}

	for _, pattern := range []string{"*/*", "/*/*"} {
		nodes, err := h.GlobFile("*")
		require.NoError(t, err)
		require.Equal(t, 2, len(nodes))
		for _, node := range nodes {
			require.EqualOneOf(t, []*Node{dirBarNode, dirBuzzNode}, node)
		}
	}
}
