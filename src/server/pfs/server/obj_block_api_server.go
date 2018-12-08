package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/groupcache"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/limit"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	prefixLength          = 2
	alphabet              = "0123456789abcdef"
	objectCacheShares     = 8
	tagCacheShares        = 1
	objectInfoCacheShares = 1
	blockCacheShares      = 8
	blockKeySeparator     = "|"
	maxCachedObjectDenom  = 4                // We will only cache objects less than 1/maxCachedObjectDenom of total cache size
	bufferSize            = 15 * 1024 * 1024 // 15 MB
)

type objBlockAPIServer struct {
	log.Logger
	dir       string
	objClient obj.Client

	// cache
	objectCache     *groupcache.Group
	tagCache        *groupcache.Group
	objectInfoCache *groupcache.Group
	blockCache      *groupcache.Group
	// The total number of bytes cached for objects
	objectCacheBytes int64
	// The GC generation number.  Incrementing this number effectively
	// invalidates all current cache.
	generation int
	genLock    sync.RWMutex

	objectIndexes     map[string]*pfsclient.ObjectIndex
	objectIndexesLock sync.RWMutex
}

// In test mode, we use unique names for cache groups, since we might want
// to run multiple block servers locally, which would conflict if groups
// had the same name. We also do not report stats to prometheus
func newObjBlockAPIServer(dir string, cacheBytes int64, etcdAddress string, objClient obj.Client, test bool) (*objBlockAPIServer, error) {
	// defensive mesaure incase IsNotExist checking breaks due to underlying changes
	if err := obj.TestIsNotExist(objClient); err != nil {
		return nil, err
	}
	oneCacheShare := cacheBytes / (objectCacheShares + tagCacheShares + objectInfoCacheShares)
	s := &objBlockAPIServer{
		Logger:           log.NewLogger("pfs.BlockAPI.Obj"),
		dir:              dir,
		objClient:        objClient,
		objectIndexes:    make(map[string]*pfsclient.ObjectIndex),
		objectCacheBytes: oneCacheShare * objectCacheShares,
	}

	objectGroupName := "object"
	tagGroupName := "tag"
	objectInfoGroupName := "objectInfo"
	blockGroupName := "block"

	if test {
		uuid := uuid.New()
		objectGroupName += uuid
		tagGroupName += uuid
		objectInfoGroupName += uuid
		blockGroupName += uuid
	}

	s.objectCache = groupcache.NewGroup(objectGroupName, oneCacheShare*objectCacheShares, groupcache.GetterFunc(s.objectGetter))
	s.tagCache = groupcache.NewGroup(tagGroupName, oneCacheShare*tagCacheShares, groupcache.GetterFunc(s.tagGetter))
	s.objectInfoCache = groupcache.NewGroup(objectInfoGroupName, oneCacheShare*objectInfoCacheShares, groupcache.GetterFunc(s.objectInfoGetter))
	s.blockCache = groupcache.NewGroup(blockGroupName, oneCacheShare*blockCacheShares, groupcache.GetterFunc(s.blockGetter))

	if !test {
		RegisterCacheStats("tag", &s.tagCache.Stats)
		RegisterCacheStats("object", &s.objectCache.Stats)
		RegisterCacheStats("object_info", &s.objectInfoCache.Stats)
	}

	go s.watchGC(etcdAddress)
	return s, nil
}

// watchGC watches for GC runs and invalidate all cache when GC happens.
func (s *objBlockAPIServer) watchGC(etcdAddress string) {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		etcdClient, err := etcd.New(etcd.Config{
			Endpoints:   []string{etcdAddress},
			DialOptions: client.DefaultDialOptions(),
		})
		if err != nil {
			return fmt.Errorf("error instantiating etcd client: %v", err)
		}

		watcher, err := watch.NewWatcher(context.Background(), etcdClient, "", client.GCGenerationKey, nil)
		if err != nil {
			return fmt.Errorf("error instantiating watch stream from generation number: %v", err)
		}
		defer watcher.Close()

		for {
			ev, ok := <-watcher.Watch()
			if ev.Err != nil {
				return fmt.Errorf("error from generation number watch: %v", ev.Err)
			}
			if !ok {
				return fmt.Errorf("generation number watch stream closed unexpectedly")
			}
			newGen, err := strconv.Atoi(string(ev.Value))
			if err != nil {
				return fmt.Errorf("error converting the generation number: %v", err)
			}
			s.setGeneration(newGen)
		}
	}, b, func(err error, d time.Duration) error {
		logrus.Errorf("error running GC watcher in block server: %v; retrying in %s", err, d)
		return nil
	})
}

func (s *objBlockAPIServer) setGeneration(newGen int) {
	s.genLock.Lock()
	defer s.genLock.Unlock()
	if newGen > s.generation {
		s.generation = newGen
	}
}

func (s *objBlockAPIServer) getGeneration() int {
	s.genLock.RLock()
	defer s.genLock.RUnlock()
	return s.generation
}

func newMinioBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewMinioClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient, false)
}

func newAmazonBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewAmazonClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient, false)
}

func newGoogleBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewGoogleClientFromSecret(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient, false)
}

func newMicrosoftBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewMicrosoftClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient, false)
}

func newLocalBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewLocalClient(dir)
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient, true)
}

func (s *objBlockAPIServer) PutObject(server pfsclient.ObjectAPI_PutObjectServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	defer drainObjectServer(server)
	putObjectReader := &putObjectReader{
		server: server,
	}
	object, err := s.putObject(server.Context(), putObjectReader, false)
	if err != nil {
		return err
	}
	var eg errgroup.Group
	for _, tag := range putObjectReader.tags {
		tag := tag
		eg.Go(func() (retErr error) {
			index := &pfsclient.ObjectIndex{Tags: map[string]*pfsclient.Object{tag.Name: object}}
			return s.writeProto(s.tagPath(tag), index)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return server.SendAndClose(object)
}

func (s *objBlockAPIServer) PutObjectSplit(server pfsclient.ObjectAPI_PutObjectSplitServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	defer drainObjectServer(server)
	var objects []*pfsclient.Object
	putObjectReader := &putObjectReader{
		server: server,
	}
	for {
		object, err := s.putObject(server.Context(), putObjectReader, true)
		if object != nil {
			objects = append(objects, object)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return server.SendAndClose(&pfsclient.Objects{Objects: objects})
}

func (s *objBlockAPIServer) putObject(ctx context.Context, dataReader io.Reader, split bool) (_ *pfsclient.Object, retErr error) {
	hash := pfsclient.NewHash()
	r := io.TeeReader(dataReader, hash)
	block := &pfsclient.Block{Hash: uuid.NewWithoutDashes()}
	var size int64
	if err := func() (retErr error) {
		blockPath := s.blockPath(block)
		w, err := s.objClient.Writer(blockPath)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		if split {
			size, err = io.CopyN(w, r, pfsclient.ChunkSize)
		} else {
			buf := grpcutil.GetBuffer()
			defer grpcutil.PutBuffer(buf)
			size, err = io.CopyBuffer(w, r, buf)
		}
		if err != nil {
			if err != io.EOF {
				s.objClient.Delete(blockPath)
			}
			return err
		}
		return nil
	}(); err != nil {
		if err == io.EOF {
			defer func() {
				if retErr == nil {
					retErr = io.EOF
				}
			}()
		} else {
			return nil, err
		}
	}
	object := &pfsclient.Object{Hash: pfsclient.EncodeHash(hash.Sum(nil))}
	// Now that we have a hash of the object we can check if it already exists.
	resp, err := s.CheckObject(ctx, &pfsclient.CheckObjectRequest{Object: object})
	if err != nil {
		return nil, err
	}
	if resp.Exists {
		// the object already exists so we delete the block we put
		if err := s.objClient.Delete(s.blockPath(block)); err != nil {
			return nil, err
		}
	} else {
		blockRef := &pfsclient.BlockRef{
			Block: block,
			Range: &pfsclient.ByteRange{
				Lower: 0,
				Upper: uint64(size),
			},
		}
		if err := s.writeProto(s.objectPath(object), blockRef); err != nil {
			return nil, err
		}
	}
	return object, nil
}

func (s *objBlockAPIServer) PutObjects(server pfsclient.ObjectAPI_PutObjectsServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	defer drainObjectServer(server)
	request, err := server.Recv()
	if err != nil {
		return err
	}
	if request.Block == nil {
		return fmt.Errorf("first put objects request should include a block")
	}

	blockPath := s.blockPath(request.Block)
	putObjectReader := &putObjectReader{
		server: server,
	}
	w, err := s.objClient.Writer(blockPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	_, err = io.CopyBuffer(w, putObjectReader, buf)
	if err != nil {
		s.objClient.Delete(blockPath)
		return err
	}
	return nil
}

func (s *objBlockAPIServer) GetObject(request *pfsclient.Object, getObjectServer pfsclient.ObjectAPI_GetObjectServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	// First we inspect the object to see how big it is.
	objectInfo, err := s.InspectObject(getObjectServer.Context(), request)
	if err != nil {
		return err
	}
	if objectInfo == nil {
		logrus.Errorf("objectInfo is nil; info: %+v; request: %v", objectInfo, request)
		return nil
	} else if objectInfo.BlockRef == nil {
		logrus.Errorf("objectInfo.BlockRef is nil; info: %+v; request: %v", objectInfo, request)
		return nil
	} else if objectInfo.BlockRef.Range == nil {
		logrus.Errorf("objectInfo.BlockRef.Range is nil; info: %+v; request: %v", objectInfo, request)
		return nil
	}
	objectSize := objectInfo.BlockRef.Range.Upper - objectInfo.BlockRef.Range.Lower
	if (objectSize) >= uint64(s.objectCacheBytes/maxCachedObjectDenom) {
		// The object is a substantial portion of the available cache space so
		// we bypass the cache and stream it directly out of the underlying store.
		blockPath := s.blockPath(objectInfo.BlockRef.Block)
		r, err := s.objClient.Reader(blockPath, objectInfo.BlockRef.Range.Lower, objectSize)
		if err != nil {
			return err
		}
		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return grpcutil.WriteToStreamingBytesServer(r, getObjectServer)
	}
	var data []byte
	sink := groupcache.AllocatingByteSliceSink(&data)
	if err := s.objectCache.Get(getObjectServer.Context(), s.splitKey(request.Hash), sink); err != nil {
		return err
	}
	return grpcutil.WriteToStreamingBytesServer(bytes.NewReader(data), getObjectServer)
}

func (s *objBlockAPIServer) GetObjects(request *pfsclient.GetObjectsRequest, getObjectsServer pfsclient.ObjectAPI_GetObjectsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	offset := request.OffsetBytes
	size := request.SizeBytes
	for _, object := range request.Objects {
		// First we inspect the object to see how big it is.
		objectInfo, err := s.InspectObject(getObjectsServer.Context(), object)
		if err != nil {
			return err
		}
		if objectInfo == nil {
			logrus.Errorf("objectInfo is nil; info: %+v; request: %v", objectInfo, request)
			continue
		} else if objectInfo.BlockRef == nil {
			logrus.Errorf("objectInfo.BlockRef is nil; info: %+v; request: %v", objectInfo, request)
			continue
		} else if objectInfo.BlockRef.Range == nil {
			logrus.Errorf("objectInfo.BlockRef.Range is nil; info: %+v; request: %v", objectInfo, request)
			continue
		}

		objectSize := objectInfo.BlockRef.Range.Upper - objectInfo.BlockRef.Range.Lower
		if offset > objectSize {
			offset -= objectSize
			continue
		}
		readSize := objectSize - offset
		if size < readSize && request.SizeBytes != 0 {
			readSize = size
		}
		if request.TotalSize >= uint64(s.objectCacheBytes/maxCachedObjectDenom) {
			blockPath := s.blockPath(objectInfo.BlockRef.Block)
			r, err := s.objClient.Reader(blockPath, objectInfo.BlockRef.Range.Lower+offset, readSize)
			if err != nil {
				return err
			}
			if err := grpcutil.WriteToStreamingBytesServer(r, getObjectsServer); err != nil {
				return err
			}
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		} else {
			var data []byte
			sink := groupcache.AllocatingByteSliceSink(&data)
			if err := s.objectCache.Get(getObjectsServer.Context(), s.splitKey(object.Hash), sink); err != nil {
				return err
			}
			if uint64(len(data)) < offset+readSize {
				return fmt.Errorf("undersized object (this is likely a bug)")
			}
			if err := grpcutil.WriteToStreamingBytesServer(bytes.NewReader(data[offset:offset+readSize]), getObjectsServer); err != nil {
				return err
			}
		}
		// We've hit the offset so we set it to 0
		offset = 0
		if request.SizeBytes != 0 {
			size -= readSize
			if size == 0 {
				break
			}
		}
	}
	return nil
}

func (s *objBlockAPIServer) GetBlocks(request *pfsclient.GetBlocksRequest, getBlockServer pfsclient.ObjectAPI_GetBlocksServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	offset := request.OffsetBytes
	size := request.SizeBytes
	for _, blockRef := range request.BlockRefs {
		blockSize := blockRef.Range.Upper - blockRef.Range.Lower
		if offset > blockSize {
			offset -= blockSize
			continue
		}
		readSize := blockSize - offset
		if size < readSize && request.SizeBytes != 0 {
			readSize = size
		}
		if request.TotalSize >= uint64(s.objectCacheBytes/maxCachedObjectDenom) {
			blockPath := s.blockPath(blockRef.Block)
			r, err := s.objClient.Reader(blockPath, blockRef.Range.Lower+offset, readSize)
			if err != nil {
				return err
			}
			if err := grpcutil.WriteToStreamingBytesServer(r, getBlockServer); err != nil {
				return err
			}
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
			continue
		}
		var data []byte
		key := blockRef.Block.Hash + "|" + strconv.FormatUint(blockRef.Range.Lower, 10) + "|" + strconv.FormatUint(blockRef.Range.Upper, 10)
		sink := groupcache.AllocatingByteSliceSink(&data)
		if err := s.blockCache.Get(getBlockServer.Context(), key, sink); err != nil {
			return err
		}
		if uint64(len(data)) < offset+readSize {
			return fmt.Errorf("undersized object (this is likely a bug)")
		}
		if err := grpcutil.WriteToStreamingBytesServer(bytes.NewReader(data[offset:offset+readSize]), getBlockServer); err != nil {
			return err
		}
		// We've hit the offset so we set it to 0
		offset = 0
		if request.SizeBytes != 0 {
			size -= readSize
			if size == 0 {
				break
			}
		}
	}
	return nil
}

func (s *objBlockAPIServer) TagObject(ctx context.Context, request *pfsclient.TagObjectRequest) (response *types.Empty, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	// First inspect the object to make sure it actually exists
	resp, err := s.CheckObject(ctx, &pfsclient.CheckObjectRequest{Object: request.Object})
	if err != nil {
		return nil, err
	}
	if !resp.Exists {
		return nil, fmt.Errorf("object %v does not exist", request.Object)
	}
	var eg errgroup.Group
	for _, tag := range request.Tags {
		tag := tag
		eg.Go(func() (retErr error) {
			index := &pfsclient.ObjectIndex{Tags: map[string]*pfsclient.Object{tag.Name: request.Object}}
			return s.writeProto(s.tagPath(tag), index)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (s *objBlockAPIServer) InspectObject(ctx context.Context, request *pfsclient.Object) (response *pfsclient.ObjectInfo, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	objectInfo := &pfsclient.ObjectInfo{}
	sink := groupcache.ProtoSink(objectInfo)
	if err := s.objectInfoCache.Get(ctx, s.splitKey(request.Hash), sink); err != nil {
		return nil, err
	}
	return objectInfo, nil
}

func (s *objBlockAPIServer) CheckObject(ctx context.Context, request *pfsclient.CheckObjectRequest) (response *pfsclient.CheckObjectResponse, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return &pfsclient.CheckObjectResponse{
		Exists: s.objClient.Exists(s.objectPath(request.Object)),
	}, nil
}

func (s *objBlockAPIServer) ListObjects(request *pfsclient.ListObjectsRequest, listObjectsServer pfsclient.ObjectAPI_ListObjectsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return s.objClient.Walk(s.objectDir(), func(key string) error {
		return listObjectsServer.Send(client.NewObject(filepath.Base(key)))
	})
}

func (s *objBlockAPIServer) ListTags(request *pfsclient.ListTagsRequest, server pfsclient.ObjectAPI_ListTagsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	var eg errgroup.Group
	limiter := limit.New(100)
	s.objClient.Walk(path.Join(s.tagDir(), request.Prefix), func(key string) error {
		tag := filepath.Base(key)
		if request.IncludeObject {
			limiter.Acquire()
			eg.Go(func() error {
				defer limiter.Release()
				tagObjectIndex := &pfsclient.ObjectIndex{}
				if err := s.readProto(key, tagObjectIndex); err != nil {
					return err
				}
				for _, object := range tagObjectIndex.Tags {
					if err := server.Send(&pfsclient.ListTagsResponse{
						Tag:    &pfsclient.Tag{Name: tag},
						Object: object,
					}); err != nil {
						return err
					}
				}
				return nil
			})
		} else {
			if err := server.Send(&pfsclient.ListTagsResponse{
				Tag: &pfsclient.Tag{Name: tag},
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return eg.Wait()
}

func (s *objBlockAPIServer) DeleteTags(ctx context.Context, request *pfsclient.DeleteTagsRequest) (response *pfsclient.DeleteTagsResponse, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())

	limiter := limit.New(100)
	var eg errgroup.Group
	for _, tag := range request.Tags {
		tag := tag
		limiter.Acquire()
		eg.Go(func() error {
			defer limiter.Release()
			tagPath := s.tagPath(tag)
			if err := s.objClient.Delete(tagPath); err != nil && !s.isNotFoundErr(err) {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &pfsclient.DeleteTagsResponse{}, nil
}

func (s *objBlockAPIServer) isNotFoundErr(err error) bool {
	// GG golang
	patterns := []string{"not found", "not exist", "NotFound", "NotExist", "404"}
	errstr := err.Error()
	for _, pattern := range patterns {
		if strings.Contains(errstr, pattern) {
			return true
		}
	}
	return s.objClient.IsNotExist(err) || s.objClient.IsIgnorable(err)
}

func (s *objBlockAPIServer) DeleteObjects(ctx context.Context, request *pfsclient.DeleteObjectsRequest) (response *pfsclient.DeleteObjectsResponse, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())

	limiter := limit.New(100)
	var eg errgroup.Group
	for _, object := range request.Objects {
		object := object
		limiter.Acquire()
		eg.Go(func() error {
			defer limiter.Release()
			objectInfo, err := s.InspectObject(ctx, object)
			if err != nil && !s.isNotFoundErr(err) {
				return err
			}

			objPath := s.objectPath(object)
			if err := s.objClient.Delete(objPath); err != nil && !s.isNotFoundErr(err) {
				return err
			}

			if objectInfo != nil && objectInfo.BlockRef != nil && objectInfo.BlockRef.Block != nil {
				blockPath := s.blockPath(objectInfo.BlockRef.Block)
				if err := s.objClient.Delete(blockPath); err != nil && !s.isNotFoundErr(err) {
					return err
				}
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &pfsclient.DeleteObjectsResponse{}, nil
}

func (s *objBlockAPIServer) GetTag(request *pfsclient.Tag, getTagServer pfsclient.ObjectAPI_GetTagServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	object := &pfsclient.Object{}
	sink := groupcache.ProtoSink(object)
	if err := s.tagCache.Get(getTagServer.Context(), s.splitKey(request.Name), sink); err != nil {
		return err
	}
	return s.GetObject(object, getTagServer)
}

func (s *objBlockAPIServer) InspectTag(ctx context.Context, request *pfsclient.Tag) (response *pfsclient.ObjectInfo, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	object := &pfsclient.Object{}
	sink := groupcache.ProtoSink(object)
	if err := s.tagCache.Get(ctx, s.splitKey(request.Name), sink); err != nil {
		return nil, err
	}
	return s.InspectObject(ctx, object)
}

func (s *objBlockAPIServer) Compact(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := s.compact(); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (s *objBlockAPIServer) objectPrefix(prefix string) string {
	return s.objectPath(&pfsclient.Object{Hash: prefix})
}

func (s *objBlockAPIServer) tagPrefix(prefix string) string {
	return s.tagPath(&pfsclient.Tag{Name: prefix})
}

func (s *objBlockAPIServer) compact() (retErr error) {
	w, err := s.newBlockWriter(&pfsclient.Block{Hash: uuid.NewWithoutDashes()})
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var mu sync.Mutex
	var eg errgroup.Group
	objectIndex := &pfsclient.ObjectIndex{
		Objects: make(map[string]*pfsclient.BlockRef),
		Tags:    make(map[string]*pfsclient.Object),
	}
	var toDelete []string
	eg.Go(func() error {
		return s.objClient.Walk(s.objectDir(), func(name string) error {
			eg.Go(func() (retErr error) {
				blockRef := &pfsclient.BlockRef{}
				if err := s.readProto(name, blockRef); err != nil {
					return err
				}
				blockPath := s.blockPath(blockRef.Block)
				r, err := s.objClient.Reader(blockPath, blockRef.Range.Lower, blockRef.Range.Upper-blockRef.Range.Lower)
				if err != nil {
					return err
				}
				defer func() {
					if err := r.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				object, err := ioutil.ReadAll(r)
				if err != nil {
					return err
				}
				blockRef, err = w.Write(object)
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				objectIndex.Objects[filepath.Base(name)] = blockRef
				toDelete = append(toDelete, name, blockPath)
				return nil
			})
			return nil
		})
	})
	eg.Go(func() error {
		return s.objClient.Walk(s.tagDir(), func(name string) error {
			eg.Go(func() error {
				tagObjectIndex := &pfsclient.ObjectIndex{}
				if err := s.readProto(name, tagObjectIndex); err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				for tag, object := range tagObjectIndex.Tags {
					objectIndex.Tags[tag] = object
				}
				toDelete = append(toDelete, name)
				return nil
			})
			return nil
		})
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	prefixes := make(map[string]bool)
	for hash := range objectIndex.Objects {
		prefixes[hash[:2]] = true
	}
	for tag := range objectIndex.Tags {
		p := tag
		if len(p) > 2 {
			p = p[:2]
		}
		prefixes[p] = true
	}
	eg = errgroup.Group{}
	for prefix := range prefixes {
		prefix := prefix
		eg.Go(func() error {
			prefixObjectIndex := &pfsclient.ObjectIndex{
				Objects: make(map[string]*pfsclient.BlockRef),
				Tags:    make(map[string]*pfsclient.Object),
			}
			if err := s.readProto(s.indexPath(prefix), prefixObjectIndex); err != nil && !s.isNotFoundErr(err) {
				return err
			}
			for hash, blockRef := range objectIndex.Objects {
				if strings.HasPrefix(hash, prefix) {
					prefixObjectIndex.Objects[hash] = blockRef
				}
			}
			for tag, object := range objectIndex.Tags {
				if strings.HasPrefix(tag, prefix) {
					prefixObjectIndex.Tags[tag] = object
				}
			}
			return s.writeProto(s.indexPath(prefix), prefixObjectIndex)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	eg = errgroup.Group{}
	for _, file := range toDelete {
		file := file
		eg.Go(func() error {
			return s.objClient.Delete(file)
		})
	}
	return eg.Wait()
}

func (s *objBlockAPIServer) readProto(path string, pb proto.Unmarshaler) (retErr error) {
	r, err := s.objClient.Reader(path, 0, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		logrus.Infof("readProto(%s) yielded len(0) data", path)
	}
	return pb.Unmarshal(data)
}

func (s *objBlockAPIServer) writeProto(path string, pb proto.Marshaler) (retErr error) {
	data, err := pb.Marshal()
	if err != nil {
		return err
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.MaxInterval = 10 * time.Second
	return backoff.RetryNotify(func() error {
		return s.writeInternal(path, data)
	}, b, func(err error, duration time.Duration) error {
		logrus.Errorf("coult not write proto: %v, retrying in %v", err, duration)
		return nil
	})
}

// writeInternal contains the essential implementation of writeProto ('data' is
// a serialized proto), but does not retry
func (s *objBlockAPIServer) writeInternal(path string, data []byte) (retErr error) {
	defer func() {
		if retErr != nil {
			return
		}
		retErr = func() (retErr error) {
			if !s.objClient.Exists(path) {
				logrus.Errorf("%s doesn't exist after write", path)
				return fmt.Errorf("%s doesn't exist after write", path)
			}
			return nil
		}()
	}()
	w, err := s.objClient.Writer(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err = w.Write(data)
	return err
}

func (s *objBlockAPIServer) blockGetter(ctx groupcache.Context, key string, dest groupcache.Sink) (retErr error) {
	fields := strings.Split(key, blockKeySeparator)
	if len(fields) != 3 {
		return fmt.Errorf("bad block key: %s", key)
	}
	lower, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return err
	}
	upper, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return err
	}
	return s.readObj(s.blockPath(client.NewBlock(fields[0])), lower, upper-lower, dest)
}

func (s *objBlockAPIServer) objectGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	objectInfo := &pfsclient.ObjectInfo{}
	sink := groupcache.ProtoSink(objectInfo)
	if err := s.objectInfoCache.Get(ctx, key, sink); err != nil {
		return err
	}
	return s.readBlockRef(objectInfo.BlockRef, dest)
}

func (s *objBlockAPIServer) tagGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	splitKey := strings.Split(key, ".")
	tag := &pfsclient.Tag{Name: strings.Join(splitKey[:len(splitKey)-1], "")}
	prefix := splitKey[0]
	var updated bool
	// First check if we already have the index for this Tag in memory, if
	// not read it for the first time.
	if _, ok := s.getObjectIndex(prefix); !ok {
		updated = true
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
	}
	objectIndex, _ := s.getObjectIndex(prefix)
	// Check if the index contains the tag we're looking for, if so read
	// it into the cache and return
	if object, ok := objectIndex.Tags[tag.Name]; ok {
		dest.SetProto(object)
		return nil
	}
	// Try reading the tag from its tag path, this happens for recently
	// written tags that haven't been incorporated into an index yet.
	// Note that we tolerate NotExist errors here because the object may have
	// been incorporated into an index and thus deleted.
	objectIndex = &pfsclient.ObjectIndex{}
	if err := s.readProto(s.tagPath(tag), objectIndex); err != nil && !s.isNotFoundErr(err) {
		return err
	} else if err == nil {
		if object, ok := objectIndex.Tags[tag.Name]; ok {
			dest.SetProto(object)
			return nil
		}
	}
	// The last chance to find this object is to update the index since the
	// object may have been recently incorporated into it.
	if len(splitKey) == 3 && !updated {
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
		objectIndex, _ = s.getObjectIndex(prefix)
		if object, ok := objectIndex.Tags[tag.Name]; ok {
			dest.SetProto(object)
			return nil
		}
	}
	return fmt.Errorf("tagGetter: tag %s not found", tag.Name)
}

func (s *objBlockAPIServer) objectInfoGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	splitKey := strings.Split(key, ".")
	if len(splitKey) != 3 {
		return fmt.Errorf("invalid key %s (this is likely a bug)", key)
	}
	prefix := splitKey[0]
	object := &pfsclient.Object{Hash: strings.Join(splitKey[:len(splitKey)-1], "")}
	result := &pfsclient.ObjectInfo{Object: object}
	updated := false
	// First check if we already have the index for this Object in memory, if
	// not read it for the first time.
	if _, ok := s.getObjectIndex(prefix); !ok {
		updated = true
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
	}
	objectIndex, _ := s.getObjectIndex(prefix)
	// Check if the index contains a the object we're looking for, if so read
	// it into the cache and return
	if blockRef, ok := objectIndex.Objects[object.Hash]; ok {
		result.BlockRef = blockRef
		dest.SetProto(result)
		return nil
	}
	// Try reading the object from its object path, this happens for recently
	// written objects that haven't been incorporated into an index yet.
	// Note that we tolerate NotExist errors here because the object may have
	// been incorporated into an index and thus deleted.
	blockRef := &pfsclient.BlockRef{}
	if err := s.readProto(s.objectPath(object), blockRef); err != nil && !s.isNotFoundErr(err) {
		return err
	} else if err == nil {
		result.BlockRef = blockRef
		dest.SetProto(result)
		return nil
	}
	// The last chance to find this object is to update the index since the
	// object may have been recently incorporated into it.
	if !updated {
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
		objectIndex, _ := s.getObjectIndex(prefix)
		if blockRef, ok := objectIndex.Objects[object.Hash]; ok {
			result.BlockRef = blockRef
			dest.SetProto(result)
			return nil
		}
	}
	return fmt.Errorf("objectInfoGetter: object %s not found", object.Hash)
}

func (s *objBlockAPIServer) readObj(path string, offset uint64, size uint64, dest groupcache.Sink) (retErr error) {
	var reader io.ReadCloser
	var err error
	backoff.RetryNotify(func() error {
		reader, err = s.objClient.Reader(path, offset, size)
		if err != nil && obj.IsRetryable(s.objClient, err) {
			return err
		}
		return nil
	}, obj.NewExponentialBackOffConfig(), func(err error, d time.Duration) error {
		logrus.Infof("Error creating reader; retrying in %s: %#v", d, obj.RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
		})
		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	buff := make([]byte, size)
	_, err = io.ReadFull(reader, buff)
	if err != nil {
		return err
	}
	return dest.SetBytes(buff)
}

func (s *objBlockAPIServer) readBlockRef(blockRef *pfsclient.BlockRef, dest groupcache.Sink) error {
	return s.readObj(s.blockPath(blockRef.Block), blockRef.Range.Lower, blockRef.Range.Upper-blockRef.Range.Lower, dest)
}

func (s *objBlockAPIServer) getObjectIndex(prefix string) (*pfsclient.ObjectIndex, bool) {
	s.objectIndexesLock.RLock()
	defer s.objectIndexesLock.RUnlock()
	index, ok := s.objectIndexes[prefix]
	return index, ok
}

func (s *objBlockAPIServer) setObjectIndex(prefix string, index *pfsclient.ObjectIndex) {
	s.objectIndexesLock.Lock()
	defer s.objectIndexesLock.Unlock()
	s.objectIndexes[prefix] = index
}

func (s *objBlockAPIServer) readObjectIndex(prefix string) error {
	objectIndex := &pfsclient.ObjectIndex{}
	if err := s.readProto(s.indexPath(prefix), objectIndex); err != nil && !s.isNotFoundErr(err) {
		return err
	}
	// Note that we only return the error above if it's something other than a
	// NonExist error, in the case of a NonExist error we'll put a blank index
	// in the map. This prevents us from having requesting an index that
	// doesn't exist everytime a request tries to access it.
	s.setObjectIndex(prefix, objectIndex)
	return nil
}

// splitKey splits a key into the format we want, and also postpends
// the generation number
func (s *objBlockAPIServer) splitKey(key string) string {
	gen := s.getGeneration()
	if len(key) < prefixLength {
		return fmt.Sprintf("%s.%d", key, gen)
	}
	return fmt.Sprintf("%s.%s.%d", key[:prefixLength], key[prefixLength:], gen)
}

type blockWriter struct {
	w       io.WriteCloser
	block   *pfsclient.Block
	written uint64
	mu      sync.Mutex
}

func (s *objBlockAPIServer) newBlockWriter(block *pfsclient.Block) (*blockWriter, error) {
	w, err := s.objClient.Writer(s.blockPath(block))
	if err != nil {
		return nil, err
	}
	return &blockWriter{
		w:     w,
		block: block,
	}, nil
}

func (w *blockWriter) Write(p []byte) (*pfsclient.BlockRef, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.w.Write(p); err != nil {
		return nil, err
	}
	lower := w.written
	w.written += uint64(len(p))
	return &pfsclient.BlockRef{
		Block: w.block,
		Range: &pfsclient.ByteRange{
			Lower: lower,
			Upper: w.written,
		}}, nil
}

func (w *blockWriter) Close() error {
	return w.w.Close()
}

func (s *objBlockAPIServer) blockDir() string {
	return filepath.Join(s.dir, "block")
}

func (s *objBlockAPIServer) blockPath(block *pfsclient.Block) string {
	return filepath.Join(s.blockDir(), block.Hash)
}

func (s *objBlockAPIServer) objectDir() string {
	return filepath.Join(s.dir, "object")
}

func (s *objBlockAPIServer) objectPath(object *pfsclient.Object) string {
	return filepath.Join(s.objectDir(), object.Hash)
}

func (s *objBlockAPIServer) tagDir() string {
	return filepath.Join(s.dir, "tag")
}

func (s *objBlockAPIServer) tagPath(tag *pfsclient.Tag) string {
	return filepath.Join(s.tagDir(), tag.Name)
}

func (s *objBlockAPIServer) indexDir() string {
	return filepath.Join(s.dir, "index")
}

func (s *objBlockAPIServer) indexPath(prefix string) string {
	return filepath.Join(s.indexDir(), prefix)
}

type putObjectServer interface {
	Recv() (*pfsclient.PutObjectRequest, error)
}

type putObjectReader struct {
	server putObjectServer
	buffer bytes.Buffer
	tags   []*pfsclient.Tag
}

func (r *putObjectReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		r.buffer.Reset()
		// buffer.Write cannot error
		r.buffer.Write(request.Value)
		r.tags = append(r.tags, request.Tags...)
	}
	return r.buffer.Read(p)
}

func drainObjectServer(putObjectServer putObjectServer) {
	for {
		if _, err := putObjectServer.Recv(); err != nil {
			break
		}
	}
}
