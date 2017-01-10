package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// This is not in "math", unfortunately
func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

// Assuming that 'ss' is sorted, inserts 's' into 'ss', preserving sorting. If
// a copy is necessary (because cap(ss) is too small), only does one copy (as
// DirectoryNode.Children might get quite large, 1e5-1e6 entries)
//
// This is used to preserve the order in DirectoryNode.Children, which must
// be maintained so that equivalent directories have the same hash.
//
// Returns 'true' if newS was added to 'ss' and 'false' otherwise (if newS is
// already in 'ss').
func insertStr(ss *[]string, newS string) bool {
	sz := cap(*ss)
	idx := sort.SearchStrings(*ss, newS)
	if idx >= len(*ss) || (*ss)[idx] != newS {
		// Need to insert new element
		if sz >= (len(*ss) + 1) {
			*ss = (*ss)[:len(*ss)+1]
			copy((*ss)[idx+1:], (*ss)[idx:])
			(*ss)[idx] = newS
		} else {
			// Need to grow ss (i.e. make a copy)
			// - Using a factor (instead of always adding a constant number of
			//   elements) ensures amortized constant time for insertions, and keeping
			//   it reasonably low avoids wasting too much space. Current value of
			//   1.33 is arbitrary.
			// - If sz is small, grow sz by at least a constant amount (must be >=1,
			//   currently 10)
			cap1, cap2 := int(float64(sz)*1.33), sz+10
			newSs := make([]string, len(*ss)+1, max(cap1, cap2))
			copy(newSs, (*ss)[:idx])
			copy(newSs[idx+1:], (*ss)[idx:])
			newSs[idx] = newS
			*ss = newSs
		}
		return true
	}
	return false
}

// Removes 's' from 'ss', preserving the sorted order of 'ss' (for removing
// child strings from DirectoryNodes.
func removeStr(ss *[]string, s string) bool {
	idx := sort.SearchStrings(*ss, s)
	if idx == len(*ss) {
		return false
	}
	copy((*ss)[idx:], (*ss)[idx+1:])
	*ss = (*ss)[:len(*ss)-1]
	return true
}

// Updates the hash of the node N at 'path'. If this changes N's hash, that will
// render the hash of N's parent (if any) invalid, and this must be called for
// all parents of 'path' back to the root.
func (h *HashTree) updateHash(path string) error {
	n, ok := h.Fs[path]
	if !ok {
		return fmt.Errorf("Could not find node \"%s\" to update hash", path)
	}

	// Compute hash of 'n'
	var b bytes.Buffer
	if n.DirNode != nil {
		// PutFile keeps n.DirNode.Children sorted, so the order is stable
		for _, child := range n.DirNode.Children {
			n, ok := h.Fs[join(path, child)]
			if !ok {
				return fmt.Errorf("could not find node for \"%s\" while updating hash "+
					"of \"%s\" (this is a bug)", join(path, child), path)
			}
			// Write Name and Hash
			_, err := b.WriteString(fmt.Sprintf("%s:%s:", n.Name, n.Hash))
			if err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\" (likely a bug)", n.Name, err)
			}
		}
		cksum := sha256.Sum256(b.Bytes())
		n.Hash = cksum[:]
	} else if n.FileNode != nil {
		for _, blockRef := range n.FileNode.BlockRefs {
			_, err := b.WriteString(fmt.Sprintf("%s:%d:%d:",
				blockRef.Block.Hash, blockRef.Range.Lower, blockRef.Range.Upper))
			if err != nil {
				return fmt.Errorf(
					"error updating the hash of %s: \"%s\" (likely a bug)", n.Name, err)
			}
		}
	} else {
		return fmt.Errorf(
			"malformed node %s: it's neither a file nor a directory", n.Name)
	}
	cksum := sha256.Sum256(b.Bytes())

	// Update hash of 'n'
	n.Hash = cksum[:]
	return nil
}

// PutFile inserts a file into the hierarchy
func (h *HashTree) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	path = clean(path)
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}

	// Compute size of new node
	var size int64
	for _, blockRef := range blockRefs {
		size += int64(blockRef.Range.Upper - blockRef.Range.Lower)
	}

	// Add new file or append new content if file exists, and update the hash
	node, ok := h.Fs[path]
	if !ok {
		name := base(path)
		node = &Node{
			Name: name,
			Size: size,
			FileNode: &FileNode{
				BlockRefs: blockRefs,
			},
		}
		h.Fs[path] = node
	} else {
		node.Size += size
		node.FileNode.BlockRefs = append(node.FileNode.BlockRefs, blockRefs...)
	}
	h.updateHash(path)

	// Create & normalize parent directories back to the root
	// TODO(msteffen): There is very similar code in {Put,Delete}{File,Dir}, but
	// there are small differences (e.g. whether 'size' is increased on the way
	// up, or whether missing directories indicate a bug). See if there's a way
	// to factor this into a helper function
	for path != "" {
		parent, child := split(path)

		// Create parent dir if missing
		pnode, ok := h.Fs[parent]
		if !ok {
			pnode = &Node{
				Name:    base(parent),
				Size:    0, // Will update below
				DirNode: &DirectoryNode{},
			}
			h.Fs[parent] = pnode
		} else if pnode.DirNode == nil {
			return fmt.Errorf("parent of the Node \"%s\" is a FileNode (this is a "+
				"bug; try deleting the parent) ", path)
		}
		insertStr(&pnode.DirNode.Children, child)
		pnode.Size += size
		h.updateHash(parent)
		path = parent
	}
	return nil
}

func (h *HashTree) PutDir(path string) error {
	path = clean(path)
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}
	if node, ok := h.Fs[path]; ok {
		if node.DirNode == nil {
			return fmt.Errorf("could not create directory at \"%s\" as a "+
				"non-directory file is already there", path)
		}
		return nil
	}
	// Create orphaned directory at 'path'
	h.Fs[path] = &Node{
		Name:    base(path),
		Size:    0,
		DirNode: &DirectoryNode{},
	}
	h.updateHash(path)

	// Update hashes back to root
	for path != "" {
		parent, child := split(path)
		// Create parent dir at 'dir(path)' if none exists (path is not root)
		pnode, ok := h.Fs[parent]
		if !ok {
			pnode = &Node{
				Name:    base(parent),
				Size:    0,
				DirNode: &DirectoryNode{},
			}
			h.Fs[parent] = pnode
		} else if pnode.DirNode == nil {
			return fmt.Errorf("parent of the Node \"%s\" is a FileNode (this is a "+
				"bug; try deleting the parent) ", path)
		}
		insertStr(&pnode.DirNode.Children, child)
		h.updateHash(parent)
		path = parent
	}
	return nil
}

func (h *HashTree) DeleteFile(path string) error {
	node, ok := h.Fs[path]
	if !ok {
		return nil
	}
	if node.FileNode == nil {
		return fmt.Errorf("node at \"%s\" is not a FileNode (try DeleteDir)", path)
	}

	// Compute size of file being removed
	var size int64
	for _, blockRef := range node.FileNode.BlockRefs {
		size += int64(blockRef.Range.Upper - blockRef.Range.Lower)
	}

	// Remove file from map (h.Fs) and from the parent's Children list
	delete(h.Fs, path)
	pnode, ok := h.Fs[dir(path)]
	if !ok {
		return fmt.Errorf("attempted to delete orphaned file \"%s\" (this is "+
			"a bug; try recreating the file and deleting)", path)
	}
	if pnode.DirNode == nil {
		return fmt.Errorf("parent of the Node \"%s\" is not a directory (this "+
			"is a bug; try deleting the parent) ", path)
	}
	if !removeStr(&pnode.DirNode.Children, base(path)) {
		return fmt.Errorf("parent of the node \"%s\" does not contain it (this "+
			"is a bug; try deleting the parent", path)
	}

	// Update hashes back to root
	for path != "" {
		parent := dir(path)
		if pnode, ok = h.Fs[parent]; !ok {
			return fmt.Errorf("attempted to delete orphaned file \"%s\" (this is "+
				"a bug; try recreating the file and deleting)", path)
		}
		fmt.Printf("    %s: %s\n", parent, proto.MarshalTextString(pnode))
		pnode.Size -= size
		h.updateHash(parent)
		fmt.Printf("(A) %s: %s\n", parent, proto.MarshalTextString(pnode))
		path = parent
	}
	return nil
}

// Delete the node at 'path' but don't update the size or hash of any parents
// This is more efficient than e.g. calling DeleteFile on every node in a
// directory, as that would call h.updateHash for every file inside the
// directory that's removed. This might result in jk
func deleteNodeNoUpdate(path string) {
}
func (h *HashTree) DeleteDir(path string) error {
	node, ok := h.Fs[path]
	if !ok {
		return nil
	}
	if node.FileNode == nil {
		return fmt.Errorf("node at \"%s\" is not a FileNode (try DeleteDir)", path)
	}

	// Remove all of the children of this node

	return nil
}

// Get returns the node associated with the path
func (h *HashTree) Get(path string) (*Node, error) {
	path = clean(path)
	node, ok := h.Fs[path]
	if !ok {
		return nil, PathNotFoundErr
	}
	return node, nil
}

// List returns the Nodes corresponding to the files and directories under 'path'
func (h *HashTree) List(path string) ([]*Node, error) {
	path = clean(path)
	node, ok := h.Fs[path]
	if !ok {
		return nil, nil // return empty list
	}
	d := node.DirNode
	if d == nil {
		return nil, fmt.Errorf("the file at %s is not a directory", path)
	}
	result := make([]*Node, len(d.Children))
	for i, child := range d.Children {
		result[i], ok = h.Fs[join(path, child)]
		if !ok {
			return nil, fmt.Errorf("could not find node for \"%s\" while listing "+
				"\"%s\" (this is a bug)", join(path, child), path)
		}
	}
	return result, nil
}

// Glob returns a list of nodes that match the given glob pattern
func (h *HashTree) Glob(pattern string) ([]*Node, error) {
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	pattern = clean(pattern)
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
