package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	pathlib "path"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type nodetype uint8

const (
	none         nodetype = iota // No file is present at this point in the tree
	directory                    // The file at this point in the tree is a directory
	file                         // ... is a regular file
	unrecognized                 // ... is an an unknown type
)

func (n *NodeProto) nodetype() nodetype {
	switch {
	case n == nil:
		return none
	case n.DirNode != nil:
		return directory
	case n.FileNode != nil:
		return file
	default:
		return unrecognized
	}
}

func (n nodetype) tostring() string {
	switch n {
	case none:
		return "none"
	case directory:
		return "directory"
	case file:
		return "file"
	default:
		return "unknown"
	}
}

// updateHash updates the hash of the node N at 'path'. If this changes N's
// hash, that will render the hash of N's parent (if any) invalid, and this
// must be called for all parents of 'path' back to the root.
func (h *HashTreeProto) updateHash(path string) error {
	n, ok := h.Fs[path]
	if !ok {
		return errorf(Internal, "Could not find node \"%s\" to update hash", path)
	}

	// Compute hash of 'n'
	var b bytes.Buffer
	switch n.nodetype() {
	case directory:
		// PutFile keeps n.DirNodeProto.Children sorted, so the order is stable
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
	case file:
		for _, blockRef := range n.FileNode.BlockRefs {
			_, err := b.WriteString(fmt.Sprintf("%s:%d:%d:",
				blockRef.Block.Hash, blockRef.Range.Lower, blockRef.Range.Upper))
			if err != nil {
				return errorf(Internal, "error updating hash of dir at \"%s\": %s",
					path, err)
			}
		}
	case unrecognized:
		return errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}

	// Update hash of 'n'
	cksum := sha256.Sum256(b.Bytes())
	n.Hash = cksum[:]
	return nil
}

func (h *HashTreeProto) init() {
	if h.Fs == nil {
		h.Fs = make(map[string]*NodeProto)
	}
	if h.Version == 0 {
		h.Version = 1
	}
}

// updateFn is used by 'visit'. The first parameter is the node being visited,
// the second parameter is the path of that node, and the third parameter is the
// child of that node from the 'path' argument to 'visit'.
//
// The *NodeProto argument is guaranteed to have DirNode set (if it's not nil)--visit
// returns a 'PathConflict' error otherwise.
type updateFn func(*NodeProto, string, string) error

// This can be passed to visit() to detect PathConflict errors early
func nop(*NodeProto, string, string) error {
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
func (h *HashTreeProto) visit(path string, update updateFn) error {
	for path != "" {
		parent, child := split(path)
		pnode, ok := h.Fs[parent]
		if ok && pnode.nodetype() != directory {
			return errorf(PathConflict, "attempted to visit \"%s\", but it's not a "+
				"directory", path)
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
func (h *HashTreeProto) removeFromMap(path string) error {
	n, ok := h.Fs[path]
	if !ok {
		return nil
	}

	switch n.nodetype() {
	case file:
		delete(h.Fs, path)
	case directory:
		for _, child := range n.DirNode.Children {
			if err := h.removeFromMap(join(path, child)); err != nil {
				return err
			}
		}
		delete(h.Fs, path)
	case unrecognized:
		return errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}
	return nil
}

// PutFile inserts a file into the hierarchy
func (h *HashTreeProto) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	h.init()
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Get/Create file node to which we'll append 'blockRefs'
	node, ok := h.Fs[path]
	if ok {
		if node.nodetype() != file {
			return errorf(PathConflict, "could not put file at \"%s\"; a node of "+
				"type %s is already there", path, node.nodetype().tostring())
		}
	} else {
		node = &NodeProto{
			Name:     base(path),
			FileNode: &FileNodeProto{},
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
	node.SubtreeSize += sizeGrowth

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			node = &NodeProto{
				Name:        base(parent),
				SubtreeSize: 0,
				DirNode:     &DirectoryNodeProto{},
			}
			h.Fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		node.SubtreeSize += sizeGrowth
		h.updateHash(parent)
		return nil
	})
}

// PutDir inserts an empty directory into the hierarchy
func (h *HashTreeProto) PutDir(path string) error {
	h.init()
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Create orphaned directory at 'path' (or end early if a directory is there)
	if node, ok := h.Fs[path]; ok {
		if node.nodetype() == directory {
			return nil
		} else if node.nodetype() != none {
			return errorf(PathConflict, "could not create directory at \"%s\"; a "+
				"file of type %s is already there", path, node.nodetype().tostring())
		}
	}
	h.Fs[path] = &NodeProto{
		Name:    base(path),
		DirNode: &DirectoryNodeProto{},
	}
	h.updateHash(path)

	// Add 'path' to parent & update hashes back to root
	return h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			node = &NodeProto{
				Name:    base(parent),
				DirNode: &DirectoryNodeProto{},
			}
			h.Fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		h.updateHash(parent)
		return nil
	})
}

// DeleteFile deletes the file at 'path', and all children recursively if 'path'
// is a subdirectory
func (h *HashTreeProto) DeleteFile(path string) error {
	h.init()
	path = clean(path)

	// Remove 'path' from h.Fs
	node, ok := h.Fs[path]
	if !ok {
		return errorf(PathNotFound, "no file at \"%s\"", path)
	}
	size := node.SubtreeSize
	h.removeFromMap(path) // Deletes children recursively

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
	return h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			return errorf(Internal,
				"encountered orphaned file \"%s\" while deleting \"%s\"", path,
				join(parent, child))
		}
		node.SubtreeSize -= size
		h.updateHash(parent)
		return nil
	})
}

// Get returns the node associated with the path
func (h *HashTreeProto) Get(path string) (*NodeProto, error) {
	h.init()
	path = clean(path)

	node, ok := h.Fs[path]
	if !ok {
		return nil, errorf(PathNotFound, "no node at \"%s\"", path)
	}
	return node, nil
}

// List returns the NodeProtos corresponding to the files and directories under
// 'path'
func (h *HashTreeProto) List(path string) ([]*NodeProto, error) {
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
	result := make([]*NodeProto, len(d.Children))
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
func (h *HashTreeProto) Glob(pattern string) ([]*NodeProto, error) {
	h.init()
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	pattern = clean(pattern)

	var res []*NodeProto
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
func (h *HashTreeProto) mergeNode(path string, srcs []HashTree) (sz int64, err error) {
	// Get the node at path in 'h' and determine its type (i.e. file, dir)
	destNode, err := h.Get(path)
	if err != nil && Code(err) != PathNotFound {
		return 0, err
	}
	if destNode.nodetype() == unrecognized {
		return 0, errorf(Internal, "malformed node at \"%s\" in destination "+
			"hashtree is neither a file nor a directory", path)
	}

	// Get node at 'path' in all 'srcs'. All such nodes must have the same type as
	// each other and the same type as 'destNode'
	pathtype := destNode.nodetype() // All nodes in 'srcs' must have same type
	// childrenToTrees is a reverse index from [child of 'path'] to [trees that
	// contain it].
	// - childrenToTrees will only be used if 'path' is a directory in all
	//   'srcs' (but it's convenient to build it here, before we've checked them
	//   all)
	// - We need to group trees by common children, so that children present in
	//   multiple 'srcNodes' are only merged once, and we only recompute the hash
	//   for that child once
	// - if every srcNode has a unique file /foo/shard-xxxxx (00000 to 99999)
	//   then we'd call mergeNode("/foo") 100k times, once for each tree in
	//   'srcs', while running mergeNode("/").
	// - We also can't pass all of 'srcs' to mergeNode("/foo"), as otherwise
	//   mergeNode("/foo/shard-xxxxx") will have to filter through all 100k trees
	//   for each shard-xxxxx (only one of which contains the file being merged),
	//   and we'd have an O(n^2) algorithm; too slow when merging 100k trees)
	childrenToTrees := make(map[string][]HashTree)
	// Amount of data being added to node at 'path' in 'h'
	sizeDelta := int64(0)
	for _, src := range srcs {
		n, err := src.Get(path)
		if err != nil && Code(err) != PathNotFound {
			return 0, err
		}
		if n.nodetype() == unrecognized {
			return 0, errorf(Internal, "malformed node at \"%s\" in source "+
				"hashtree is neither a file nor a directory", path)
		}
		if n.nodetype() != none {
			if pathtype == none {
				pathtype = n.nodetype()
			} else if pathtype != n.nodetype() {
				return 0, errorf(PathConflict, "could not merge path \"%s\" which is "+
					"not consistently a file/directory in the hashtrees being merged")
			}
			if n.nodetype() == directory {
				// Create destination directory if none exists
				if destNode == nil {
					destNode = &NodeProto{
						Name:        base(path),
						SubtreeSize: 0,
						DirNode:     &DirectoryNodeProto{},
					}
					h.Fs[path] = destNode
				}
				// Instead of merging here, we build a reverse-index and merge below
				for _, c := range n.DirNode.Children {
					childrenToTrees[c] = append(childrenToTrees[c], src)
				}
			} else if n.nodetype() == file {
				// Create destination file if none exists
				if destNode == nil {
					destNode = &NodeProto{
						Name:        base(path),
						SubtreeSize: 0,
						FileNode:    &FileNodeProto{},
					}
					h.Fs[path] = destNode
				}
				// Append new blocks
				destNode.FileNode.BlockRefs = append(destNode.FileNode.BlockRefs,
					n.FileNode.BlockRefs...)
				sizeDelta += n.SubtreeSize
			}
		}
	}

	// If this is a directory, go back and merge all children encountered above
	if pathtype == directory {
		// Merge all children (collected in childrenToTrees)
		for c, cSrcs := range childrenToTrees {
			sz, err := h.mergeNode(join(path, c), cSrcs)
			if err != nil {
				return 0, err
			}
			insertStr(&destNode.DirNode.Children, c)
			sizeDelta += sz
		}
	}
	h.updateHash(path)
	destNode.SubtreeSize += sizeDelta
	return sizeDelta, nil
}

// Merge merges the HashTrees in 'trees' into 'h'. The result is nil if no
// errors are encountered while merging any tree, or else a new error e, where:
// - Code(e) is the error code of the first error encountered
// - e.Error() contains the error messages of the first 10 errors encountered
func (h *HashTreeProto) Merge(trees []HashTree) error {
	h.init()
	backup := proto.Clone(h)
	_, err := h.mergeNode("", trees) // Empty string is internal repr of "/"
	if err != nil {
		h.Reset()
		proto.Merge(h, backup)
		return err
	}
	return nil
}
