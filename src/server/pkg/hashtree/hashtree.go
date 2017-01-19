package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

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

	// Update hash of 'n'
	cksum := sha256.Sum256(b.Bytes())
	n.Hash = cksum[:]
	return nil
}

// Visits every ancestor of 'path' (excluding 'path' itself, leaf to root (i.e.
// end of 'path' to beginning), and modifying each node along the way. This is
// useful for propagating changes to size and hash upwards.
//
// - If 'createDirs' is true, and there is any prefix along 'path' with no node,
//   a new empty directory will be created at that path (it will be up to the
//   user to add children to that node and recompute its hash, in update())
//
// - The *Node argument to update is guaranteed to be the node corresponding to
//   the first string argument to 'update' (i.e. the parent path), and is always
//   a Directory Node
func (h *HashTree) visit(path string, createDirs bool,
	update func(*Node, string, string) error) error {
	for path != "" {
		parent, child := split(path)
		pnode, ok := h.Fs[parent]
		if !ok {
			if createDirs {
				pnode = &Node{
					Name:    base(parent),
					Size:    0,
					DirNode: &DirectoryNode{},
				}
				h.Fs[parent] = pnode
			} else {
				return fmt.Errorf("attempted to visit or delete orphaned file \"%s\" "+
					"(this is a bug; try recreating the file and deleting)", path)
			}
		} else if pnode.DirNode == nil {
			return fmt.Errorf("parent of the Node \"%s\" is a FileNode (this is a "+
				"bug; try deleting the parent) ", path)
		}
		if err := update(pnode, parent, child); err != nil {
			return err
		}
		path = parent
	}
	return nil
}

// Removes the node at 'path' from h.Fs if it's present, along with all of its
// children, recursively.
//
// This will not update the hash of any parent of 'path'. This helps us avoid
// updating the hash of path's parents unnecessarily; if 'path' is a directory
// with e.g. 10k children, updating the parents' hashes after all files have
// been removed from h.Fs (instead of updating all parents' hashesafter
// removing each file) may save substantial time.
func (h *HashTree) removeFromMap(path string) error {
	n, ok := h.Fs[path]
	if !ok {
		return nil
	} else if n.FileNode != nil {
		delete(h.Fs, path)
	} else if n.DirNode != nil {
		for _, child := range n.DirNode.Children {
			if err := h.removeFromMap(pathlib.Join(path, child)); err != nil {
				return err
			}
		}
		delete(h.Fs, path)
	} else {
		return fmt.Errorf("could not remove unidentified node at \"%s\"", path)
	}
	return nil
}

// Removes the node at 'path' from 'h', updating the hash of its ancestors
func (h *HashTree) deleteNode(path string) error {
	// Remove 'path' from h.Fs
	node, ok := h.Fs[path]
	if !ok {
		return fmt.Errorf("cannot remove node at \"%s\", as there is no such node " +
			"in the tree")
	}
	size := node.Size
	h.removeFromMap(path)

	// Remove 'path' from its parent directory
	parent, child := split(path)
	pnode, ok := h.Fs[parent]
	if !ok {
		return fmt.Errorf("attempted to delete orphaned file \"%s\" (this is "+
			"a bug; try recreating the file and deleting)", path)
	}
	if pnode.DirNode == nil {
		return fmt.Errorf("parent of the Node \"%s\" is not a directory (this "+
			"is a bug; try deleting the parent) ", path)
	}
	if !removeStr(&pnode.DirNode.Children, child) {
		return fmt.Errorf("parent of the node \"%s\" does not contain it (this "+
			"is a bug; try deleting the parent", path)
	}
	// Update hashes back to root
	return h.visit(path, false, func(pnode *Node, parent, child string) error {
		pnode.Size -= size
		h.updateHash(parent)
		return nil
	})
}

// Inserts a file into the hierarchy
func (h *HashTree) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	path = clean(path)
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}

	node, ok := h.Fs[path]
	if ok {
		if node.FileNode == nil {
			return fmt.Errorf("could not create regular file at \"%s\", as a non-"+
				"regular file (e.g. directory) is already there", path)
		}
	} else {
		// Create empty file at 'path' (new blocks will be appended below)
		node = &Node{
			Name:     base(path),
			FileNode: &FileNode{},
		}
		h.Fs[path] = node
	}
	// Append new block
	node.FileNode.BlockRefs = append(node.FileNode.BlockRefs, blockRefs...)

	// Compute size growth of ode
	var sizeGrowth int64
	for _, blockRef := range blockRefs {
		sizeGrowth += int64(blockRef.Range.Upper - blockRef.Range.Lower)
	}
	node.Size += sizeGrowth
	h.updateHash(path)

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, true, func(pnode *Node, parent, child string) error {
		insertStr(&pnode.DirNode.Children, child)
		pnode.Size += sizeGrowth
		h.updateHash(parent)
		return nil
	})
}

// Inserts an empty directory into the hierarchy
func (h *HashTree) PutDir(path string) error {
	path = clean(path)
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}
	if node, ok := h.Fs[path]; ok {
		if node.DirNode == nil {
			return fmt.Errorf("could not create directory at \"%s\", as a non-"+
				"directory file is already there", path)
		}
		return nil
	}
	// Create orphaned directory at 'path'
	h.Fs[path] = &Node{
		Name:    base(path),
		DirNode: &DirectoryNode{},
	}
	h.updateHash(path)

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, true, func(pnode *Node, parent, child string) error {
		insertStr(&(pnode.DirNode.Children), child)
		h.updateHash(parent)
		return nil
	})
}

// Deletes the file at 'path'.
func (h *HashTree) DeleteFile(path string) error {
	node, ok := h.Fs[path]
	if !ok {
		return fmt.Errorf("No file at \"%s\"", path)
	}
	if node.FileNode == nil {
		return fmt.Errorf("node at \"%s\" is not a FileNode (try DeleteDir)", path)
	}
	// Remove file from map (h.Fs) and from the parent's Children list
	return h.deleteNode(path)
}

func (h *HashTree) DeleteDir(path string) error {
	node, ok := h.Fs[path]
	if !ok {
		return fmt.Errorf("No directory at \"%s\"", path)
	}
	if node.DirNode == nil {
		return fmt.Errorf("node at \"%s\" is not a DirectoryNode (try DeleteFile)",
			path)
	}
	// Remove file from map (h.Fs) and from the parent's Children list
	return h.deleteNode(path)
}

// Returns the node associated with the path
func (h *HashTree) Get(path string) (*Node, error) {
	path = clean(path)
	node, ok := h.Fs[path]
	if !ok {
		return nil, PathNotFoundErr
	}
	return node, nil
}

// Returns the Nodes corresponding to the files and directories under 'path'
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

// Returns a list of nodes that match the given glob pattern
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

// Merges the node at 'path' from 'from' into 'h'.
func (h *HashTree) mergeNode(path string, from Interface) error {
	// Fetch the nodes, and return error if one can't be fetched
	fromNode, err := from.Get(path)
	if err != nil && err != PathNotFoundErr {
		return err
	}
	toNode, err := h.Get(path)
	if err != nil && err != PathNotFoundErr {
		return err
	}

	if fromNode.DirNode != nil {
		if toNode != nil && toNode.DirNode == nil {
			return fmt.Errorf("node at \"%s\" is a directory in the target "+
				"HashTree, but not in the tree being merged", path)
		}
		h.PutDir(path)
		for _, child := range fromNode.DirNode.Children {
			h.mergeNode(join(path, child), from)
		}
	} else if fromNode.FileNode != nil {
		if toNode != nil && toNode.FileNode == nil {
			return fmt.Errorf("node at \"%s\" is a regular file in the target "+
				"HashTree, but not in the tree being merged", path)
		}
		h.PutFile(path, fromNode.FileNode.GetBlockRefs())
	} else {
		return fmt.Errorf("node at \"%s\" has unrecognized type: neither file "+
			"nor dir", path)
	}
	h.updateHash(path)
	return nil
}

// Merges 'from' into 'h'. Any files/directories in both 'from' and 'h' are
// merged (the content in 'from' is appended) and any files/directories that are
// only in 'from' are simply added to 'h'.
func (h *HashTree) Merge(from Interface) error {
	if _, err := from.Get("/"); err == PathNotFoundErr {
		return nil // No work necessary to merge blank tree
	}
	return h.mergeNode("/", from)
}
