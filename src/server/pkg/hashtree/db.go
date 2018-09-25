package hashtree

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	pathlib "path"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	bolt "github.com/coreos/bbolt"
	globlib "github.com/gobwas/glob"
	"github.com/golang/snappy"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	FsBucket      = "fs"
	ChangedBucket = "changed"
	ObjectBucket  = "object"
	DatumBucket   = "datum"
	perm          = 0666
	mergeBufSize  = 1 << (10 * 2)
)

var (
	buckets   = []string{DatumBucket, FsBucket, ChangedBucket, ObjectBucket}
	exists    = []byte{1}
	nullByte  = []byte{0}
	slashByte = []byte{'/'}
	// A path should not have a globbing character
	SentinelByte = []byte{'*'}
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

func datum(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(b(DatumBucket))
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
	result, err := NewDBHashTree(storageRoot)
	if err != nil {
		return nil, err
	}
	if err := result.Deserialize(r); err != nil {
		return nil, err
	}
	return result, nil
}

func newDBHashTree(file string) (HashTree, error) {
	db, err := bolt.Open(file, perm, nil)
	if err != nil {
		return nil, err
	}
	db.NoSync = true
	db.NoGrowSync = true
	db.MaxBatchDelay = 0
	if err := db.Batch(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
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

// (bryce) This get implementation is very naive and will be changed later.
func Get(mq *MergePriorityQueue, filePath string) (*NodeProto, error) {
	filePath = clean(filePath)
	var fileNode *NodeProto
	if err := Nodes(mq, b(filePath), func(path string, node *NodeProto) error {
		if path == filePath {
			fileNode = node
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if fileNode == nil {
		return nil, errorf(PathNotFound, "file \"%s\" not found", filePath)
	}
	return fileNode, nil
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
func iterDir(tx *bolt.Tx, path string, f func(k, v []byte, c *bolt.Cursor) error) error {
	node, err := get(tx, path)
	if err != nil {
		return err
	}
	if node.DirNode == nil {
		return errorf(PathConflict, "the file at \"%s\" is not a directory",
			path)
	}
	c := NewChildCursor(tx, path)
	for k, v := c.K(), c.V(); k != nil; k, v = c.Next() {
		if err := f(k, v, c.c); err != nil {
			if err == errutil.ErrBreak {
				return nil
			}
			return err
		}
	}
	return nil
}

func list(tx *bolt.Tx, path string, f func(*NodeProto) error) error {
	return iterDir(tx, path, func(_, v []byte, _ *bolt.Cursor) error {
		node := &NodeProto{}
		if err := node.Unmarshal(v); err != nil {
			return err
		}
		return f(node)
	})
}

// List retrieves the list of files and subdirectories of the directory at
// 'path'.
func (h *dbHashTree) List(path string, f func(*NodeProto) error) error {
	path = clean(path)
	return h.View(func(tx *bolt.Tx) error {
		return list(tx, path, f)
	})
}

func (h *dbHashTree) ListAll(path string) ([]*NodeProto, error) {
	var result []*NodeProto
	if err := h.List(path, func(node *NodeProto) error {
		result = append(result, node)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func List(mq *MergePriorityQueue, root string, pattern string, f func(string, *NodeProto) error) (retErr error) {
	root = clean(root)
	if pattern != "/" {
		pattern = clean(pattern)
	}
	g, err := globlib.Compile(pattern, '/')
	if err != nil {
		return errorf(MalformedGlob, err.Error())
	}
	return Nodes(mq, b(root), func(path string, node *NodeProto) error {
		if (g.Match(path) && node.DirNode == nil) || (g.Match(pathlib.Dir(path))) {
			return f(path, node)
		}
		return nil
	})
}

func glob(tx *bolt.Tx, pattern string, f func(string, *NodeProto) error) error {
	if !isGlob(pattern) {
		node, err := get(tx, pattern)
		if err != nil {
			return err
		}
		return f(pattern, node)
	}

	g, err := globlib.Compile(pattern, '/')
	if err != nil {
		return errorf(MalformedGlob, err.Error())
	}
	c := fs(tx).Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if g.Match(s(k)) {
			node := &NodeProto{}
			if node.Unmarshal(v); err != nil {
				return err
			}
			if err := f(s(k), node); err != nil {
				if err == errutil.ErrBreak {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (h *dbHashTree) Glob(pattern string, f func(string, *NodeProto) error) error {
	pattern = clean(pattern)
	return h.View(func(tx *bolt.Tx) error {
		return glob(tx, pattern, f)
	})
}

func (h *dbHashTree) GlobAll(pattern string) (map[string]*NodeProto, error) {
	res := make(map[string]*NodeProto)
	if err := h.Glob(pattern, func(path string, node *NodeProto) error {
		res[path] = node
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func Glob(mq *MergePriorityQueue, root string, pattern string, f func(string, *NodeProto) error) (retErr error) {
	root = clean(root)
	pattern = clean(pattern)
	g, err := globlib.Compile(pattern, '/')
	if err != nil {
		return errorf(MalformedGlob, err.Error())
	}
	return Nodes(mq, b(root), func(path string, node *NodeProto) error {
		if g.Match(path) {
			return f(path, node)
		}
		return nil
	})
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
				if err == errutil.ErrBreak {
					return nil
				}
				return err
			}
		}
		return nil
	})
}

func Walk(mq *MergePriorityQueue, root string, f func(path string, node *NodeProto) error) error {
	root = clean(root)
	return Nodes(mq, b(root), func(path string, node *NodeProto) error {
		if err := f(path, node); err != nil {
			if err == errutil.ErrBreak {
				return nil
			}
			return err
		}
		return nil
	})
}

func diff(newTx, oldTx *bolt.Tx, newPath string, oldPath string, recursiveDepth int64, f func(string, *NodeProto, bool) error) error {
	newNode, err := get(newTx, clean(newPath))
	if err != nil && Code(err) != PathNotFound {
		return err
	}
	oldNode, err := get(oldTx, clean(oldPath))
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
			for k := newC.K(); k != nil; k, _ = newC.Next() {
				child := pathlib.Base(s(k))
				if err := diff(newTx, oldTx, pathlib.Join(newPath, child), pathlib.Join(oldPath, child), newDepth, f); err != nil {
					return err
				}
			}
		case newC == nil:
			for k := oldC.K(); k != nil; k, _ = oldC.Next() {
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
	return diff(newTx, oldTx, newPath, oldPath, recursiveDepth, f)
}

func (h *dbHashTree) Serialize(_w io.Writer) error {
	w := pbutil.NewWriter(snappy.NewWriter(_w))
	return h.View(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			b := tx.Bucket(b(bucket))
			if err := w.Write(
				&BucketHeader{
					Bucket: bucket,
				}); err != nil {
				return err
			}
			if err := b.ForEach(func(k, v []byte) error {
				if err := w.WriteBytes(k); err != nil {
					return err
				}
				return w.WriteBytes(v)
			}); err != nil {
				return err
			}
			if err := w.WriteBytes(SentinelByte); err != nil {
				return err
			}
		}
		return nil
	})
}

type keyValue struct {
	k, v []byte
}

func (h *dbHashTree) Deserialize(_r io.Reader) error {
	r := pbutil.NewReader(snappy.NewReader(_r))
	hdr := &BucketHeader{}
	batchSize := 10000
	kvs := make(chan *keyValue, batchSize/10)
	var eg errgroup.Group
	eg.Go(func() error {
		var bucket []byte
		for {
			count := 0
			if err := h.Update(func(tx *bolt.Tx) error {
				if bucket != nil {
					tx.Bucket(bucket).FillPercent = 1
				}
				for kv := range kvs {
					if kv.k == nil {
						bucket = kv.v
						continue
					}
					if err := tx.Bucket(bucket).Put(kv.k, kv.v); err != nil {
						return err
					}
					count++
					if count >= batchSize {
						return nil
					}
				}
				return nil
			}); err != nil {
				return err
			}
			if count <= 0 {
				return nil
			}
		}
	})
	for {
		hdr.Reset()
		if err := r.Read(hdr); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		bucket := b(hdr.Bucket)
		kvs <- &keyValue{nil, bucket}
		for {
			_k, err := r.ReadBytes()
			if err != nil {
				return err
			}
			if bytes.Equal(_k, SentinelByte) {
				break
			}
			// we need to make copies of k and v because the memory will be reused
			k := make([]byte, len(_k))
			copy(k, _k)
			_v, err := r.ReadBytes()
			if err != nil {
				return err
			}
			v := make([]byte, len(_v))
			copy(v, _v)
			kvs <- &keyValue{k, v}
		}
	}
	close(kvs)
	return eg.Wait()
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

func (h *dbHashTree) Destroy() error {
	path := h.Path()
	if err := h.Close(); err != nil {
		return err
	}
	return os.Remove(path)
}

func put(tx *bolt.Tx, path string, node *NodeProto) error {
	data, err := node.Marshal()
	if err != nil {
		return err
	}
	if err := changed(tx).Put(b(path), exists); err != nil {
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
func visit(tx *bolt.Tx, path string, update updateFn) error {
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
		return visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.DirNode == nil {
				// node created as part of this visit call, fill in the basics
				node.Name = base(parent)
				node.DirNode = &DirectoryNodeProto{}
			}
			node.SubtreeSize += sizeDelta
			return nil
		})
	})
}

func (h *dbHashTree) PutFileBlockRefs(path string, blockRefs []*pfs.BlockRef, size int64) error {
	return h.putFileBlockRefs(path, blockRefs, size)
}

func (h *dbHashTree) putFileBlockRefs(path string, blockRefs []*pfs.BlockRef, sizeDelta int64) error {
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
		node.SubtreeSize += sizeDelta
		node.FileNode.BlockRefs = append(node.FileNode.BlockRefs, blockRefs...)
		// Put the node
		if err := put(tx, path, node); err != nil {
			return err
		}
		return visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.DirNode == nil {
				// node created as part of this visit call, fill in the basics
				node.Name = base(parent)
				node.DirNode = &DirectoryNodeProto{}
			}
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
		return visit(tx, path, func(node *NodeProto, parent, child string) error {
			if node.DirNode == nil {
				// node created as part of this visit call, fill in the basics
				node.Name = base(parent)
				node.DirNode = &DirectoryNodeProto{}
			}
			return nil
		})
	})
}

// deleteDir deletes a directory and all the children under it
func deleteDir(tx *bolt.Tx, path string) error {
	c := fs(tx).Cursor()
	prefix := append(b(path), nullByte[0])
	for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		if err := c.Delete(); err != nil {
			return err
		}
	}
	return fs(tx).Delete(b(path))
}

func (h *dbHashTree) DeleteFile(path string) error {
	path = clean(path)

	// Delete root means delete all files
	if path == "" {
		path = "/*"
	}
	return h.Batch(func(tx *bolt.Tx) error {
		if err := glob(tx, path, func(path string, node *NodeProto) error {
			// Check if the file has been deleted already
			if _, err := get(tx, path); err != nil && Code(err) == PathNotFound {
				return nil
			}
			// Remove 'path' and all nodes underneath it from h.fs
			if err := deleteDir(tx, path); err != nil {
				return err
			}
			size := node.SubtreeSize
			// Remove 'path' from its parent directory
			// TODO(bryce) Decide if this should be removed.
			parent, _ := split(path)
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
			put(tx, parent, pnode)
			// Mark nodes as 'changed' back to root
			if err := visit(tx, path, func(node *NodeProto, parent, child string) error {
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
			return nil
		}); err != nil && Code(err) != PathNotFound {
			// Deleting a non-existent file should be a no-op
			return err
		}
		return nil
	})
}

type Mergeable interface {
	NextBucket(bucket string) error
	Next() error
	K() []byte
	V() []byte
	Close() error
}

type mergeCursor struct {
	tx     *bolt.Tx
	cursor *bolt.Cursor
	k, v   []byte
}

func NewMergeCursor(h *dbHashTree) (*mergeCursor, error) {
	tx, err := h.Begin(false)
	if err != nil {
		return nil, err
	}
	m := &mergeCursor{
		tx: tx,
	}
	return m, nil
}

func (m *mergeCursor) NextBucket(bucket string) error {
	m.cursor = m.tx.Bucket(b(bucket)).Cursor()
	m.k, m.v = m.cursor.First()
	return nil
}

func (m *mergeCursor) Next() error {
	m.k, m.v = m.cursor.Next()
	return nil
}

func (m *mergeCursor) K() []byte {
	return m.k
}

func (m *mergeCursor) V() []byte {
	return m.v
}

func (m *mergeCursor) Close() error {
	return m.tx.Rollback()
}

type mergeStream struct {
	r         io.ReadCloser
	pbr       pbutil.Reader
	hdr       *BucketHeader
	done      bool
	k, v      []byte
	filter    func(path []byte) (bool, error)
	nextChan  chan *keyValue
	errChan   chan error
	closeChan chan struct{}
}

func NewMergeStream(r io.ReadCloser, filter func(path []byte) (bool, error)) (*mergeStream, error) {
	rBuf := bufio.NewReaderSize(r, mergeBufSize)
	if _, err := rBuf.Peek(mergeBufSize); err != nil {
		return nil, err
	}
	m := &mergeStream{
		r:         r,
		pbr:       pbutil.NewReader(snappy.NewReader(rBuf)),
		hdr:       &BucketHeader{},
		filter:    filter,
		nextChan:  make(chan *keyValue, 10),
		errChan:   make(chan error, 1),
		closeChan: make(chan struct{}, 1),
	}
	return m, nil
}

func (m *mergeStream) NextBucket(bucket string) error {
	m.hdr.Reset()
	if err := m.pbr.Read(m.hdr); err != nil {
		return err
	}
	for m.hdr.Bucket != bucket {
		k, err := m.pbr.ReadBytes()
		if err != nil {
			return err
		}
		for !bytes.Equal(k, SentinelByte) {
			k, err = m.pbr.ReadBytes()
			if err != nil {
				return err
			}
		}
		if err := m.pbr.Read(m.hdr); err != nil {
			return err
		}
	}
	m.done = false
	// Creating goroutine that handles getting the next key/values in a stream
	go func() {
		for {
			_k, err := m.pbr.ReadBytes()
			if err != nil {
				m.errChan <- err
				return
			}
			if bytes.Equal(_k, SentinelByte) {
				m.nextChan <- &keyValue{nil, nil}
				return
			}
			if m.hdr.Bucket == FsBucket {
				ok, err := m.filter(_k)
				if err != nil {
					m.errChan <- err
					return
				}
				if !ok {
					if _, err := m.pbr.ReadBytes(); err != nil {
						m.errChan <- err
						return
					}
					continue
				}
			}
			// we need to make copies of k and v because the memory will be reused
			k := make([]byte, len(_k))
			copy(k, _k)
			_v, err := m.pbr.ReadBytes()
			if err != nil {
				m.errChan <- err
				return
			}
			v := make([]byte, len(_v))
			copy(v, _v)
			select {
			case <-m.closeChan:
				close(m.nextChan)
				close(m.errChan)
				if err := m.r.Close(); err != nil {
					fmt.Println("Error closing merge stream")
				}
				return
			case m.nextChan <- &keyValue{k, v}:
			}
		}
	}()
	if err := m.Next(); err != nil {
		return err
	}
	return nil
}

func (m *mergeStream) Next() error {
	if m.done {
		return nil
	}
	select {
	case kv := <-m.nextChan:
		m.k = kv.k
		m.v = kv.v
		if m.k == nil {
			m.done = true
		}
		return nil
	case err := <-m.errChan:
		return err
	}
}

func (m *mergeStream) K() []byte {
	return m.k
}

func (m *mergeStream) V() []byte {
	return m.v
}

func (m *mergeStream) Close() error {
	m.closeChan <- struct{}{}
	return nil
}

func MergeTrees(w io.WriteCloser, rs []io.ReadCloser, filter func(path []byte) (bool, error)) (size uint64, retErr error) {
	var srcs []Mergeable
	for _, r := range rs {
		src, err := NewMergeStream(r, filter)
		if err != nil {
			return 0, err
		}
		srcs = append(srcs, src)
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
		for _, src := range srcs {
			if err := src.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}
	}()
	pbw := pbutil.NewWriter(snappy.NewWriter(w))
	out := func(k, v []byte) error {
		if bytes.Equal(k, nullByte) {
			n := &NodeProto{}
			if err := n.Unmarshal(v); err != nil {
				return err
			}
			size = uint64(n.SubtreeSize)
		}
		if err := pbw.WriteBytes(k); err != nil {
			return err
		}
		if err := pbw.WriteBytes(v); err != nil {
			return err
		}
		return nil
	}
	if err := MergeBucket(DatumBucket, pbw, srcs, func(k []byte, vs [][]byte) error {
		return out(k, vs[0])
	}); err != nil {
		return 0, err
	}
	if err := MergeBucket(FsBucket, pbw, srcs, func(k []byte, vs [][]byte) error {
		if len(vs) <= 1 {
			return out(k, vs[0])
		}
		destNode := &NodeProto{
			Name:        base(s(k)),
			SubtreeSize: 0,
			Hash:        nil,
			FileNode:    nil,
			DirNode:     nil,
		}
		n := &NodeProto{}
		if err := n.Unmarshal(vs[0]); err != nil {
			return err
		}
		if n.nodetype() == directory {
			destNode.DirNode = &DirectoryNodeProto{}
		} else if n.nodetype() == file {
			destNode.FileNode = &FileNodeProto{}
		} else {
			return errorf(Internal, "could not merge unrecognized file type at "+
				"\"%s\", which is neither a file nor a directory", s(k))
		}
		sizeDelta := int64(0)
		for _, v := range vs {
			n := &NodeProto{}
			if err := n.Unmarshal(v); err != nil {
				return err
			}
			if n.nodetype() != destNode.nodetype() {
				return errorf(PathConflict, "could not merge path \"%s\" "+
					"which is a regular-file in some hashtrees and a directory in others", s(k))
			}
			switch n.nodetype() {
			case directory:
				sizeDelta += n.SubtreeSize
			case file:
				// Append new block refs, and update size of target node (since that can't be
				// done in canonicalize)
				destNode.FileNode.BlockRefs = append(destNode.FileNode.BlockRefs, n.FileNode.BlockRefs...)
				sizeDelta += n.SubtreeSize
			default:
				return errorf(Internal, "malformed file at \"%s\" in source "+
					"hashtree is neither a regular-file nor a directory", s(k))
			}
		}
		destNode.SubtreeSize += sizeDelta
		v, err := destNode.Marshal()
		if err != nil {
			return err
		}
		return out(k, v)
	}); err != nil {
		return 0, err
	}
	if err := MergeBucket(ObjectBucket, pbw, srcs, func(k []byte, vs [][]byte) error {
		return out(k, vs[0])
	}); err != nil {
		return 0, err
	}
	return size, nil
}

type MergePriorityQueue struct {
	queue    []Mergeable
	size     int
	nextChan chan Mergeable
	count    int
	mu       *sync.Mutex
	errChan  chan error
}

func NewMergePriorityQueue(srcs []Mergeable, numWorkers int) *MergePriorityQueue {
	mq := &MergePriorityQueue{
		queue:    make([]Mergeable, len(srcs)+1),
		nextChan: make(chan Mergeable, len(srcs)),
		mu:       &sync.Mutex{},
		errChan:  make(chan error, 1),
	}
	for _, src := range srcs {
		mq.Insert(src)
	}
	// First set of streams are ready
	mq.errChan <- nil
	// Creating goroutines that handle getting the next key/values in a set of streams
	for i := 0; i < numWorkers; i++ {
		go func() {
			for src := range mq.nextChan {
				if err := src.Next(); err != nil {
					mq.errChan <- err
				}
				mq.mu.Lock()
				mq.Insert(src)
				mq.count--
				if mq.count <= 0 {
					mq.errChan <- nil
				}
				mq.mu.Unlock()
			}
		}()
	}
	return mq
}

func (mq *MergePriorityQueue) Insert(src Mergeable) {
	// Streams with a nil key are at the end
	if src.K() == nil {
		return
	}
	mq.size++
	idx := mq.size
	mq.queue[idx] = src
	for idx > 1 {
		if bytes.Compare(mq.queue[idx].K(), mq.queue[idx/2].K()) > 0 {
			break
		}
		mq.Swap(idx, idx/2)
		idx /= 2
	}
}

func (mq *MergePriorityQueue) Next() ([]byte, [][]byte, error) {
	if err := <-mq.errChan; err != nil || mq.size <= 0 {
		close(mq.nextChan)
		return nil, nil, err
	}
	// Get next streams to be merged
	srcs := []Mergeable{mq.queue[1]}
	mq.Fill(1)
	for mq.size > 0 && bytes.Equal(srcs[0].K(), mq.queue[1].K()) {
		srcs = append(srcs, mq.queue[1])
		mq.Fill(1)
	}
	mq.count = len(srcs)
	k := srcs[0].K()
	vs := [][]byte{}
	// Create value slice, and request next key/value in streams
	for _, src := range srcs {
		vs = append(vs, src.V())
		mq.nextChan <- src
	}
	return k, vs, nil
}

func (mq *MergePriorityQueue) Swap(i, j int) {
	tmp := mq.queue[i]
	mq.queue[i] = mq.queue[j]
	mq.queue[j] = tmp
}

func (mq *MergePriorityQueue) Fill(idx int) {
	mq.queue[idx] = mq.queue[mq.size]
	mq.size--
	var next int
	for {
		left, right := idx*2, idx*2+1
		if right <= mq.size && bytes.Compare(mq.queue[left].K(), mq.queue[right].K()) > 0 {
			next = right
		} else if left <= mq.size {
			next = left
		} else {
			break
		}
		if bytes.Compare(mq.queue[next].K(), mq.queue[idx].K()) >= 0 {
			break
		}
		mq.Swap(idx, next)
		idx = next
	}
}

func MergeBucket(bucket string, w pbutil.Writer, srcs []Mergeable, f func(k []byte, vs [][]byte) error) error {
	var eg errgroup.Group
	for _, s := range srcs {
		s := s
		eg.Go(func() error {
			if err := s.NextBucket(bucket); err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	if err := w.Write(&BucketHeader{Bucket: bucket}); err != nil {
		return err
	}
	mq := NewMergePriorityQueue(srcs, 10)
	if err := Merge(mq, f); err != nil {
		return err
	}
	if err := w.WriteBytes(SentinelByte); err != nil {
		return err
	}
	return nil
}

func Merge(mq *MergePriorityQueue, f func(k []byte, vs [][]byte) error) error {
	var total, total2 time.Duration
	t := time.Now()
	for {
		k, vs, err := mq.Next()
		if err != nil {
			return err
		}
		if k == nil {
			break
		}
		total += time.Since(t)
		t = time.Now()
		if err := f(k, vs); err != nil {
			return err
		}
		total2 += time.Since(t)
		t = time.Now()
	}
	fmt.Printf("(mergefilter) Time spent getting next in stream: %v\nTime spent merging and writing: %v\n", total, total2)
	return nil
}

func Nodes(mq *MergePriorityQueue, prefix []byte, f func(path string, node *NodeProto) error) error {
	var k []byte
	var vs [][]byte
	var err error
	// Skip to the prefix (this will be replaced by the object storage indexing later)
	for {
		k, vs, err = mq.Next()
		if err != nil {
			return err
		}
		if k == nil {
			return nil
		}
		if bytes.HasPrefix(k, prefix) {
			break
		}
	}
	// Iterate through paths with prefix
	for {
		if k == nil || !bytes.HasPrefix(k, prefix) {
			return nil
		}
		node := &NodeProto{}
		if err := node.Unmarshal(vs[0]); err != nil {
			return err
		}
		if err := f(s(k), node); err != nil {
			return err
		}
		k, vs, err = mq.Next()
		if err != nil {
			return err
		}
	}
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

func (h *dbHashTree) PutDatum(d string) error {
	return h.Batch(func(tx *bolt.Tx) error {
		return datum(tx).Put(b(d), exists)
	})
}

func (h *dbHashTree) HasDatum(d string) (bool, error) {
	var result bool
	if err := h.View(func(tx *bolt.Tx) error {
		val := datum(tx).Get(b(d))
		if val != nil {
			result = true
		}
		return nil
	}); err != nil {
		return false, err
	}
	return result, nil
}

func (h *dbHashTree) DiffDatums(datums chan string) (bool, error) {
	var result bool
	if err := h.Update(func(tx *bolt.Tx) error {
		for d := range datums {
			val := datum(tx).Get(b(d))
			if val != nil {
				if err := datum(tx).Put(b(d), nullByte); err != nil {
					return err
				}
			}
		}
		if err := datum(tx).ForEach(func(k, v []byte) error {
			if string(v) != string(nullByte) {
				result = true
			}
			return datum(tx).Put(k, exists)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return false, err
	}
	return result, nil
}

func hasChanged(tx *bolt.Tx, path string) bool {
	return changed(tx).Get(b(path)) != nil
}

func canonicalize(tx *bolt.Tx, path string) error {
	path = clean(path)
	if !hasChanged(tx, path) {
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
		// Note that the order of the children of n.DirNode are sorted when iterating.
		if err := iterDir(tx, path, func(k, _ []byte, _ *bolt.Cursor) error {
			childPath := s(k)
			if err := canonicalize(tx, childPath); err != nil {
				return err
			}
			childNode, err := get(tx, childPath)
			if err != nil {
				if Code(err) == PathNotFound {
					return errorf(Internal, "could not find file for \"%s\" while "+
						"updating hash of \"%s\"", childPath, path)
				}
				return err
			}
			// append child.Name and child.Hash to b
			hash.Write([]byte(fmt.Sprintf("%s:%s:", childNode.Name, childNode.Hash)))
			return nil
		}); err != nil {
			return err
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
		return canonicalize(tx, "")
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
	if !bytes.Equal(dir, nullByte) {
		dir = append(dir, nullByte[0])
	}
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

// DatumHashTree is an in memory version of the hashtree that is optimized and only works for sequential inserts followed by serialization.
// This data structure is useful for datum output because a datum's hashtree is created by walking the local filesystem then is uploaded.
type DatumHashTree struct {
	fs       []*node
	dirStack []*node
}

type node struct {
	path     string
	protoBuf *NodeProto
	hash     hash.Hash
}

func NewDatumHashTree(path string) *DatumHashTree {
	path = clean(path)
	n := &node{
		path: path,
		protoBuf: &NodeProto{
			Name:    base(path),
			DirNode: &DirectoryNodeProto{},
		},
		hash: sha256.New(),
	}
	return &DatumHashTree{
		fs:       []*node{n},
		dirStack: []*node{n},
	}
}

func (d *DatumHashTree) PutDir(path string) {
	path = clean(path)
	d.handleEndOfDirectory(path)
	n := &node{
		path: path,
		protoBuf: &NodeProto{
			Name:    base(path),
			DirNode: &DirectoryNodeProto{},
		},
		hash: sha256.New(),
	}
	d.fs = append(d.fs, n)
	d.dirStack = append(d.dirStack, n)
}

func (d *DatumHashTree) PutFile(path string, hash []byte, size int64, fileNode *FileNodeProto) {
	path = clean(path)
	d.handleEndOfDirectory(path)
	n := &node{
		path: path,
		protoBuf: &NodeProto{
			Name:        base(path),
			Hash:        hash,
			SubtreeSize: size,
			FileNode:    fileNode,
		},
	}
	d.fs = append(d.fs, n)
	d.dirStack[len(d.dirStack)-1].hash.Write([]byte(fmt.Sprintf("%s:%s:", n.protoBuf.Name, n.protoBuf.Hash)))
	d.dirStack[len(d.dirStack)-1].protoBuf.SubtreeSize += size
}

func (d *DatumHashTree) handleEndOfDirectory(path string) {
	parent, _ := split(path)
	if parent != d.dirStack[len(d.dirStack)-1].path {
		child := d.dirStack[len(d.dirStack)-1]
		child.protoBuf.Hash = child.hash.Sum(nil)
		d.dirStack = d.dirStack[:len(d.dirStack)-1]
		parent := d.dirStack[len(d.dirStack)-1]
		parent.hash.Write([]byte(fmt.Sprintf("%s:%s:", child.protoBuf.Name, child.protoBuf.Hash)))
		parent.protoBuf.SubtreeSize += child.protoBuf.SubtreeSize
	}
}

func (d *DatumHashTree) Serialize(_w io.Writer, datum string) error {
	w := pbutil.NewWriter(snappy.NewWriter(_w))
	if err := w.Write(&BucketHeader{Bucket: DatumBucket}); err != nil {
		return err
	}
	if err := w.WriteBytes([]byte(datum)); err != nil {
		return err
	}
	if err := w.WriteBytes(exists); err != nil {
		return err
	}
	if err := w.WriteBytes(SentinelByte); err != nil {
		return err
	}
	if err := w.Write(&BucketHeader{Bucket: FsBucket}); err != nil {
		return err
	}
	d.fs[0].protoBuf.Hash = d.fs[0].hash.Sum(nil)
	for _, n := range d.fs {
		if err := w.WriteBytes(b(n.path)); err != nil {
			return err
		}
		if err := w.Write(n.protoBuf); err != nil {
			return err
		}
	}
	if err := w.WriteBytes(SentinelByte); err != nil {
		return err
	}
	if err := w.Write(&BucketHeader{Bucket: ObjectBucket}); err != nil {
		return err
	}
	if err := w.WriteBytes(SentinelByte); err != nil {
		return err
	}
	return nil
}
