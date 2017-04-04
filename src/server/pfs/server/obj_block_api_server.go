package server

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	protolion "go.pedge.io/lion"
	protorpclog "go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/cenkalti/backoff"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/groupcache"
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
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
	protorpclog.Logger
	dir               string
	localServer       *localBlockAPIServer
	objClient         obj.Client
	blockCache        *groupcache.Group
	objectCache       *groupcache.Group
	tagCache          *groupcache.Group
	objectInfoCache   *groupcache.Group
	objectIndexes     map[string]*pfsclient.ObjectIndex
	objectIndexesLock sync.RWMutex
	objectCacheBytes  int64
}

func newObjBlockAPIServer(dir string, cacheBytes int64, objClient obj.Client) (*objBlockAPIServer, error) {
	// defensive mesaure incase IsNotExist checking breaks due to underlying changes
	if err := obj.TestIsNotExist(objClient); err != nil {
		return nil, err
	}
	localServer, err := newLocalBlockAPIServer(dir)
	if err != nil {
		return nil, err
	}
	oneCacheShare := cacheBytes / (objectCacheShares + tagCacheShares + objectInfoCacheShares)
	server := &objBlockAPIServer{
		Logger:           protorpclog.NewLogger("pfs.BlockAPI.Obj"),
		dir:              dir,
		localServer:      localServer,
		objClient:        objClient,
		objectIndexes:    make(map[string]*pfsclient.ObjectIndex),
		objectCacheBytes: oneCacheShare * objectCacheShares,
	}
	server.blockCache = groupcache.NewGroup("block", cacheBytes,
		groupcache.GetterFunc(server.blockGetter))
	server.objectCache = groupcache.NewGroup("object", oneCacheShare*objectCacheShares,
		groupcache.GetterFunc(server.objectGetter))
	server.tagCache = groupcache.NewGroup("tag", oneCacheShare*tagCacheShares,
		groupcache.GetterFunc(server.tagGetter))
	server.objectInfoCache = groupcache.NewGroup("objectInfo", oneCacheShare*objectInfoCacheShares,
		groupcache.GetterFunc(server.objectInfoGetter))
	return server, nil
}

func newMinioBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	objClient, err := obj.NewMinioClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, objClient)
}

func newAmazonBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	objClient, err := obj.NewAmazonClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, objClient)
}

func newGoogleBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	objClient, err := obj.NewGoogleClientFromSecret(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, objClient)
}

func newMicrosoftBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	objClient, err := obj.NewMicrosoftClientFromSecret("")
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, objClient)
}

func (s *objBlockAPIServer) PutBlock(putBlockServer pfsclient.BlockAPI_PutBlockServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	result := &pfsclient.BlockRefs{}
	defer func(start time.Time) { s.Log(nil, result, retErr, time.Since(start)) }(time.Now())
	defer drainBlockServer(putBlockServer)
	putBlockRequest, err := putBlockServer.Recv()
	if err != nil {
		if err != io.EOF {
			return err
		}
		return putBlockServer.SendAndClose(result)
	}
	reader := bufio.NewReader(&putBlockReader{
		server: putBlockServer,
		buffer: bytes.NewBuffer(putBlockRequest.Value),
	})
	var eg errgroup.Group
	decoder := json.NewDecoder(reader)
	for {
		blockRef, data, err := readBlock(putBlockRequest.Delimiter, reader, decoder)
		if err != nil {
			return err
		}
		result.BlockRef = append(result.BlockRef, blockRef)
		eg.Go(func() error {
			var outerErr error
			path := s.localServer.blockPath(blockRef.Block)
			backoff.RetryNotify(func() error {
				if err := func() error {
					// We don't want to overwrite blocks that already exist, since:
					// 1) blocks are content-addressable, so it will be the same block
					// 2) we risk exceeding the object store's rate limit
					if s.objClient.Exists(path) {
						return nil
					}
					writer, err := s.objClient.Writer(path)
					if err != nil {
						return err
					}
					if _, err := writer.Write(data); err != nil {
						return err
					}
					if err := writer.Close(); err != nil {
						return err
					}
					return nil
				}(); err != nil {
					if obj.IsRetryable(s.objClient, err) {
						return err
					}
					outerErr = err
					return nil
				}
				return nil
			}, obj.NewExponentialBackOffConfig(), func(err error, d time.Duration) {
				protolion.Infof("Error writing; retrying in %s: %#v", d, obj.RetryError{
					Err:               err.Error(),
					TimeTillNextRetry: d.String(),
				})
			})
			// Weird effects can happen with clients racing. Ultimately if the
			// path exists then it doesn't make sense to consider this
			// operation as having errored because we know that it contains the
			// data we want thanks to content addressing.
			if outerErr != nil && !s.objClient.Exists(path) {
				return outerErr
			}
			return nil
		})
		if (blockRef.Range.Upper - blockRef.Range.Lower) < uint64(blockSize) {
			break
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return putBlockServer.SendAndClose(result)
}

func (s *objBlockAPIServer) GetBlock(request *pfsclient.GetBlockRequest, getBlockServer pfsclient.BlockAPI_GetBlockServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	var data []byte
	sink := groupcache.AllocatingByteSliceSink(&data)
	if err := s.blockCache.Get(getBlockServer.Context(), request.Block.Hash, sink); err != nil {
		return err
	}
	if request.SizeBytes != 0 && request.SizeBytes+request.OffsetBytes < uint64(len(data)) {
		data = data[request.OffsetBytes : request.OffsetBytes+request.SizeBytes]
	} else if request.OffsetBytes < uint64(len(data)) {
		data = data[request.OffsetBytes:]
	} else {
		data = nil
	}
	return getBlockServer.Send(&types.BytesValue{Value: data})
}

func (s *objBlockAPIServer) DeleteBlock(ctx context.Context, request *pfsclient.DeleteBlockRequest) (response *types.Empty, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	backoff.RetryNotify(func() error {
		if err := s.objClient.Delete(s.localServer.blockPath(request.Block)); err != nil && !s.objClient.IsNotExist(err) {
			return err
		}
		return nil
	}, obj.NewExponentialBackOffConfig(), func(err error, d time.Duration) {
		protolion.Infof("Error deleting block; retrying in %s: %#v", d, obj.RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
		})
	})
	return &types.Empty{}, nil
}

func (s *objBlockAPIServer) InspectBlock(ctx context.Context, request *pfsclient.InspectBlockRequest) (response *pfsclient.BlockInfo, retErr error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *objBlockAPIServer) ListBlock(ctx context.Context, request *pfsclient.ListBlockRequest) (response *pfsclient.BlockInfos, retErr error) {
	return nil, fmt.Errorf("not implemented")
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
	if err := func() error {
		w, err := s.objClient.Writer(s.localServer.blockPath(block))
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		size, err = io.CopyBuffer(w, r, make([]byte, bufferSize))
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
	var blockRef *pfsclient.BlockRef
	// Now that we have a hash of the object we can check if it already exists.
	objectInfo, err := s.InspectObject(server.Context(), object)
	if err == nil {
		// the object already exists so we delete the block we put
		eg.Go(func() error {
			return s.objClient.Delete(s.localServer.blockPath(block))
		})
		blockRef = objectInfo.BlockRef
	} else {
		blockRef = &pfsclient.BlockRef{
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
		tag := hashTag(tag)
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
	if (objectSize) > uint64(s.objectCacheBytes/maxCachedObjectDenom) {
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
	if err := s.objectCache.Get(getObjectServer.Context(), splitKey(request.Hash), sink); err != nil {
		return err
	}
	return getObjectServer.Send(&types.BytesValue{Value: data})
}

func (s *objBlockAPIServer) TagObject(ctx context.Context, request *pfsclient.TagObjectRequest) (response *types.Empty, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	// First inspect the object to make sure it actually exists
	if _, err := s.InspectObject(ctx, request.Object); err != nil {
		return nil, err
	}
	var eg errgroup.Group
	for _, tag := range request.Tags {
		tag := hashTag(tag)
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
	if err := s.objectInfoCache.Get(ctx, splitKey(request.Hash), sink); err != nil {
		return nil, err
	}
	return objectInfo, nil
}

func (s *objBlockAPIServer) GetTag(request *pfsclient.Tag, getTagServer pfsclient.ObjectAPI_GetTagServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	object := &pfsclient.Object{}
	sink := groupcache.ProtoSink(object)
	if err := s.tagCache.Get(getTagServer.Context(), splitKey(hashTag(request).Name), sink); err != nil {
		return err
	}
	return s.GetObject(object, getTagServer)
}

func (s *objBlockAPIServer) InspectTag(ctx context.Context, request *pfsclient.Tag) (response *pfsclient.ObjectInfo, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	object := &pfsclient.Object{}
	sink := groupcache.ProtoSink(object)
	if err := s.tagCache.Get(ctx, splitKey(hashTag(request).Name), sink); err != nil {
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
		prefixes[tag[:2]] = true
	}
	eg = errgroup.Group{}
	for prefix := range prefixes {
		prefix := prefix
		eg.Go(func() error {
			prefixObjectIndex := &pfsclient.ObjectIndex{
				Objects: make(map[string]*pfsclient.BlockRef),
				Tags:    make(map[string]*pfsclient.Object),
			}
			if err := s.readProto(s.localServer.indexPath(prefix), prefixObjectIndex); err != nil && !s.objClient.IsNotExist(err) {
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

func (s *objBlockAPIServer) readProto(path string, pb proto.Message) (retErr error) {
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
	return proto.Unmarshal(data, pb)
}

func (s *objBlockAPIServer) writeProto(path string, pb proto.Message) (retErr error) {
	w, err := s.objClient.Writer(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	data, err := proto.Marshal(pb)
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
	if len(splitKey) != 2 {
		return fmt.Errorf("invalid key %s (this is likely a bug)", key)
	}
	prefix := splitKey[0]
	tag := &pfsclient.Tag{Name: strings.Join(splitKey, "")}
	updated := false
	// First check if we already have the index for this Tag in memory, if
	// not read it for the first time.
	if _, ok := s.objectIndexes[prefix]; !ok {
		updated = true
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
	}
	// Check if the index contains the tag we're looking for, if so read
	// it into the cache and return
	if object, ok := s.objectIndexes[prefix].Tags[tag.Name]; ok {
		dest.SetProto(object)
		return nil
	}
	// Try reading the tag from its tag path, this happens for recently
	// written tags that haven't been incorporated into an index yet.
	// Note that we tolerate NotExist errors here because the object may have
	// been incorporated into an index and thus deleted.
	objectIndex := &pfsclient.ObjectIndex{}
	if err := s.readProto(s.localServer.tagPath(tag), objectIndex); err != nil && !s.objClient.IsNotExist(err) {
		return err
	} else if err == nil {
		if object, ok := objectIndex.Tags[tag.Name]; ok {
			dest.SetProto(object)
			return nil
		}
	}
	// The last chance to find this object is to update the index since the
	// object may have been recently incorporated into it.
	if !updated {
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
		if object, ok := s.objectIndexes[prefix].Tags[tag.Name]; ok {
			dest.SetProto(object)
			return nil
		}
	}
	return fmt.Errorf("tagGetter: tag %s not found", tag.Name)
}

func (s *objBlockAPIServer) objectInfoGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	splitKey := strings.Split(key, ".")
	if len(splitKey) != 2 {
		return fmt.Errorf("invalid key %s (this is likely a bug)", key)
	}
	prefix := splitKey[0]
	object := &pfsclient.Object{Hash: strings.Join(splitKey, "")}
	result := &pfsclient.ObjectInfo{Object: object}
	updated := false
	// First check if we already have the index for this Object in memory, if
	// not read it for the first time.
	if _, ok := s.objectIndexes[prefix]; !ok {
		updated = true
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
	}
	// Check if the index contains a the object we're looking for, if so read
	// it into the cache and return
	if blockRef, ok := s.objectIndexes[prefix].Objects[object.Hash]; ok {
		result.BlockRef = blockRef
		dest.SetProto(result)
		return nil
	}
	// Try reading the object from its object path, this happens for recently
	// written objects that haven't been incorporated into an index yet.
	// Note that we tolerate NotExist errors here because the object may have
	// been incorporated into an index and thus deleted.
	blockRef := &pfsclient.BlockRef{}
	if err := s.readProto(s.localServer.objectPath(object), blockRef); err != nil && !s.objClient.IsNotExist(err) {
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
		if blockRef, ok := s.objectIndexes[prefix].Objects[object.Hash]; ok {
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
	}, obj.NewExponentialBackOffConfig(), func(err error, d time.Duration) {
		protolion.Infof("Error creating reader; retrying in %s: %#v", d, obj.RetryError{
			Err:               err.Error(),
			TimeTillNextRetry: d.String(),
		})
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

func (s *objBlockAPIServer) readObjectIndex(prefix string) error {
	objectIndex := &pfsclient.ObjectIndex{}
	if err := s.readProto(s.localServer.indexPath(prefix), objectIndex); err != nil && !s.objClient.IsNotExist(err) {
		return err
	}
	// Note that we only return the error above if it's something other than a
	// NonExist error, in the case of a NonExist error we'll put a blank index
	// in the map. This prevents us from having requesting an index that
	// doesn't exist everytime a request tries to access it.
	s.objectIndexesLock.Lock()
	defer s.objectIndexesLock.Unlock()
	s.objectIndexes[prefix] = objectIndex
	return nil
}

func splitKey(key string) string {
	if len(key) < prefixLength {
		return fmt.Sprintf("%s.", key)
	}
	return fmt.Sprintf("%s.%s", key[:prefixLength], key[prefixLength:])
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

func hashTag(tag *pfsclient.Tag) *pfsclient.Tag {
	hash := newHash()
	// writing to a hasher can't fail
	hash.Write([]byte(tag.Name))
	return &pfsclient.Tag{Name: hex.EncodeToString(hash.Sum(nil))}
}
