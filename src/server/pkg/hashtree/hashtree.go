package hashtree

import (
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

// hashtree is an implementation of the HashTree and OpenHashTree interfaces.
// It's intended to describe the state of a single commit C, in a repo R.
type hashtree struct {
	// fs (short for files) maps the path of each file F in the repo R to a
	// protobuf message describing F. It's equivalent to HashTree.Fs.
	fs map[string]*NodeProto

	// changed maps a path P to 'true' if P or one of its children has been
	// modified in 'fs', and its hash needs to be updated.
	changed map[string]bool
}

// toProto converts 'h' to a HashTree proto message. This is not public; it's
// a helper function for Serialize() and is also used for testing (to test
// whether e.g. a failed PutFile call modifies 'h'), so it must not modify 'h'.
func (h *hashtree) toProto() *HashTreeProto {
	return &HashTreeProto{
		Fs:      h.fs,
		Version: 1,
	}
}

// fromProto creates a hashtree struct from a HashTreeProto (used by
// Deserialize())
func fromProto(htproto *HashTreeProto) (*hashtree, error) {
	if htproto.Version != 1 {
		return nil, errorf(Unsupported, "unsupported HashTreeProto "+
			"version %d", htproto.Version)
	}
	res := &hashtree{
		fs:      htproto.Fs,
		changed: make(map[string]bool),
	}
	return res, nil
}

// Serialize serializes a HashTree so that it can be persisted. Also see
// Deserialize(bytes).
func (h *hashtree) Serialize() ([]byte, error) {
	return proto.Marshal(h.toProto())
}

// Deserialize deserializes a hash tree so that it can be read or modified.
func Deserialize(serialized []byte) (HashTree, error) {
	h := &HashTreeProto{}
	proto.Unmarshal(serialized, h)
	return fromProto(h)
}

// NewHashTree creates a new hash tree implementing Interface.
func NewHashTree() HashTree {
	return &hashtree{
		fs:      make(map[string]*NodeProto),
		changed: make(map[string]bool),
	}
}

// canonicalize updates the hash and size of the node N at 'path'. If N is a
// directory canonicalize will also update the hash of all of N's children
// recursively. Thus, h.canonicalize("/") will update all hash and subtree_size
// fields in h, making the whole tree consistent.
func (h *hashtree) canonicalize(path string) error {
	path = clean(path)
	if !h.changed[path] {
		return nil // Node is already canonical
	}
	n, ok := h.fs[path]
	if !ok {
		return errorf(Internal, "no node at \"%s\"; cannot canonicalize", path)
	}

	// Compute hash and size of 'n'
	hash := sha256.New()
	size := int64(0)
	switch n.nodetype() {
	case directory:
		// Compute n.Hash by concatenating name + hash of all children of n.DirNode
		// Note that PutFile keeps n.DirNode.Children sorted, so the order is stable
		for _, child := range n.DirNode.Children {
			childpath := join(path, child)
			if err := h.canonicalize(childpath); err != nil {
				return err
			}
			n, ok := h.fs[childpath]
			if !ok {
				return errorf(Internal, "could not find node for \"%s\" while "+
					"updating hash of \"%s\"", join(path, child), path)
			}
			// append child.Name and child.Hash to b
			if _, err := hash.Write([]byte(fmt.Sprintf("%s:%s:", n.Name, n.Hash))); err != nil {
				return errorf(Internal, "error updating hash of file at \"%s\": %s",
					path, err)
			}
			size += n.SubtreeSize
		}
	case file:
		// Compute n.Hash by concatenating all BlockRef hashes in n.FileNode
		for _, blockRef := range n.FileNode.BlockRefs {
			_, err := hash.Write([]byte(fmt.Sprintf("%s:%d:%d:",
				blockRef.Block.Hash, blockRef.Range.Lower, blockRef.Range.Upper)))
			if err != nil {
				return errorf(Internal, "error updating hash of dir at \"%s\": %s",
					path, err)
			}
			size += int64(blockRef.Range.Upper - blockRef.Range.Lower)
		}
	default:
		return errorf(Internal,
			"malformed node at \"%s\" is neither a file nor a directory", path)
	}

	// Update hash and size of 'n'
	n.Hash = hash.Sum(nil)
	n.SubtreeSize = size
	delete(h.changed, path)
	return nil
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
func (h *hashtree) visit(path string, update updateFn) error {
	for path != "" {
		parent, child := split(path)
		pnode, ok := h.fs[parent]
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

// removeFromMap removes the node at 'path' from h.fs if it's present, along
// with all of its children, recursively.
//
// This will not update the hash of any parent of 'path'. This helps us avoid
// updating the hash of path's parents unnecessarily; if 'path' is a directory
// with e.g. 10k children, updating the parents' hashes after all files have
// been removed from h.fs (instead of updating all parents' hashesafter
// removing each file) may save substantial time.
func (h *hashtree) removeFromMap(path string) error {
	n, ok := h.fs[path]
	if !ok {
		return nil
	}

	switch n.nodetype() {
	case file:
		delete(h.fs, path)
	case directory:
		for _, child := range n.DirNode.Children {
			if err := h.removeFromMap(join(path, child)); err != nil {
				return err
			}
		}
		delete(h.fs, path)
	case unrecognized:
		return errorf(Internal,
			"malformed node at \"%s\": it's neither a file nor a directory", path)
	}
	return nil
}

// PutFile appends data to a file (and creates the file if it doesn't exist).
func (h *hashtree) PutFile(path string, blockRefs []*pfs.BlockRef) error {
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Get/Create file node to which we'll append 'blockRefs'
	node, ok := h.fs[path]
	if !ok {
		node = &NodeProto{
			Name:     base(path),
			FileNode: &FileNodeProto{},
		}
		h.fs[path] = node
	} else if node.nodetype() != file {
		return errorf(PathConflict, "could not put file at \"%s\"; a node of "+
			"type %s is already there", path, node.nodetype().tostring())
	}

	// Append new blocks
	node.FileNode.BlockRefs = append(node.FileNode.BlockRefs, blockRefs...)
	h.changed[path] = true

	// Compute size growth of node (i.e. amount of data we're appending)
	var sizeGrowth int64
	for _, blockRef := range blockRefs {
		sizeGrowth += int64(blockRef.Range.Upper - blockRef.Range.Lower)
	}
	node.SubtreeSize += sizeGrowth

	// Add 'path' to parent & update hashes back to root
	if err := h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			node = &NodeProto{
				Name:        base(parent),
				SubtreeSize: 0,
				DirNode:     &DirectoryNodeProto{},
			}
			h.fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		node.SubtreeSize += sizeGrowth
		h.changed[parent] = true
		return nil
	}); err != nil {
		return err
	}
	return h.canonicalize("/")
}

// PutDir creates a directory (or does nothing if one exists).
func (h *hashtree) PutDir(path string) error {
	path = clean(path)

	// Detect any path conflicts before modifying 'h'
	if err := h.visit(path, nop); err != nil {
		return err
	}

	// Create orphaned directory at 'path' (or end early if a directory is there)
	if node, ok := h.fs[path]; ok {
		if node.nodetype() == directory {
			return nil
		} else if node.nodetype() != none {
			return errorf(PathConflict, "could not create directory at \"%s\"; a "+
				"file of type %s is already there", path, node.nodetype().tostring())
		}
	}
	h.fs[path] = &NodeProto{
		Name:    base(path),
		DirNode: &DirectoryNodeProto{},
	}
	h.changed[path] = true

	// Add 'path' to parent & update hashes back to root
	if err := h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			node = &NodeProto{
				Name:    base(parent),
				DirNode: &DirectoryNodeProto{},
			}
			h.fs[parent] = node
		}
		insertStr(&node.DirNode.Children, child)
		h.changed[parent] = true
		return nil
	}); err != nil {
		return err
	}
	return h.canonicalize("/")
}

// DeleteFile deletes a regular file or directory (along with its children).
func (h *hashtree) DeleteFile(path string) error {
	path = clean(path)

	// Remove 'path' and all nodes underneath it from h.fs
	node, ok := h.fs[path]
	if !ok {
		return errorf(PathNotFound, "no file at \"%s\"", path)
	}
	h.removeFromMap(path) // Deletes children recursively

	// Remove 'path' from its parent directory
	parent, child := split(path)
	node, ok = h.fs[parent]
	if !ok {
		return errorf(Internal, "delete discovered orphaned file \"%s\"", path)
	}
	if node.DirNode == nil {
		return errorf(Internal, "node at \"%s\" is a file, but \"%s\" exists "+
			"under it (likely an uncaught PathConflict in prior PutFile or Merge)")
	}
	if !removeStr(&node.DirNode.Children, child) {
		return errorf(Internal, "parent of \"%s\" does not contain it", path)
	}
	// Update hashes back to root
	if err := h.visit(path, func(node *NodeProto, parent, child string) error {
		if node == nil {
			return errorf(Internal,
				"encountered orphaned file \"%s\" while deleting \"%s\"", path,
				join(parent, child))
		}
		h.changed[parent] = true
		return nil
	}); err != nil {
		return err
	}
	return h.canonicalize("/")
}

// Get retrieves the contents of a file.
func (h *hashtree) Get(path string) (*NodeProto, error) {
	path = clean(path)

	node, ok := h.fs[path]
	if !ok {
		return nil, errorf(PathNotFound, "no node at \"%s\"", path)
	}
	return node, nil
}

// List retrieves the list of files and subdirectories of the directory at
// 'path'.
func (h *hashtree) List(path string) ([]*NodeProto, error) {
	path = clean(path)

	node, err := h.Get(path)
	if err != nil {
		return nil, err
	}
	d := node.DirNode
	if d == nil {
		return nil, errorf(PathConflict, "the file at \"%s\" is not a directory",
			path)
	}
	var ok bool
	result := make([]*NodeProto, len(d.Children))
	for i, child := range d.Children {
		result[i], ok = h.fs[join(path, child)]
		if !ok {
			return nil, errorf(Internal, "could not find node for the child \"%s\" "+
				"while listing \"%s\"", join(path, child), path)
		}
	}
	return result, nil
}

// Glob returns a list of files and directories that match 'pattern'.
func (h *hashtree) Glob(pattern string) ([]*NodeProto, error) {
	// "*" should be an allowed pattern, but our paths always start with "/", so
	// modify the pattern to fit our path structure.
	pattern = clean(pattern)

	var res []*NodeProto
	for p, node := range h.fs {
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

// mergeNode merges the node at 'path' from the trees in 'srcs' into 'h'.
func (h *hashtree) mergeNode(path string, srcs []HashTree) error {
	path = clean(path)
	// Get the node at path in 'h' and determine its type (i.e. file, dir)
	destNode, err := h.Get(path)
	if err != nil && Code(err) != PathNotFound {
		return err
	}
	if destNode.nodetype() == unrecognized {
		return errorf(Internal, "malformed node at \"%s\" in destination "+
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
			return err
		}
		if n.nodetype() == none {
			continue
		}
		if pathtype == none {
			pathtype = n.nodetype()
		} else if pathtype != n.nodetype() {
			return errorf(PathConflict, "could not merge path \"%s\" which is "+
				"not consistently a file/directory in the hashtrees being merged")
		}
		switch n.nodetype() {
		case directory:
			// Create destination directory if none exists
			if destNode == nil {
				destNode = &NodeProto{
					Name:    base(path),
					DirNode: &DirectoryNodeProto{},
				}
				h.fs[path] = destNode
			}
			// Instead of merging here, we build a reverse-index and merge below
			for _, c := range n.DirNode.Children {
				childrenToTrees[c] = append(childrenToTrees[c], src)
			}
		case file:
			// Create destination file if none exists
			if destNode == nil {
				destNode = &NodeProto{
					Name:     base(path),
					FileNode: &FileNodeProto{},
				}
				h.fs[path] = destNode
			}
			// Append new blocks
			destNode.FileNode.BlockRefs = append(destNode.FileNode.BlockRefs,
				n.FileNode.BlockRefs...)
			sizeDelta += n.SubtreeSize
		default:
			return errorf(Internal, "malformed node at \"%s\" in source "+
				"hashtree is neither a file nor a directory", path)
		}
	}

	// If this is a directory, go back and merge all children encountered above
	if pathtype == directory {
		// Merge all children (collected in childrenToTrees)
		for c, cSrcs := range childrenToTrees {
			err := h.mergeNode(join(path, c), cSrcs)
			if err != nil {
				return err
			}
			insertStr(&destNode.DirNode.Children, c)
		}
	}
	h.changed[path] = true
	return nil
}

// Merge merges the HashTrees in 'trees' into 'h'. The result is nil if no
// errors are encountered while merging any tree, or else a new error e, where:
// - Code(e) is the error code of the first error encountered
// - e.Error() contains the error messages of the first 10 errors encountered
func (h *hashtree) Merge(trees []HashTree) error {
	b, err := h.Serialize()
	if err != nil {
		return errorf(Internal, "could not serialize hashtree before merge: %s", err)
	}
	if err = h.mergeNode("/", trees); err != nil {
		htInterfaceTmp, deserializeErr := Deserialize(b)
		if deserializeErr != nil {
			return errorf(Internal, "could not deserialize hashtree (due to \"%s\") "+
				"after merge error: %s", deserializeErr, err)
		}
		if htTmp, ok := htInterfaceTmp.(*hashtree); ok {
			*h = *htTmp
		} else {
			return errorf(Internal, "could not convert deserialized hash tree after "+
				"merge error: %s", err)
		}
		return err
	}
	return h.canonicalize("/")
}
