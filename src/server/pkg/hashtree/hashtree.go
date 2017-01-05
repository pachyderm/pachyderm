package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path"
	pathlib "path"
	"sort"
	"strings"
)

func (h *HashTree) GlobFile(pattern string) ([]*Node, error) {
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	var res []*Node
	for p, node := range h.Fs {
		matched, err := path.Match(pattern, p)
		if err != nil {
			if err == path.ErrBadPattern {
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

func (h *HashTree) ListDir(pattern string) ([]*Node, error) {
	d := n.DirNode
	if d == nil {
		return nil, fmt.Errorf("The file at %s is not a directory", dir)
	}
	return d.Child, nil
}

// Updates the hash of every node that is at a prefix of 'path'. This is called
// e.g. at the end of PutFile, when the hash of the directory receiving the new
// file must be updated, as well as the parent of that directory, and the parent
// of the parent, etc. up to the root.
func (h *HashTree) updateHashes(path string) {
	// Must update tree from leaf to root, otherwise intermediate directory hashes
	// will be wrong, as child directory hashes are updated
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] != '/' {
			continue
		}
		children, err := h.ListDir(path[:i+1])
		if err != nil {
			// This method should only be called internally--any errors are our fault
			panic(fmt.Sprintf(
				"Error while attempting to update the hash of %s: \"%s\"", path, err))
		}
		sort.Strings(children)
		var buf bytes.Buffer
		for _, child := range children {
			buf.WriteString(child)
		}
		cksum := sha256.Sum256(buf.Bytes())
		h.Fs[path].DirNode.Hash = cksum[:]
	}
}

// Inserts a file into the hierarchy. If 'path' ends in "/", it is created as
// a directory and 'hash' is ignored.
func (h *HashTree) PutFile(path string, hash []byte) error {
	// Clean up 'path'
	isDir := path[len(path)-1] == '/'
	path = pathlib.Clean(path)
	if isDir {
		path += "/"
	}

	// Create all directories in 'path'
	var curDir *DirectoryNode = nil
	curPath := ""
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			curPath = path[:i+1]
			if h.Fs[curPath] == nil {
				// Create new directory if none exists
				newDir := &Node{
					DirNode: &DirectoryNode{},
				}
				if curDir != nil {
					curDir.Child = append(curDir.Child, newDir)
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
