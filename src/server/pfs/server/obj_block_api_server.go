package server

import (
	"bytes"
	"encoding/hex"
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
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	prefixLength          = 2
	alphabet              = "0123456789abcdef"
	objectCacheShares     = 8
	tagCacheShares        = 1
	objectInfoCacheShares = 1
	maxCachedObjectDenom  = 4                // We will only cache objects less than 1/maxCachedObjectDenom of total cache size
	bufferSize            = 15 * 1024 * 1024 // 15 MB
)

type objBlockAPIServer struct {
	log.Logger
	dir         string
	localServer *localBlockAPIServer
	objClient   obj.Client

	// cache
	objectCache     *groupcache.Group
	tagCache        *groupcache.Group
	objectInfoCache *groupcache.Group
	// The total number of bytes cached for objects
	objectCacheBytes int64
	// The GC generation number.  Incrementing this number effectively
	// invalidates all current cache.
	generation int
	genLock    sync.RWMutex

	objectIndexes     map[string]*pfsclient.ObjectIndex
	objectIndexesLock sync.RWMutex
}

func newObjBlockAPIServer(dir string, cacheBytes int64, etcdAddress string, objClient obj.Client) (*objBlockAPIServer, error) {
	// defensive mesaure incase IsNotExist checking breaks due to underlying changes
	if err := obj.TestIsNotExist(objClient); err != nil {
		return nil, err
	}
	localServer, err := newLocalBlockAPIServer(dir)
	if err != nil {
		return nil, err
	}
	oneCacheShare := cacheBytes / (objectCacheShares + tagCacheShares + objectInfoCacheShares)
	s := &objBlockAPIServer{
		Logger:           log.NewLogger("pfs.BlockAPI.Obj"),
		dir:              dir,
		localServer:      localServer,
		objClient:        objClient,
		objectIndexes:    make(map[string]*pfsclient.ObjectIndex),
		objectCacheBytes: oneCacheShare * objectCacheShares,
	}
	s.objectCache = groupcache.NewGroup("object", oneCacheShare*objectCacheShares, groupcache.GetterFunc(s.objectGetter))
	s.tagCache = groupcache.NewGroup("tag", oneCacheShare*tagCacheShares, groupcache.GetterFunc(s.tagGetter))
	s.objectInfoCache = groupcache.NewGroup("objectInfo", oneCacheShare*objectInfoCacheShares, groupcache.GetterFunc(s.objectInfoGetter))
	// Periodically print cache stats for debugging purposes
	// TODO: make the stats accessible via HTTP or gRPC.
	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			logrus.Infof("objectCache stats: %+v", s.objectCache.Stats)
			logrus.Infof("tagCache stats: %+v", s.tagCache.Stats)
			logrus.Infof("objectInfoCache stats: %+v", s.objectInfoCache.Stats)
		}
	}()
	go s.watchGC(etcdAddress)
	return s, nil
}

// watchGC watches for GC runs and invalidate all cache when GC happens.
func (s *objBlockAPIServer) watchGC(etcdAddress string) {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		etcdClient, err := etcd.New(etcd.Config{
			Endpoints:   []string{etcdAddress},
			DialOptions: client.EtcdDialOptions(),
		})
		if err != nil {
			return fmt.Errorf("error instantiating etcd client: %v", err)
		}

		watcher, err := watch.NewWatcher(context.Background(), etcdClient, client.GCGenerationKey)
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
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient)
}

func newAmazonBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewAmazonClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient)
}

func newGoogleBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewGoogleClientFromSecret(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient)
}

func newMicrosoftBlockAPIServer(dir string, cacheBytes int64, etcdAddress string) (*objBlockAPIServer, error) {
	objClient, err := obj.NewMicrosoftClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, etcdAddress, objClient)
}

func (s *objBlockAPIServer) PutObject(server pfsclient.ObjectAPI_PutObjectServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	defer drainObjectServer(server)
	hash := newHash()
	putObjectReader := &putObjectReader{
		server: server,
	}
	r := io.TeeReader(putObjectReader, hash)
	block := &pfsclient.Block{Hash: uuid.NewWithoutDashes()}
	var size int64
	if err := func() (retErr error) {
		w, err := s.objClient.Writer(s.localServer.blockPath(block))
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
		size, err = io.CopyBuffer(w, r, buf)
		if err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}
	object := &pfsclient.Object{Hash: hex.EncodeToString(hash.Sum(nil))}
	if err := server.SendAndClose(object); err != nil {
		return err
	}
	var eg errgroup.Group
	// Now that we have a hash of the object we can check if it already exists.
	resp, err := s.CheckObject(server.Context(), &pfsclient.CheckObjectRequest{object})
	if err != nil {
		return err
	}
	if resp.Exists {
		// the object already exists so we delete the block we put
		eg.Go(func() error {
			return s.objClient.Delete(s.localServer.blockPath(block))
		})
	} else {
		blockRef := &pfsclient.BlockRef{
			Block: block,
			Range: &pfsclient.ByteRange{
				Lower: 0,
				Upper: uint64(size),
			},
		}
		eg.Go(func() error {
			return s.writeProto(s.localServer.objectPath(object), blockRef)
		})
	}
	for _, tag := range putObjectReader.tags {
		tag := tag
		eg.Go(func() (retErr error) {
			index := &pfsclient.ObjectIndex{Tags: map[string]*pfsclient.Object{tag.Name: object}}
			return s.writeProto(s.localServer.tagPath(tag), index)
		})
	}
	return eg.Wait()
}

func (s *objBlockAPIServer) GetObject(request *pfsclient.Object, getObjectServer pfsclient.ObjectAPI_GetObjectServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	// First we inspect the object to see how big it is.
	objectInfo, err := s.InspectObject(getObjectServer.Context(), request)
	if err != nil {
		return err
	}
	objectSize := objectInfo.BlockRef.Range.Upper - objectInfo.BlockRef.Range.Lower
	if (objectSize) >= uint64(s.objectCacheBytes/maxCachedObjectDenom) {
		// The object is a substantial portion of the available cache space so
		// we bypass the cache and stream it directly out of the underlying store.
		blockPath := s.localServer.blockPath(objectInfo.BlockRef.Block)
		r, err := s.objClient.Reader(blockPath, objectInfo.BlockRef.Range.Lower, objectSize)
		if err != nil {
			return err
		}
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
			logrus.Debugf("objectInfo is nil; info: %+v; request: %v", objectInfo, request)
		} else if objectInfo.BlockRef == nil {
			logrus.Debugf("objectInfo.BlockRef is nil; info: %+v; request: %v", objectInfo, request)
		} else if objectInfo.BlockRef.Range == nil {
			logrus.Debugf("objectInfo.BlockRef.Range is nil; info: %+v; request: %v", objectInfo, request)
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
		if s.objectCacheBytes == 0 || (objectSize) > uint64(s.objectCacheBytes/maxCachedObjectDenom) {
			// The object is a substantial portion of the available cache space so
			// we bypass the cache and stream it directly out of the underlying store.
			blockPath := s.localServer.blockPath(objectInfo.BlockRef.Block)
			r, err := s.objClient.Reader(blockPath, objectInfo.BlockRef.Range.Lower+offset, readSize)
			if err != nil {
				return err
			}
			return grpcutil.WriteToStreamingBytesServer(r, getObjectsServer)
		}
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
	resp, err := s.CheckObject(ctx, &pfsclient.CheckObjectRequest{request.Object})
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
			return s.writeProto(s.localServer.tagPath(tag), index)
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
		Exists: s.objClient.Exists(s.localServer.objectPath(request.Object)),
	}, nil
}

func (s *objBlockAPIServer) ListObjects(request *pfsclient.ListObjectsRequest, listObjectsServer pfsclient.ObjectAPI_ListObjectsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return s.objClient.Walk(s.localServer.objectDir(), func(key string) error {
		return listObjectsServer.Send(&pfsclient.Object{filepath.Base(key)})
	})
}

func (s *objBlockAPIServer) ListTags(request *pfsclient.ListTagsRequest, server pfsclient.ObjectAPI_ListTagsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	var eg errgroup.Group
	limiter := limit.New(100)
	s.objClient.Walk(path.Join(s.localServer.tagDir(), request.Prefix), func(key string) error {
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
					server.Send(&pfsclient.ListTagsResponse{
						Tag:    tag,
						Object: object,
					})
				}
				return nil
			})
		}
		server.Send(&pfsclient.ListTagsResponse{
			Tag: tag,
		})
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
			tagPath := s.localServer.tagPath(&pfsclient.Tag{tag})
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

			objPath := s.localServer.objectPath(object)
			if err := s.objClient.Delete(objPath); err != nil && !s.isNotFoundErr(err) {
				return err
			}

			if objectInfo != nil && objectInfo.BlockRef != nil && objectInfo.BlockRef.Block != nil {
				blockPath := s.localServer.blockPath(objectInfo.BlockRef.Block)
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
	return s.localServer.objectPath(&pfsclient.Object{Hash: prefix})
}

func (s *objBlockAPIServer) tagPrefix(prefix string) string {
	return s.localServer.tagPath(&pfsclient.Tag{Name: prefix})
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
		return s.objClient.Walk(s.localServer.objectDir(), func(name string) error {
			eg.Go(func() (retErr error) {
				blockRef := &pfsclient.BlockRef{}
				if err := s.readProto(name, blockRef); err != nil {
					return err
				}
				blockPath := s.localServer.blockPath(blockRef.Block)
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
		return s.objClient.Walk(s.localServer.tagDir(), func(name string) error {
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
			if err := s.readProto(s.localServer.indexPath(prefix), prefixObjectIndex); err != nil && !s.isNotFoundErr(err) {
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
			return s.writeProto(s.localServer.indexPath(prefix), prefixObjectIndex)
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
	return pb.Unmarshal(data)
}

func (s *objBlockAPIServer) writeProto(path string, pb proto.Marshaler) (retErr error) {
	w, err := s.objClient.Writer(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	data, err := pb.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (s *objBlockAPIServer) blockGetter(ctx groupcache.Context, key string, dest groupcache.Sink) (retErr error) {
	return s.readObj(s.localServer.blockPath(client.NewBlock(key)), 0, 0, dest)
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
	if err := s.readProto(s.localServer.tagPath(tag), objectIndex); err != nil && !s.isNotFoundErr(err) {
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
	if err := s.readProto(s.localServer.objectPath(object), blockRef); err != nil && !s.isNotFoundErr(err) {
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
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return dest.SetBytes(data)
}

func (s *objBlockAPIServer) readBlockRef(blockRef *pfsclient.BlockRef, dest groupcache.Sink) error {
	return s.readObj(s.localServer.blockPath(blockRef.Block), blockRef.Range.Lower, blockRef.Range.Upper-blockRef.Range.Lower, dest)
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
	if err := s.readProto(s.localServer.indexPath(prefix), objectIndex); err != nil && !s.isNotFoundErr(err) {
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
	w, err := s.objClient.Writer(s.localServer.blockPath(block))
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
