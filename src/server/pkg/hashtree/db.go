package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	pathlib "path"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"

	bolt "github.com/coreos/bbolt"
	globlib "github.com/gobwas/glob"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	FsBucket      = "fs"
	ChangedBucket = "changed"
	ObjectBucket  = "object"
	perm          = 0666
)

var (
	changedVal = []byte{1}
	nullByte   = []byte{0}
	slashByte  = []byte{'/'}
)

func fs(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(FsBucket))
}

func changed(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(ChangedBucket))
}

func object(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(ObjectBucket))
}

type dbHashTree struct {
	*bolt.DB
}

func slashEncode(b []byte) []byte {
	return bytes.Replace(b, slashByte, nullByte, -1)
}

func slashDecode(b []byte) []byte {
	return bytes.Replace(b, nullByte, slashByte, -1)
}

func s(b []byte) (result string) {
	if bytes.Equal(b, nullByte) {
		return ""
	}
	return string(slashDecode(b))
}

func b(s string) (result []byte) {
	if s == "" {
		return nullByte
	}
	return slashEncode([]byte(s))
}

func dbFile(storageRoot string) string {
	if storageRoot == "" {
		storageRoot = "/tmp"
	}
	return fmt.Sprintf("%s/hashtree/%s", storageRoot, uuid.NewWithoutDashes())
}

func NewDBHashTree(storageRoot string) (HashTree, error) {
	file := dbFile(storageRoot)
	if err := os.MkdirAll(pathlib.Dir(file), 0777); err != nil {
		return nil, err
	}
	result, err := newDBHashTree(file)
	if err != nil {
		return nil, err
	}
	if err := result.PutDir("/"); err != nil {
		return nil, err
	}
	return result, err
}

func DeserializeDBHashTree(storageRoot string, r io.Reader) (_ HashTree, retErr error) {
	file := dbFile(storageRoot)
	if err := os.MkdirAll(pathlib.Dir(file), 0777); err != nil {
		return nil, err
	}
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := io.Copy(f, r); err != nil {
		return nil, err
	}
	return newDBHashTree(file)
}

func newDBHashTree(file string) (HashTree, error) {
	db, err := bolt.Open(file, perm, nil)
	if err != nil {
		return nil, err
	}
	db.NoSync = true
	if err := db.Batch(func(tx *bolt.Tx) error {
		for _, bucket := range []string{FsBucket, ChangedBucket, ObjectBucket} {
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

func get(tx *bolt.Tx, path string) (*NodeProto, error) {
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
		node, err = get(tx, path)
		return err
	}); err != nil {
		return nil, err
	}
	return node, nil
}

// dirPrefix returns the prefix that keys must have to be considered under the
// directory at path
func dirPrefix(path string) string {
	if path == "" {
		return "" // all paths are under the root
	}
	return path + "/"
}

// iterDir iterates through the nodes under path, it errors with PathNotFound if path doesn't exist, it errors with PathConflict if path exists but isn't a directory.
func (h *dbHashTree) iterDir(tx *bolt.Tx, path string, f func(k, v []byte, c *bolt.Cursor) error) error {
	c := fs(tx).Cursor()
	k, v := c.Seek(b(path))
	if k == nil || s(k) != path {
		return errorf(PathNotFound, "file \"%s\" not found", path)
	}
	node := &NodeProto{}
	if err := node.Unmarshal(v); err != nil {
		return err
	}
	if node.DirNode == nil {
		return errorf(PathConflict, "the file at \"%s\" is not a directory",
			path)
	}
	for k, v = c.Next(); k != nil && strings.HasPrefix(s(k), dirPrefix(path)); k, v = c.Next() {
		trimmed := strings.TrimPrefix(s(k), path)
		if trimmed == "" || strings.Count(trimmed, "/") > 1 {
			// trimmed == "" -> this is path itself
			// trimmed[0] != '/' -> this is a sibling of path
			// strings.Count(trimmed, "/") > 1 -> this is the child of a child of path
			// If any of these are true we don't call f on the value because it's not in the directory.
			continue
		}
		if err := f(k, v, c); err != nil {
			return err
		}
	}
	return nil
}

// List retrieves the list of files and subdirectories of the directory at
// 'path'.
func (h *dbHashTree) List(path string) ([]*NodeProto, error) {
	path = clean(path)
	var result []*NodeProto
	if err := h.View(func(tx *bolt.Tx) error {
		return h.iterDir(tx, path, func(k, v []byte, _ *bolt.Cursor) error {
			node := &NodeProto{}
			if err := node.Unmarshal(v); err != nil {
				return err
			}
			result = append(result, node)
			return nil
		})
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
		for k, v := c.Seek(b(path)); k != nil && strings.HasPrefix(s(k), path); k, v = c.Next() {
			node := &NodeProto{}
			if err := node.Unmarshal(v); err != nil {
				return err
			}
			nodePath := s(k)
			if nodePath == "" {
				nodePath = "/"
			}
			if nodePath != path && !strings.HasPrefix(nodePath, path+"/") {
				// node is a sibling of path, and thus doesn't get walked
				continue
			}
			if err := f(nodePath, node); err != nil {
				return err
			}
		}
		return nil
	})
}

func diff(newTx, oldTx *bolt.Tx, newPath string, oldPath string, recursiveDepth int64, f func(string, *NodeProto, bool) error) error {
	oldPath = clean(oldPath)
	newPath = clean(newPath)
	newNode, err := get(newTx, newPath)
	if err != nil && Code(err) != PathNotFound {
		return err
	}
	oldNode, err := get(oldTx, oldPath)
	if err != nil && Code(err) != PathNotFound {
		return err
	}
	if (newNode == nil && oldNode == nil) ||
		(newNode != nil && oldNode != nil && bytes.Equal(newNode.Hash, oldNode.Hash)) {
		return nil
	}
	var newC *childCursor
	if newNode != nil {
		if newNode.FileNode != nil || recursiveDepth == 0 {
			if err := f(newPath, newNode, true); err != nil {
				return err
			}
		} else if newNode.DirNode != nil {
			newC = NewChildCursor(newTx, newPath)
		}
	}
	var oldC *childCursor
	if oldNode != nil {
		if oldNode.FileNode != nil || recursiveDepth == 0 {
			if err := f(oldPath, oldNode, false); err != nil {
				return err
			}
		} else if oldNode.DirNode != nil {
			oldC = NewChildCursor(oldTx, newPath)
		}
	}
	if recursiveDepth > 0 || recursiveDepth == -1 {
		newDepth := recursiveDepth
		if recursiveDepth > 0 {
			newDepth--
		}
		switch {
		case oldC == nil && newC == nil:
			return nil
		case oldC == nil:
			for k := oldC.K(); k != nil; k, _ = oldC.Next() {
				child := pathlib.Base(s(k))
				if err := diff(newTx, oldTx, pathlib.Join(newPath, child), pathlib.Join(oldPath, child), newDepth, f); err != nil {
					return err
				}
			}
		case newC == nil:
			for k := newC.K(); k != nil; k, _ = newC.Next() {
				child := pathlib.Base(s(k))
				if err := diff(newTx, oldTx, pathlib.Join(newPath, child), pathlib.Join(oldPath, child), newDepth, f); err != nil {
					return err
				}
			}
		default:
		Children:
			for {
				var child string
				switch compare(newC, oldC) {
				case -1:
					child = pathlib.Base(s(newC.K()))
					newC.Next()
				case 0:
					if len(newC.K()) == 0 {
						break Children
					}
					child = pathlib.Base(s(newC.K()))
					newC.Next()
					oldC.Next()
				case 1:
					child = pathlib.Base(s(oldC.K()))
					oldC.Next()
				}
				if err := diff(newTx, oldTx, pathlib.Join(newPath, child), pathlib.Join(oldPath, child), newDepth, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (h *dbHashTree) Diff(oldHashTree HashTree, newPath string, oldPath string, recursiveDepth int64, f func(path string, node *NodeProto, new bool) error) (retErr error) {
	// Setup a txn for each hashtree, this is a bit complicated because we don't want to make 2 read tx to the same tree, if we did then should someone start a write tx inbetween them we would have a deadlock
	old := oldHashTree.(*dbHashTree)
	if old == nil {
		return fmt.Errorf("unrecognized HashTree type")
	}
	rollback := func(tx *bolt.Tx) {
		if err := tx.Rollback(); err != nil && retErr == nil {
			retErr = err
		}
	}
	var newTx *bolt.Tx
	var oldTx *bolt.Tx
	if h == oldHashTree {
		tx, err := h.Begin(false)
		if err != nil {
			return err
		}
		newTx = tx
		oldTx = tx
		defer rollback(tx)
	} else {
		var err error
		newTx, err = h.Begin(false)
		if err != nil {
			return err
		}
		defer rollback(newTx)
		oldTx, err = old.Begin(false)
		if err != nil {
			return err
		}
		defer rollback(oldTx)
	}
	return diff(newTx, oldTx, newPath, oldPath, recursiveDepth, func(path string, node *NodeProto, new bool) error {
		if new {
			return f(strings.TrimPrefix(path, newPath), node, new)
		} else {
			return f(strings.TrimPrefix(path, oldPath), node, new)
		}
	})
}

func (h *dbHashTree) Serialize(w io.Writer) error {
	return h.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

func (h *dbHashTree) Copy() (HashTree, error) {
	if err := h.Hash(); err != nil {
		return nil, err
	}
	r, w := io.Pipe()
	var eg errgroup.Group
	eg.Go(func() (retErr error) {
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return h.Serialize(w)
	})
	var result HashTree
	eg.Go(func() error {
		var err error
		result, err = DeserializeDBHashTree(pathlib.Dir(h.Path()), r)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *dbHashTree) Deserialize(r io.Reader) error {
	path := h.Path()
	if err := h.Close(); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	db, err := bolt.Open(path, perm, nil)
	if err != nil {
		return err
	}
	h.DB = db
	return nil
}

func put(tx *bolt.Tx, path string, node *NodeProto) error {
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
		pnode, err := get(tx, parent)
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
		if err := put(tx, parent, pnode); err != nil {
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
		node, err := get(tx, path)
		if err != nil && Code(err) != PathNotFound {
			return err
		}
		if node == nil {
			node = &NodeProto{
				Name:     base(path),
				FileNode: &FileNodeProto{},
			}
		} else if node.nodetype() != file {
			return errorf(PathConflict, "could not put file at \"%s\"; a file of "+
				"type %s is already there", path, node.nodetype().tostring())
		}
		// Append new objects.  Remove existing objects if overwriting.
		if overwriteIndex != nil && overwriteIndex.Index <= int64(len(node.FileNode.Objects)) {
			node.FileNode.Objects = node.FileNode.Objects[:overwriteIndex.Index]
		}
		node.SubtreeSize += sizeDelta
		node.FileNode.Objects = append(node.FileNode.Objects, objects...)
		// Put the node
		if err := put(tx, path, node); err != nil {
			return err
		}
		return h.visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.DirNode == nil {
				// node created as part of this visit call, fill in the basics
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
		node, err := get(tx, path)
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
		if err := put(tx, path, node); err != nil {
			return err
		}
		return h.visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.DirNode == nil {
				// node created as part of this visit call, fill in the basics
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
	if err := h.iterDir(tx, path, func(k, v []byte, c *bolt.Cursor) error {
		return c.Delete()
	}); err != nil && Code(err) != PathConflict {
		return err
	}
	return fs(tx).Delete(b(path))
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
			if _, err := get(tx, path); err != nil && Code(err) == PathNotFound {
				continue
			}
			// Remove 'path' and all nodes underneath it from h.fs
			if err := h.deleteDir(tx, path); err != nil {
				return err
			}
			size := paths[path].SubtreeSize
			// Remove 'path' from its parent directory
			parent, child := split(path)
			pnode, err := get(tx, parent)
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
			put(tx, parent, pnode)
			// Mark nodes as 'changed' back to root
			if err := h.visit(tx, path, func(node *NodeProto, parent, child string) error {
				// If node.DirNode is nil it means either the parent didn't
				// exist (and thus was deserialized fron nil) or it does exist
				// but thinks it's a file, both are errors.
				if node.DirNode == nil {
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
	destNode, err := get(tx, path)
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
		put(tx, path, destNode)
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
	put(tx, path, destNode)
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

func (h *dbHashTree) PutObject(o *pfs.Object, blockRef *pfs.BlockRef) error {
	data, err := blockRef.Marshal()
	if err != nil {
		return err
	}
	return h.Batch(func(tx *bolt.Tx) error {
		return object(tx).Put(b(o.Hash), data)
	})
}

func (h *dbHashTree) GetObject(o *pfs.Object) (*pfs.BlockRef, error) {
	result := &pfs.BlockRef{}
	if err := h.View(func(tx *bolt.Tx) error {
		data := object(tx).Get(b(o.Hash))
		if data == nil {
			return errorf(ObjectNotFound, "object \"%s\" not found", o.Hash)
		}
		return result.Unmarshal(data)
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *dbHashTree) changed(tx *bolt.Tx, path string) bool {
	return changed(tx).Get(b(path)) != nil
}

func (h *dbHashTree) canonicalize(tx *bolt.Tx, path string) error {
	path = clean(path)
	if !h.changed(tx, path) {
		return nil // Node is already canonical
	}
	n, err := get(tx, path)
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
			childnode, err := get(tx, childpath)
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
	if err := put(tx, path, n); err != nil {
		return err
	}
	return changed(tx).Delete(b(path))
}

func (h *dbHashTree) Hash() error {
	return h.Batch(func(tx *bolt.Tx) error {
		return h.canonicalize(tx, "")
	})
}

type nodetype uint8

const (
	none         nodetype = iota // No file is present at this point in the tree
	directory                    // The file at this point in the tree is a directory
	file                         // ... is a regular file
	unrecognized                 // ... is an an unknown type
)

func (n *NodeProto) nodetype() nodetype {
	switch {
	case n == nil || (n.DirNode == nil && n.FileNode == nil):
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

var globRegex = regexp.MustCompile(`[*?\[\]\{\}!]`)

// isGlob checks if the pattern contains a glob character
func isGlob(pattern string) bool {
	return globRegex.Match([]byte(pattern))
}

// GetHashTreeObject is a convenience function to deserialize a HashTree from an object in the object store.
func GetHashTreeObject(pachClient *client.APIClient, storageRoot string, treeRef *pfs.Object) (HashTree, error) {
	return getHashTree(storageRoot, func(w io.Writer) error {
		return pachClient.GetObject(treeRef.Hash, w)
	})
}

// GetHashTreeObject is a convenience function to deserialize a HashTree from an tagged object in the object store.
func GetHashTreeTag(pachClient *client.APIClient, storageRoot string, treeRef *pfs.Tag) (HashTree, error) {
	return getHashTree(storageRoot, func(w io.Writer) error {
		return pachClient.GetTag(treeRef.Name, w)
	})
}

func getHashTree(storageRoot string, f func(io.Writer) error) (HashTree, error) {
	r, w := io.Pipe()
	var eg errgroup.Group
	eg.Go(func() (retErr error) {
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return f(w)
	})
	var tree HashTree
	eg.Go(func() error {
		var err error
		tree, err = DeserializeDBHashTree(storageRoot, r)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return tree, nil
}

// PutHashTree is a convenience function for putting a HashTree to an object store.
func PutHashTree(pachClient *client.APIClient, tree HashTree, tags ...string) (*pfs.Object, error) {
	r, w := io.Pipe()
	var eg errgroup.Group
	eg.Go(func() (retErr error) {
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return tree.Serialize(w)
	})
	var treeRef *pfs.Object
	eg.Go(func() error {
		var err error
		treeRef, _, err = pachClient.PutObject(r, tags...)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return treeRef, nil
}

// childCursor efficiently iterates the children of a directory
type childCursor struct {
	c *bolt.Cursor
	// childCursor efficiently iterates the children of a directory
	dir []byte
	k   []byte
	v   []byte
}

func NewChildCursor(tx *bolt.Tx, path string) *childCursor {
	path = clean(path)
	c := fs(tx).Cursor()
	dir := b(path)
	k, v := c.Seek(append(dir, nullByte[0]))
	if !bytes.HasPrefix(k, dir) {
		k, v = nil, nil
	}
	return &childCursor{
		c:   c,
		dir: dir,
		k:   k,
		v:   v,
	}
}

func (d *childCursor) K() []byte {
	return d.k
}

func (d *childCursor) V() []byte {
	return d.v
}

func (d *childCursor) Next() ([]byte, []byte) {
	if d.k == nil {
		return nil, nil
	}
	k, v := d.c.Seek(append(d.k, 1))
	if !bytes.HasPrefix(k, d.dir) {
		k, v = nil, nil
	}
	d.k, d.v = k, v
	return k, v
}

func compare(a, b *childCursor) int {
	switch {
	case a.k == nil && b.k == nil:
		return 0
	case b.k == nil:
		return -1
	case a.k == nil:
		return 1
	default:
		return bytes.Compare(bytes.TrimPrefix(a.k, a.dir), bytes.TrimPrefix(b.k, b.dir))
	}
}
