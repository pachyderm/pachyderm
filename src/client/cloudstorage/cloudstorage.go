package cloudstorage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	pathpkg "path"
	"strings"
	"time"

	"github.com/araddon/gou"
	"github.com/gogo/protobuf/types"
	"github.com/lytics/cloudstorage"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func init() {
	cloudstorage.Register(StoreType, pfsProvider)
}
func pfsProvider(conf *cloudstorage.Config) (cloudstorage.Store, error) {
	return NewPFSStore()
}

var (
	// Ensure Our PFSStore implement CloudStorage interfaces
	_ cloudstorage.StoreReader = (*PFSStore)(nil)
)

const (
	// AuthFileSystem Authentication Method
	AuthFileSystem cloudstorage.AuthMethod = "localfiles"

	// StoreType name of our PFS Storage provider = "localfs"
	StoreType = "pfs"
)

// PFSStore is client to local-filesystem store.
type PFSStore struct {
	pachClient *client.APIClient
}

// NewPFSStore create local store from storage path on local filesystem, and cachepath.
func NewPFSStore() (*PFSStore, error) {
	config := &serviceenv.GlobalConfiguration{}
	if err := cmdutil.Populate(config); err != nil {
		return nil, err
	}
	env := serviceenv.InitPachOnlyEnv(serviceenv.NewConfiguration(config))
	pachClient := env.GetPachClient(context.Background())
	return &PFSStore{
		pachClient: pachClient,
	}, nil
}

// Type is store type = "pfs"
func (l *PFSStore) Type() string {
	return StoreType
}
func (l *PFSStore) Client() interface{} {
	return l
}

// NewObject create new object of given name.
func (l *PFSStore) NewObject(objectname string) (cloudstorage.Object, error) {
	obj, err := l.Get(context.Background(), objectname)
	if err != nil && err != cloudstorage.ErrObjectNotFound {
		return nil, err
	} else if obj != nil {
		return nil, cloudstorage.ErrObjectExists
	}
	return &object{
		name: objectname,
	}, nil
}

func (l *PFSStore) parsePath(path string) (string, string, string) {
	repoAndCommit, path := pathpkg.Split(path)
	repoParts := strings.Split(repoAndCommit, ".")
	repo := repoParts[0]
	commit := "master"
	if len(repoParts) > 1 {
		commit = repoParts[1]
	}
	return repo, commit, path
}

// List objects at Query location.
func (l *PFSStore) List(ctx context.Context, query cloudstorage.Query) (*cloudstorage.ObjectsResponse, error) {
	resp := cloudstorage.NewObjectsResponse()
	objects := make(map[string]*object)
	metadatas := make(map[string]map[string]string)

	repo, commit, path := l.parsePath(query.Prefix)
	if err := l.pachClient.WithCtx(ctx).WalkFile(client.NewCommit(repo, "master", commit), path, func(fi *pfs.FileInfo) error {
		t, err := types.TimestampFromProto(fi.Committed)
		if err != nil {
			return err
		}
		objects[path] = &object{
			name:    fi.File.Path,
			updated: t,
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to walk files")
	}
	resp.Objects = query.ApplyFilters(resp.Objects)
	return resp, nil
}

// Objects returns an iterator over the objects in the local folder that match the Query q.
// If q is nil, no filtering is done.
func (l *PFSStore) Objects(ctx context.Context, csq cloudstorage.Query) (cloudstorage.ObjectIterator, error) {
	resp, err := l.List(ctx, csq)
	if err != nil {
		return nil, err
	}
	return &objectIterator{objects: resp.Objects}, nil
}

// Folders list of folders for given path query.
func (l *PFSStore) Folders(ctx context.Context, csq cloudstorage.Query) ([]string, error) {
	repo, commit, path := l.parsePath(csq.Prefix)
	var folders []string
	if err := l.pachClient.WithCtx(ctx).ListFile(client.NewCommit(repo, "master", commit), path, func(fi *pfs.FileInfo) error {
		if fi.FileType == pfs.FileType_DIR {
			folders = append(folders, fi.File.Path)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to list files")
	}
	return folders, nil
}

// NewReader create local file-system store reader.
func (l *PFSStore) NewReader(o string) (io.ReadCloser, error) {
	return l.NewReaderWithContext(context.Background(), o)
}
func (l *PFSStore) NewReaderWithContext(ctx context.Context, o string) (io.ReadCloser, error) {
	if !cloudstorage.Exists(o) {
		return nil, cloudstorage.ErrObjectNotFound
	}
	repo, commit, path := l.parsePath(o)
	r, err := l.pachClient.GetFileReader(client.NewCommit(repo, "master", commit), path)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(r), nil
}

func (l *PFSStore) NewWriter(o string, metadata map[string]string) (io.WriteCloser, error) {
	return l.NewWriterWithContext(context.Background(), o, metadata)
}
func (l *PFSStore) NewWriterWithContext(ctx context.Context, o string, metadata map[string]string, opts ...cloudstorage.Opts) (io.WriteCloser, error) {
	repo, commit, path := l.parsePath(o)
	return l.pachClient.PutFileWriter(client.NewCommit(repo, "master", commit), path), nil
}

func (l *PFSStore) Get(ctx context.Context, o string) (cloudstorage.Object, error) {
	fo := path.Join(l.storepath, o)

	if !cloudstorage.Exists(fo) {
		return nil, cloudstorage.ErrObjectNotFound
	}
	var updated time.Time
	if stat, err := os.Stat(fo); err == nil {
		updated = stat.ModTime()
	}

	return &object{
		name:      o,
		updated:   updated,
		storepath: fo,
		cachepath: cloudstorage.CachePathObj(l.cachepath, o, l.Id),
	}, nil
}

// Delete the object from underlying store.
func (l *PFSStore) Delete(ctx context.Context, obj string) error {
	fo := path.Join(l.storepath, obj)
	os.Remove(fo)
	mf := fo + ".metadata"
	if cloudstorage.Exists(mf) {
		os.Remove(mf)
	}
	return nil
}

func (l *PFSStore) String() string {
	return fmt.Sprintf("[id:%s file://%s/]", l.Id, l.storepath)
}

type objectIterator struct {
	objects cloudstorage.Objects
	err     error
	cursor  int
}

func (l *objectIterator) Next() (cloudstorage.Object, error) {
	if l.err != nil {
		return nil, l.err
	}
	if l.cursor >= len(l.objects) {
		return nil, iterator.Done
	}
	o := l.objects[l.cursor]
	l.cursor++
	return o, nil
}
func (l *objectIterator) Close() {}

type object struct {
	name     string
	updated  time.Time
	metadata map[string]string

	storepath string
	cachepath string

	cachedcopy *os.File
	readonly   bool
	opened     bool
}

func (o *object) StorageSource() string {
	return StoreType
}
func (o *object) Name() string {
	return o.name
}
func (o *object) String() string {
	return o.name
}
func (o *object) Updated() time.Time {
	return o.updated
}
func (o *object) MetaData() map[string]string {
	return o.metadata
}
func (o *object) SetMetaData(meta map[string]string) {
	o.metadata = meta
}

func (o *object) Delete() error {
	if err := o.Release(); err != nil {
		gou.Errorf("could not release %v", err)
	}
	if err := os.Remove(o.storepath); err != nil {
		return err
	}
	mf := o.storepath + ".metadata"
	if cloudstorage.Exists(mf) {
		if err := os.Remove(mf); err != nil {
			return err
		}
	}
	return nil
}

func (o *object) Open(accesslevel cloudstorage.AccessLevel) (*os.File, error) {
	if o.opened {
		return nil, fmt.Errorf("the store object is already opened. %s", o.storepath)
	}

	var readonly = accesslevel == cloudstorage.ReadOnly

	storecopy, err := os.OpenFile(o.storepath, os.O_RDWR|os.O_CREATE, 0665)
	if err != nil {
		return nil, fmt.Errorf("localfs: local=%q could not create storecopy err=%v", o.storepath, err)
	}
	defer storecopy.Close()

	err = cloudstorage.EnsureDir(o.cachepath)
	if err != nil {
		return nil, fmt.Errorf("localfs: cachepath=%s could not create cachedcopy dir err=%v", o.cachepath, err)
	}

	cachedcopy, err := os.Create(o.cachepath)
	if err != nil {
		return nil, fmt.Errorf("localfs: cachepath=%s could not create cachedcopy err=%v", o.cachepath, err)
	}

	_, err = io.Copy(cachedcopy, storecopy)
	if err != nil {
		return nil, fmt.Errorf("localfs: storepath=%s cachedcopy=%v could not copy from store to cache err=%v", o.storepath, cachedcopy.Name(), err)
	}

	if readonly {
		cachedcopy.Close()
		cachedcopy, err = os.Open(o.cachepath)
		if err != nil {
			return nil, fmt.Errorf("localfs: storepath=%s cachedcopy=%v could not opencache err=%v", o.storepath, cachedcopy.Name(), err)
		}
	} else {
		if _, err := cachedcopy.Seek(0, os.SEEK_SET); err != nil {
			return nil, fmt.Errorf("error seeking to start of cachedcopy err=%v", err) //don't retry on local fs errors
		}
	}

	o.cachedcopy = cachedcopy
	o.readonly = readonly
	o.opened = true
	return o.cachedcopy, nil
}

func (o *object) File() *os.File {
	return o.cachedcopy
}
func (o *object) Read(p []byte) (n int, err error) {
	return o.cachedcopy.Read(p)
}

// Write the given bytes to object.  Won't be writen until Close() or Sync() called.
func (o *object) Write(p []byte) (n int, err error) {
	if o.cachedcopy == nil {
		_, err := o.Open(cloudstorage.ReadWrite)
		if err != nil {
			return 0, err
		}
	}
	return o.cachedcopy.Write(p)
}

func (o *object) Sync() error {
	if !o.opened {
		return fmt.Errorf("object isn't opened %s", o.name)
	}
	if o.readonly {
		return fmt.Errorf("trying to Sync a readonly object %s", o.name)
	}

	cachedcopy, err := os.OpenFile(o.cachepath, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer cachedcopy.Close()

	storecopy, err := os.OpenFile(o.storepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return err
	}
	defer storecopy.Close()

	_, err = io.Copy(storecopy, cachedcopy)
	if err != nil {
		return err
	}

	if o.metadata != nil && len(o.metadata) > 0 {
		o.metadata = make(map[string]string)
	}

	fmd := o.storepath + ".metadata"
	return writemeta(fmd, o.metadata)
}

func writemeta(filename string, meta map[string]string) error {
	bm, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, bm, 0664)
	if err != nil {
		return err
	}
	return nil
}

func (o *object) Close() error {
	if !o.opened {
		return nil
	}

	defer func() {
		if o.cachedcopy != nil {
			n := o.cachedcopy.Name()
			os.Remove(n)
		}

		o.cachedcopy = nil
		o.opened = false
	}()

	if !o.readonly {
		err := o.cachedcopy.Sync()
		if err != nil {
			return err
		}
	}

	err := o.cachedcopy.Close()
	if err != nil {
		if !strings.Contains(err.Error(), os.ErrClosed.Error()) {
			return err
		}
	}

	if o.opened && !o.readonly {
		err := o.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *object) Release() error {
	if o.cachedcopy != nil {
		o.cachedcopy.Close()
		o.cachedcopy = nil
		o.opened = false
		err := os.Remove(o.cachepath)
		if err != nil {
			return err
		}
	}
	// most likely this doesn't exist so don't return error
	os.Remove(o.cachepath)
	return nil
}
