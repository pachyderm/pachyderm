package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"
	"sort"
	"strings"
)

// cleanPath converts a path into a form that we use internally
// Basically we make sure that it has a leading slash and no trailing slash.
func cleanPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return pathlib.Clean(path)
}

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
	d := h.Fs[path].DirNode
	if d == nil {
		return nil, fmt.Errorf("the file at %s is not a directory", path)
	}
	var result []*Node
	for _, childName := range d.Child {
		childPath := pathlib.Join(path, childName)
		child, ok := h.Fs[pathlib.Join(path, childPath)]
		if !ok {
			return nil, fmt.Errorf("malformed hash tree; the node %s is expected to exist but is not found; this is likely a bug", childPath)
		}
		result = append(result, child)
	}
	return result, nil
}

// Custom wrapper type to sort the list of children returned by ListFile
// lexicographically
type NodeList []*Node

func (l NodeList) Len() int {
	return len(l)
}

func (l NodeList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l NodeList) Swap(i, j int) {
	tmp := l[i]
	l[i] = l[j]
	l[j] = tmp
}

// Updates the hash of every node that is at a prefix of 'path'. This is called
// e.g. at the end of PutFile, when the hash of the directory receiving the new
// file must be updated, as well as the parent of that directory, and the parent
// of the parent, etc. up to the root.
func (h *HashTree) updateHashes(path string) error {
	// Must update tree from leaf to root, otherwise intermediate directory hashes
	// will be wrong, as child directory hashes are updated
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] != '/' {
			continue
		}
		children, err := h.ListFile(path[:i+1])
		if err != nil {
			// This method should only be called internally--any errors are our fault
			return fmt.Errorf(
				"error updating the hash of %s: \"%s\"; this is likely a bug", path, err)
		}
		sort.Sort(NodeList(children))
		var buf bytes.Buffer
		for _, child := range children {
			if _, err := buf.WriteString(child.Name); err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\"; this is likely a bug", path, err)
			}
			if _, err := buf.Write(child.Hash); err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\"; this is likely a bug", path, err)
			}
		}
		cksum := sha256.Sum256(buf.Bytes())
		h.Fs[path].Hash = cksum[:]
	}
	return nil
}

// PutFile inserts a file into the hierarchy
func (h *HashTree) PutFile(path string, hash []byte) error {
	path = cleanPath(path)

	// Create all directories in 'path'
	var curDir *DirectoryNode
	curPath := ""
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			name := path[len(curPath) : i+1]
			curPath = path[:i+1]
			if h.Fs[curPath] == nil {
				// Create new directory if none exists
				newDir := &Node{
					Name:    name,
					Hash:    []byte{}, // update later
					DirNode: &DirectoryNode{},
				}
				if curDir != nil {
					curDir.Children = append(curDir.Children, name)
				}
				h.Fs[curPath] = newDir
			}
			curDir = h.Fs[curPath].DirNode
		}
	}

	// Create the file at the end of the path (if the path doesn't end in a
	// directory)
	newFileName := path[len(curPath):]
	if len(newFileName) != 0 {
		curDir.Child = append(curDir.Child, &Node{
			FileNode: &FileNode{hash: hash},
		})
	}

	// Update the hash values of all directories in 'path'
	h.updateHashes(path)
}
