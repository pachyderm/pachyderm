package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	bolt "github.com/coreos/bbolt"
	globlib "github.com/gobwas/glob"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	FsBucket      = "fs"
	ChangedBucket = "changed"
)

var (
	changedVal     = []byte{1}
	emptyStringKey = []byte{0}
)

func fs(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(FsBucket))
}

func changed(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(ChangedBucket))
}

type dbHashTree struct {
	*bolt.DB
}

func s(b []byte) string {
	if bytes.Equal(b, emptyStringKey) {
		return ""
	}
	return string(b)
}

func b(s string) []byte {
	if s == "" {
		return emptyStringKey
	}
	return []byte(s)
}

func dbFile() string {
	return fmt.Sprintf("/tmp/db/%s", uuid.NewWithoutDashes())
}

func NewDBHashTree() (OpenHashTree, error) {
	return newDBHashTree(dbFile())
}

func newDBHashTree(file string) (OpenHashTree, error) {
	db, err := bolt.Open(file, 0666, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Batch(func(tx *bolt.Tx) error {
		for _, bucket := range []string{FsBucket, ChangedBucket} {
			if _, err := tx.CreateBucketIfNotExists(b(bucket)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &dbHashTree{db}, nil
}

// HashTree interface
func (h *dbHashTree) Open() OpenHashTree {
	// Already open
	return h
}

func (h *dbHashTree) get(tx *bolt.Tx, path string) (*NodeProto, error) {
	node := &NodeProto{}
	data := fs(tx).Get(b(path))
	if data == nil {
		return nil, errorf(PathNotFound, "file \"%s\" not found", path)
	}
	if err := node.Unmarshal(data); err != nil {
		return nil, err
	}
	return node, nil
}

func (h *dbHashTree) Get(path string) (*NodeProto, error) {
	path = clean(path)
	var node *NodeProto
	if err := h.View(func(tx *bolt.Tx) error {
		var err error
		node, err = h.get(tx, path)
		return err
	}); err != nil {
		return nil, err
	}
	return node, nil
}

// List retrieves the list of files and subdirectories of the directory at
// 'path'.
func (h *dbHashTree) List(path string) ([]*NodeProto, error) {
	path = clean(path)
	var result []*NodeProto
	if err := h.View(func(tx *bolt.Tx) error {
		c := fs(tx).Cursor()
		for k, v := c.Seek(b(path)); k != nil && strings.HasPrefix(s(k), path); k, v = c.Next() {
			trimmed := strings.TrimPrefix(s(k), path)
			if trimmed == "" || strings.Count(trimmed, "/") > 1 {
				// don't return the path itself or the children of children
				// TODO seeking to the directory will greatly improve performance here
				continue
			}
			node := &NodeProto{}
			if err := node.Unmarshal(v); err != nil {
				return err
			}
			result = append(result, node)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *dbHashTree) glob(tx *bolt.Tx, pattern string) (map[string]*NodeProto, error) {
	if !isGlob(pattern) {
		node, err := h.Get(pattern)
		if err != nil {
			return nil, err
		}
		return map[string]*NodeProto{clean(pattern): node}, nil
	}

	pattern = clean(pattern)
	g, err := globlib.Compile(pattern, '/')
	if err != nil {
		return nil, errorf(MalformedGlob, err.Error())
	}
	res := make(map[string]*NodeProto)
	c := fs(tx).Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if g.Match(s(k)) {
			node := &NodeProto{}
			if node.Unmarshal(v); err != nil {
				return nil, err
			}
			res[s(k)] = node
		}
	}
	return res, nil
}

func (h *dbHashTree) Glob(pattern string) (map[string]*NodeProto, error) {
	var result map[string]*NodeProto
	if err := h.View(func(tx *bolt.Tx) error {
		var err error
		result, err = h.glob(tx, pattern)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *dbHashTree) FSSize() int64 {
	rootNode, err := h.Get("/")
	if err != nil {
		return 0
	}
	return rootNode.SubtreeSize
}

func (h *dbHashTree) Walk(path string, f func(path string, node *NodeProto) error) error {
	path = clean(path)
	return h.View(func(tx *bolt.Tx) error {
		c := fs(tx).Cursor()
		for k, v := c.First(); k != nil && strings.HasPrefix(s(k), path); k, v = c.Next() {
			node := &NodeProto{}
			if err := node.Unmarshal(v); err != nil {
				return err
			}
			if err := f(s(k), node); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *dbHashTree) Diff(oldHashTree HashTree, newPath string, oldPath string, recursiveDepth int64, f func(path string, node *NodeProto, new bool) error) error {
	return diff(h, oldHashTree, newPath, oldPath, recursiveDepth, f)
}

func (h *dbHashTree) GetOpen(path string) (*OpenNode, error) {
	node, err := h.Get(path)
	if err != nil {
		return nil, err
	}
	return &OpenNode{
		Name:     node.Name,
		Size:     node.SubtreeSize,
		FileNode: node.FileNode,
		DirNode:  node.DirNode,
	}, nil
}

func (h *dbHashTree) put(tx *bolt.Tx, path string, node *NodeProto) error {
	data, err := node.Marshal()
	if err != nil {
		return err
	}
	if err := changed(tx).Put(b(path), changedVal); err != nil {
		return fmt.Errorf("error putting \"%s\": %v", path, err)
	}
	return fs(tx).Put(b(path), data)
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
// This is useful for propagating changes to size upwards.
func (h *dbHashTree) visit(tx *bolt.Tx, path string, update updateFn) error {
	for path != "" {
		parent, child := split(path)
		pnode, err := h.get(tx, parent)
		if err != nil && Code(err) != PathNotFound {
			return err
		}
		if pnode != nil && pnode.nodetype() != directory {
			return errorf(PathConflict, "attempted to visit \"%s\", but it's not a "+
				"directory", path)
		}
		if pnode == nil {
			pnode = &NodeProto{}
		}
		if err := update(pnode, parent, child); err != nil {
			return err
		}
		if err := h.put(tx, parent, pnode); err != nil {
			return err
		}
		path = parent
	}
	return nil
}

func (h *dbHashTree) PutFile(path string, objects []*pfs.Object, size int64) error {
	return h.putFile(path, objects, nil, size)
}

func (h *dbHashTree) PutFileOverwrite(path string, objects []*pfs.Object, overwriteIndex *pfs.OverwriteIndex, sizeDelta int64) error {
	return h.putFile(path, objects, overwriteIndex, sizeDelta)
}

func (h *dbHashTree) putFile(path string, objects []*pfs.Object, overwriteIndex *pfs.OverwriteIndex, sizeDelta int64) error {
	path = clean(path)
	return h.Batch(func(tx *bolt.Tx) error {
		node, err := h.get(tx, path)
		if err != nil && Code(err) != PathNotFound {
			return err
		}
		if node == nil {
			node = &NodeProto{
				Name:     base(path),
				FileNode: &FileNodeProto{},
			}
		}
		// Append new objects.  Remove existing objects if overwriting.
		if overwriteIndex != nil && overwriteIndex.Index <= int64(len(node.FileNode.Objects)) {
			node.FileNode.Objects = node.FileNode.Objects[:overwriteIndex.Index]
		}
		node.SubtreeSize += sizeDelta
		node.FileNode.Objects = append(node.FileNode.Objects, objects...)
		// Put the node
		if err := h.put(tx, path, node); err != nil {
			return err
		}
		return h.visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.Name == "" {
				node.Name = base(parent)
				node.DirNode = &DirectoryNodeProto{}
			}
			insertStr(&node.DirNode.Children, child)
			node.SubtreeSize += sizeDelta
			return nil
		})
	})
}

func (h *dbHashTree) PutDir(path string) error {
	path = clean(path)
	return h.Batch(func(tx *bolt.Tx) error {
		node, err := h.get(tx, path)
		if err != nil && Code(err) != PathNotFound {
			return err
		}
		if node != nil {
			if node.nodetype() == directory {
				return nil
			} else if node.nodetype() != none {
				return errorf(PathConflict, "could not create directory at \"%s\"; a "+
					"file of type %s is already there", path, node.nodetype().tostring())
			}
		}
		node = &NodeProto{
			Name:    base(path),
			DirNode: &DirectoryNodeProto{},
		}
		if err := h.put(tx, path, node); err != nil {
			return err
		}
		return h.visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.Name == "" {
				node.Name = base(parent)
				node.DirNode = &DirectoryNodeProto{}
			}
			insertStr(&node.DirNode.Children, child)
			return nil
		})
	})
}

// deleteDir deletes a directory and all the children under it
func (h *dbHashTree) deleteDir(tx *bolt.Tx, path string) error {
	c := fs(tx).Cursor()
	for k, _ := c.Seek(b(path)); k != nil && strings.HasPrefix(s(k), path); k, _ = c.Next() {
		if err := c.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (h *dbHashTree) DeleteFile(path string) error {
	path = clean(path)

	// Delete root means delete all files
	if path == "" {
		path = "*"
	}
	return h.Batch(func(tx *bolt.Tx) error {
		paths, err := h.glob(tx, path)
		// Deleting a non-existent file should be a no-op
		if err != nil && Code(err) != PathNotFound {
			return err
		}
		for path := range paths {
			// Check if the file has been deleted already
			if _, err := h.get(tx, path); err != nil && Code(err) == PathNotFound {
				continue
			}
			// Remove 'path' and all nodes underneath it from h.fs
			if err := h.deleteDir(tx, path); err != nil {
				return err
			}
			size := paths[path].SubtreeSize
			// Remove 'path' from its parent directory
			parent, child := split(path)
			pnode, err := h.get(tx, parent)
			if err != nil {
				if Code(err) == PathNotFound {
					return errorf(Internal, "delete discovered orphaned file \"%s\"", path)
				}
				return err
			}
			if pnode.DirNode == nil {
				return errorf(Internal, "file at \"%s\" is a regular-file, but \"%s\" already exists "+
					"under it (likely an uncaught PathConflict in prior PutFile or Merge)", path, pnode.DirNode)
			}
			if !removeStr(&pnode.DirNode.Children, child) {
				return errorf(Internal, "parent of \"%s\" does not contain it", path)
			}
			// Mark nodes as 'changed' back to root
			if err := h.visit(tx, path, func(node *NodeProto, parent, child string) error {
				if node.Name == "" {
					return errorf(Internal,
						"encountered orphaned file \"%s\" while deleting \"%s\"", path,
						join(parent, child))
				}
				node.SubtreeSize -= size
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *dbHashTree) mergeNode(tx *bolt.Tx, path string, srcs []HashTree) (int64, error) {
	path = clean(path)
	// Get the node at path in 'h' and determine its type (i.e. file, dir)
	destNode, err := h.get(tx, path)
	if err != nil && Code(err) != PathNotFound {
		return 0, err
	}
	if destNode == nil {
		destNode = &NodeProto{
			Name:        base(path),
			SubtreeSize: 0,
			Hash:        nil,
			FileNode:    nil,
			DirNode:     nil,
		}
		h.put(tx, path, destNode)
	} else if destNode.nodetype() == unrecognized {
		return 0, errorf(Internal, "malformed file at \"%s\" in destination "+
			"hashtree is neither a regular-file nor a directory", path)
	}
	sizeDelta := int64(0) // We return this to propagate file additions upwards

	// Get node at 'path' in all 'srcs'. All such nodes must have the same type as
	// each other and the same type as 'destNode'
	pathtype := destNode.nodetype() // All nodes in 'srcs' must have same type
	// childrenToTrees is a reverse index from [child of 'path'] to [trees that
	// contain it].
	// - childrenToTrees will only be used if 'path' is a directory in all
	//   'srcs' (but it's convenient to build it here, before we've checked them
	//   all)
	// - We need to group trees by common children, so that children present in
	//   multiple 'srcNodes' are only merged once
	// - if every srcNode has a unique file /foo/shard-xxxxx (00000 to 99999)
	//   then we'd call mergeNode("/foo") 100k times, once for each tree in
	//   'srcs', while running mergeNode("/").
	// - We also can't pass all of 'srcs' to mergeNode("/foo"), as otherwise
	//   mergeNode("/foo/shard-xxxxx") will have to filter through all 100k trees
	//   for each shard-xxxxx (only one of which contains the file being merged),
	//   and we'd have an O(n^2) algorithm; too slow when merging 100k trees)
	childrenToTrees := make(map[string][]HashTree)
	// Amount of data being added to node at 'path' in 'h'
	for _, src := range srcs {
		n, err := src.Get(path)
		if Code(err) == PathNotFound {
			continue
		}
		if pathtype == none {
			// 'h' is uninitialized at this path
			if n.nodetype() == directory {
				destNode.DirNode = &DirectoryNodeProto{}
			} else if n.nodetype() == file {
				destNode.FileNode = &FileNodeProto{}
			} else {
				return 0, errorf(Internal, "could not merge unrecognized file type at "+
					"\"%s\", which is neither a file nore a directory", path)
			}
			pathtype = n.nodetype()
		} else if pathtype != n.nodetype() {
			return sizeDelta, errorf(PathConflict, "could not merge path \"%s\" "+
				"which is a regular-file in some hashtrees and a directory in others", path)
		}
		switch n.nodetype() {
		case directory:
			// Instead of merging here, we build a reverse-index and merge below
			for _, c := range n.DirNode.Children {
				childrenToTrees[c] = append(childrenToTrees[c], src)
			}
		case file:
			// Append new objects, and update size of target node (since that can't be
			// done in canonicalize)
			destNode.FileNode.Objects = append(destNode.FileNode.Objects,
				n.FileNode.Objects...)
			sizeDelta += n.SubtreeSize
		default:
			return sizeDelta, errorf(Internal, "malformed file at \"%s\" in source "+
				"hashtree is neither a regular-file nor a directory", path)
		}
	}

	// If this is a directory, go back and merge all children encountered above
	if pathtype == directory {
		// Merge all children (collected in childrenToTrees)
		for c, cSrcs := range childrenToTrees {
			childSizeDelta, err := h.mergeNode(tx, join(path, c), cSrcs)
			if err != nil {
				return sizeDelta, err
			}
			sizeDelta += childSizeDelta
			insertStr(&destNode.DirNode.Children, c)
		}
	}
	// Update the size of destNode, and mark it changed
	destNode.SubtreeSize += sizeDelta
	h.put(tx, path, destNode)
	return sizeDelta, nil
}

func (h *dbHashTree) Merge(trees ...HashTree) error {
	// Skip empty trees
	var nonEmptyTrees []HashTree
	for _, tree := range trees {
		_, err := tree.Get("/")
		if err != nil {
			continue
		}
		nonEmptyTrees = append(nonEmptyTrees, tree)
	}
	if len(nonEmptyTrees) == 0 {
		return nil
	}
	return h.Batch(func(tx *bolt.Tx) error {
		_, err := h.mergeNode(tx, "/", nonEmptyTrees)
		return err
	})
}

func (h *dbHashTree) changed(tx *bolt.Tx, path string) bool {
	return changed(tx).Get(b(path)) != nil
}

func (h *dbHashTree) canonicalize(tx *bolt.Tx, path string) error {
	path = clean(path)
	if !h.changed(tx, path) {
		return nil // Node is already canonical
	}
	n, err := h.get(tx, path)
	if err != nil {
		if Code(err) == PathNotFound {
			return errorf(Internal, "file \"%s\" not found; cannot canonicalize", path)
		}
		return err
	}

	// Compute hash of 'n'
	hash := sha256.New()
	switch n.nodetype() {
	case directory:
		// Compute n.Hash by concatenating name + hash of all children of n.DirNode
		// Note that PutFile keeps n.DirNode.Children sorted, so the order is
		// stable.
		for _, child := range n.DirNode.Children {
			childpath := join(path, child)
			if err := h.canonicalize(tx, childpath); err != nil {
				return err
			}
			childnode, err := h.get(tx, childpath)
			if err != nil {
				if Code(err) == PathNotFound {
					return errorf(Internal, "could not find file for \"%s\" while "+
						"updating hash of \"%s\"", join(path, child), path)
				}
				return err
			}
			// append child.Name and child.Hash to b
			hash.Write([]byte(fmt.Sprintf("%s:%s:", childnode.Name, childnode.Hash)))
		}
	case file:
		// Compute n.Hash by concatenating all BlockRef hashes in n.FileNode.
		for _, object := range n.FileNode.Objects {
			hash.Write([]byte(object.Hash))
		}
	default:
		return errorf(Internal,
			"malformed file at \"%s\" is neither a file nor a directory", path)
	}

	// Update hash of 'n'
	n.Hash = hash.Sum(nil)
	if err := h.put(tx, path, n); err != nil {
		return err
	}
	return changed(tx).Delete(b(path))
}

func (h *dbHashTree) Finish() (_ HashTree, retErr error) {
	if err := h.Batch(func(tx *bolt.Tx) error {
		return h.canonicalize(tx, "")
	}); err != nil {
		return nil, err
	}
	newFile := dbFile()
	f, err := os.Create(newFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err := h.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(f)
		return err
	}); err != nil {
		return nil, err
	}
	return newDBHashTree(newFile)
}
