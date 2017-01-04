package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path"
	"sort"
)

// List all files that are direct children of the directory 'dir'
func (h *HashTree) ListDir(dir string) ([]string, error) {
	n := h.Fs[dir]
	if n == nil {
		return nil, fmt.Errorf("Directory does not exist: %s", dir)
	}
	d := n.DirNode
	if d == nil {
		return nil, fmt.Errorf("The file at %s is not a directory", dir)
	}
	return d.child, nil
}

// Updates the hash of every node that is at a prefix of 'path'. This is called
// e.g. at the end of PutFile, when the hash of the directory receiving the new
// file must be updated, as well as the parent of that directory, and the parent
// of the parent, etc. up to the root.
func (h *HashTree) updateHashes(path string) {
	// Must update tree from leaf to root, otherwise intermediate directory hashes
	// will be wrong, as child directory hashes are updated
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] != "/" {
			continue
		}
		children, err := ListDir(path[:i+1])
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
		h.Fs[path].DirNode.hash = sha256.Sum256(buf.Bytes())
	}
}

// Inserts a file into the hierarchy. If 'path' ends in "/", it is created as
// a directory and 'hash' is ignored.
func (h *HashTree) PutFile(path string, hash []byte) error {
	// Clean up 'path'
	isDir = path[len(path)-1] == "/"
	path = path.Clean(path)
	if isDir {
		path += "/"
	}

	// Create all directories in 'path'
	var curDir *Node = nil
	curPath := ""
	for i := 0; i < len(path); i++ {
		if path[i] == "/" {
			if h.Fd[path[:i+1]] == nil {
				// Create new directory if none exists
				newDir := &Node{
					dir_node: &DirectoryNode{},
				}
				if curDir != nil {
					curDir.child = append(curDir.child, newDir)
				}
				h.Fd[path[:i+1]] = newDir
			}
			curPath = path[:i+1]
			curDir = h.Fd[curPath]
		}
	}

	// Create the file at the end of the path (if the path doesn't end in a
	// directory)
	newFileName = path[len(curPath):]
	if len(newFileName) != 0 {
		curDir.child = append(curDir.child, &Node{
			file_node: &FileNode{hash: hash},
		})
	}

	// Update the hash values of all directories in 'path'
	updateHashes(path)
}
