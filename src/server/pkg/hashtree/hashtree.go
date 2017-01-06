package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type Names []string

func (n Names) Len() int {
	return len(n)
}

func (n Names) Less(i, j int) bool {
	return n[i] < n[j]
}

func (n Names) Swap(i, j int) {
	tmp := n[i]
	n[i] = n[j]
	n[j] = tmp
}

// cleanPath converts a path into a form that we use internally
// Basically we make sure that it has a leading slash and no trailing slash.
func cleanPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return pathlib.Clean(path)
}

// UpdateHash uses the node's internal state to update the hash
// of the node
func (n *Node) UpdateHash() error {
	if n.DirNode != nil {
		sort.Sort(Names(n.DirNode.Children))
		var buf bytes.Buffer
		for _, child := range n.DirNode.Children {
			if _, err := buf.WriteString(child); err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\"; this is likely a bug", n.Name, err)
			}
		}
		cksum := sha256.Sum256(buf.Bytes())
		n.Hash = cksum[:]
	} else if n.FileNode != nil {
		var buf bytes.Buffer
		for _, blockRef := range n.FileNode.BlockRefs {
			if _, err := buf.WriteString(fmt.Sprintf("%s:%d:%d", blockRef.Block.Hash, blockRef.Range.Lower, blockRef.Range.Upper)); err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\"; this is likely a bug", n.Name, err)
			}
		}
	} else {
		return fmt.Errorf("malformed node %s: it's neither a file nor a directory", n.Name)
	}
	return nil
}

// GlobFile returns a list of nodes that match the given glob pattern
func (h *HashTree) GlobFile(pattern string) ([]*Node, error) {
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	pattern = cleanPath(pattern)

	var res []*Node
	for p, node := range h.Fs {
		matched, err := pathlib.Match(pattern, p)
		if err != nil {
			if err == pathlib.ErrBadPattern {
				return nil, MalformedGlobErr
			}
			return nil, err
		}
		if matched {
			res = append(res, node)
		}
	}
	return res, nil
}

// ListFile returns the Nodes corresponding to the files and directories under 'path'
func (h *HashTree) ListFile(path string) ([]*Node, error) {
	path = cleanPath(path)
	node, ok := h.Fs[path]
	if !ok {
		return nil, PathNotFoundErr
	}
	d := node.DirNode
	if d == nil {
		return nil, fmt.Errorf("the file at %s is not a directory", path)
	}
	var result []*Node
	for _, childName := range d.Children {
		childPath := pathlib.Join(path, childName)
		child, ok := h.Fs[pathlib.Join(path, childPath)]
		if !ok {
			return nil, fmt.Errorf("malformed hash tree; the node %s is expected to exist but is not found; this is likely a bug", childPath)
		}
		result = append(result, child)
	}
	return result, nil
}

// PutFile inserts a file into the hierarchy
func (h *HashTree) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	path = cleanPath(path)

	// Update/create the file node
	node, ok := h.Fs[path]
	if !ok {
		name := pathlib.Base(path)
		node = &Node{
			Name: name,
			FileNode: &FileNode{
				BlockRefs: blockRefs,
			},
		}
		h.Fs[path] = node
	} else {
		node.FileNode.BlockRefs = append(node.FileNode.BlockRefs)
	}
	if err := node.UpdateHash(); err != nil {
		return err
	}

	// Update/create parent directory nodes
	for path != "/" {
		dir, child := pathlib.Split(path)
		node, ok := h.Fs[dir]
		if !ok {
			node = &Node{
				Name: pathlib.Base(dir),
				DirNode: &DirectoryNode{
					Children: []string{child},
				},
			}
			h.Fs[dir] = node
		} else {
			node.DirNode.Children = append(node.DirNode.Children, child)
		}
		if err := node.UpdateHash(); err != nil {
			return err
		}
		path = cleanPath(dir)
	}
	return nil
}
