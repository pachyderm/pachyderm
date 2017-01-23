package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// updateHash updates the hash of the node N at 'path'. If this changes N's
// hash, that will render the hash of N's parent (if any) invalid, and this
// must be called for all parents of 'path' back to the root.
func (h *HashTree) updateHash(path string) error {
	n, ok := h.Fs[path]
	if !ok {
		return errorf(Internal, "Could not find node \"%s\" to update hash", path)
	}

	// Compute hash of 'n'
	var b bytes.Buffer
	if n.DirNode != nil {
		// PutFile keeps n.DirNode.Children sorted, so the order is stable
		for _, child := range n.DirNode.Children {
			n, ok := h.Fs[join(path, child)]
			if !ok {
				return errorf(Internal, "could not find node for \"%s\" while "+
					"updating hash of \"%s\"", join(path, child), path)
			}
			// Write Name and Hash
			_, err := b.WriteString(fmt.Sprintf("%s:%s:", n.Name, n.Hash))
			if err != nil {
				return errorf(Internal, "error updating hash of file at \"%s\": %s",
					path, err)
			}
		}
	} else if n.FileNode != nil {
		for _, blockRef := range n.FileNode.BlockRefs {
			_, err := b.WriteString(fmt.Sprintf("%s:%d:%d:",
				blockRef.Block.Hash, blockRef.Range.Lower, blockRef.Range.Upper))
			if err != nil {
				return errorf(Internal, "error updating hash of dir at \"%s\": %s",
					path, err)
			}
		}
	} else {
		return errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}

	// Update hash of 'n'
	cksum := sha256.Sum256(b.Bytes())
	n.Hash = cksum[:]
	return nil
}

func (h *HashTree) init() {
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}
	if h.Version == 0 {
		h.Version = 1
	}
}

// updateFn is used by 'visit'. The first parameter is the node being visited,
// the second parameter is the path of that node, and the third parameter is the
// child of that node from the 'path' argument to 'visit'.
//
// The *Node argument is guaranteed to have DirNode set (if it's not nil)--visit
// returns a 'PathConflict' error otherwise.
type updateFn func(*Node, string, string) error

// This can be passed to visit() to detect PathConflict errors early
func nop(*Node, string, string) error {
	return nil
}

// visit visits every ancestor of 'path' (excluding 'path' itself), leaf to
// root (i.e.  end of 'path' to beginning), and calls 'update' on each node
// along the way. For example, if 'visit' is called with 'path'="/path/to/file",
// then updateFn is called as follows:
//
// 1. update(node at "/path/to" or nil, "/path/to", "file")
// 2. update(node at "/path"    or nil, "/path",    "to")
// 3. update(node at "/"        or nil, "",         "path")
//
// This is useful for propagating changes to size and hash upwards.
func (h *HashTree) visit(path string, update updateFn) error {
	for path != "" {
		parent, child := split(path)
		pnode, ok := h.Fs[parent]
		if ok && pnode.DirNode == nil {
			return errorf(PathConflict, "attempted to visit \"%s\", but that is a "+
				"file", path)
		}
		if err := update(pnode, parent, child); err != nil {
			return err
		}
		path = parent
	}
	return nil
}

// removeFromMap removes the node at 'path' from h.Fs if it's present, along
// with all of its children, recursively.
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
			if err := h.removeFromMap(join(path, child)); err != nil {
				return err
			}
		}
		delete(h.Fs, path)
	} else {
		return errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}
	return nil
}

// PutFile inserts a file into the hierarchy
func (h *HashTree) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	h.init()
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Get/Create file node to which we'll append 'blockRefs'
	node, ok := h.Fs[path]
	if ok {
		if node.FileNode == nil {
			return errorf(PathConflict, "node at \"%s\" is not a regular file ", path)
		}
	} else {
		node = &Node{
			Name:     base(path),
			FileNode: &FileNode{},
		}
		h.Fs[path] = node
	}

	// Append new blocks
	node.FileNode.BlockRefs = append(node.FileNode.BlockRefs, blockRefs...)
	h.updateHash(path)

	// Compute size growth of node (i.e. amount of data we're appending)
	var sizeGrowth int64
	for _, blockRef := range blockRefs {
		sizeGrowth += int64(blockRef.Range.Upper - blockRef.Range.Lower)
	}
	node.Size += sizeGrowth

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, func(node *Node, parent, child string) error {
		if node == nil {
			node = &Node{
				Name:    base(parent),
				Size:    0,
				DirNode: &DirectoryNode{},
			}
			h.Fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		node.Size += sizeGrowth
		h.updateHash(parent)
		return nil
	})
}

// PutDir inserts an empty directory into the hierarchy
func (h *HashTree) PutDir(path string) error {
	h.init()
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Create orphaned directory at 'path' (or end early if a directory is there)
	if node, ok := h.Fs[path]; ok {
		if node.DirNode == nil {
			return errorf(PathConflict, "could not create directory at \"%s\"; a "+
				"non-directory file is already there", path)
		}
		return nil
	}
	h.Fs[path] = &Node{
		Name:    base(path),
		DirNode: &DirectoryNode{},
	}
	h.updateHash(path)

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, func(node *Node, parent, child string) error {
		if node == nil {
			node = &Node{
				Name:    base(parent),
				DirNode: &DirectoryNode{},
			}
			h.Fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		h.updateHash(parent)
		return nil
	})
}

// DeleteFile deletes the file at 'path'.
func (h *HashTree) DeleteFile(path string) error {
	h.init()
	path = clean(path)

	// Remove 'path' from h.Fs
	node, ok := h.Fs[path]
	if !ok {
		return errorf(PathNotFound, "no file at \"%s\"", path)
	}
	size := node.Size
	h.removeFromMap(path)

	// Remove 'path' from its parent directory
	parent, child := split(path)
	node, ok = h.Fs[parent]
	if !ok {
		return errorf(Internal, "attempted to delete orphaned file \"%s\"", path)
	}
	if node.DirNode == nil {
		return errorf(Internal, "node at \"%s\" is a file, but \"%s\" exists "+
			"under it")
	}
	if !removeStr(&node.DirNode.Children, child) {
		return errorf(Internal, "parent of \"%s\" does not contain it", path)
	}
	// Update hashes back to root
	return h.visit(path, func(node *Node, parent, child string) error {
		if node == nil {
			return errorf(Internal,
				"encountered orphaned file \"%s\" while deleting \"%s\"", path,
				join(parent, child))
		}
		node.Size -= size
		h.updateHash(parent)
		return nil
	})
}

// Get returns the node associated with the path
func (h *HashTree) Get(path string) (*Node, error) {
	h.init()
	path = clean(path)

	node, ok := h.Fs[path]
	if !ok {
		return nil, errorf(PathNotFound, "no node at \"%s\"", path)
	}
	return node, nil
}

// List returns the Nodes corresponding to the files and directories under
// 'path'
func (h *HashTree) List(path string) ([]*Node, error) {
	h.init()
	path = clean(path)

	node, ok := h.Fs[path]
	if !ok {
		return nil, nil // return empty list
	}
	d := node.DirNode
	if d == nil {
		return nil, errorf(PathConflict, "the file at \"%s\" is not a directory",
			path)
	}
	result := make([]*Node, len(d.Children))
	for i, child := range d.Children {
		result[i], ok = h.Fs[join(path, child)]
		if !ok {
			return nil, errorf(Internal, "could not find node for \"%s\" while "+
				"listing \"%s\"", join(path, child), path)
		}
	}
	return result, nil
}

// Glob beturns a list of nodes that match 'pattern'.
func (h *HashTree) Glob(pattern string) ([]*Node, error) {
	h.init()
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	pattern = clean(pattern)

	var res []*Node
	for p, node := range h.Fs {
		matched, err := pathlib.Match(pattern, p)
		if err != nil {
			if err == pathlib.ErrBadPattern {
				return nil, errorf(MalformedGlob, "glob \"%s\" is malformed", pattern)
			}
			return nil, err
		}
		if matched {
			res = append(res, node)
		}
	}
	return res, nil
}

// mergeNode merges the node at 'path' from 'from' into 'h'. The return value
// 's' is the number of bytes added to the node at 'path' (the size increase).
func (h *HashTree) mergeNode(path string, from Interface) (s int64, err error) {
	if h.Fs == nil {
		h.Fs = map[string]*Node{}
	}

	// Fetch the nodes, and return error if one can't be fetched
	fromNode, err := from.Get(path)
	if err != nil && Code(err) != PathNotFound {
		return 0, err
	}
	toNode, err := h.Get(path)
	if err != nil && Code(err) != PathNotFound {
		return 0, err
	}

	// Merge 'fromNode' into 'toNode' ('fromNode' will always be defined, because
	// mergeNode() traverses the 'from' tree)
	var sizeDelta int64
	if fromNode.DirNode != nil {
		if toNode != nil && toNode.DirNode == nil {
			return 0, errorf(PathConflict, "node at \"%s\" is a directory in the "+
				"target HashTree, but not in the source", path)
		}
		// Create empty directory in 'to' if none exists
		if toNode == nil {
			toNode = &Node{
				Name:    base(path),
				Size:    0,
				DirNode: &DirectoryNode{},
			}
			h.Fs[path] = toNode
		}

		// Merge files in 'from' into 'to' (including adding them if they're new)
		for _, child := range fromNode.DirNode.Children {
			if s, err := h.mergeNode(join(path, child), from); err == nil {
				insertStr(&toNode.DirNode.Children, child)
				sizeDelta += s
			} else {
				return 0, err
			}
		}
	} else if fromNode.FileNode != nil {
		if toNode != nil && toNode.FileNode == nil {
			return 0, errorf(PathConflict, "node at \"%s\" is a regular file in the "+
				"source HashTree.Interface, but not in the target", path)
		}
		// Create empty file in 'to' if none exists
		if toNode == nil {
			toNode = &Node{
				Name:     base(path),
				FileNode: &FileNode{},
			}
			h.Fs[path] = toNode
		}

		// Append new blocks
		toNode.FileNode.BlockRefs = append(
			toNode.FileNode.BlockRefs, fromNode.FileNode.BlockRefs...)

		// Compute size growth of node
		for _, br := range fromNode.FileNode.BlockRefs {
			sizeDelta += int64(br.Range.Upper - br.Range.Lower)
		}
	} else {
		return 0, errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}
	toNode.Size += sizeDelta
	h.updateHash(path)
	return sizeDelta, nil
}

// Merge merges the HashTrees in 'trees' into 'h'. The result is nil if no
// errors are encountered while merging any tree, or else a new error e, where:
// - Code(e) is the error code of the first error encountered
// - e.Error() contains the error messages of the first 10 errors encountered
func (h *HashTree) Merge(trees []Interface) error {
	h.init()

	errors := []error{}
	for _, tree := range trees {
		if _, err := tree.Get("/"); Code(err) == PathNotFound {
			continue // No work necessary to merge blank tree
		}
		_, err := h.mergeNode("", tree) // Empty string is internal repr of "/"
		if err != nil && len(errors) < 10 {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		msgs := make([]string, len(errors))
		for i := 0; i < len(errors); i++ {
			msgs[i] = errors[i].Error()
		}
		return errorf(Code(errors[0]),
			"encountered in merge: "+strings.Join(msgs, "; "))
	}
	return nil
}
