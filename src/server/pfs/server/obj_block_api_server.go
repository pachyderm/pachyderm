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
	"sync/atomic"
	"time"

	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/cenkalti/backoff"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/groupcache"
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	prefixLength        = 2
	compactionThreshold = 100
	cacheSize           = 1024 * 1024 * 1024 * 10 // 10 Gigabytes
	alphabet            = "0123456789abcdef"
)

type objBlockAPIServer struct {
	protorpclog.Logger
	dir               string
	localServer       *localBlockAPIServer
	objClient         obj.Client
	blockCache        *groupcache.Group
	objectCache       *groupcache.Group
	tagCache          *groupcache.Group
	objectIndexes     map[string]*pfsclient.ObjectIndex
	objectIndexesLock sync.RWMutex
}

func newObjBlockAPIServer(dir string, cacheBytes int64, objClient obj.Client) (*objBlockAPIServer, error) {
	localServer, err := newLocalBlockAPIServer(dir)
	if err != nil {
		return nil, err
	}
	server := &objBlockAPIServer{
		Logger:        protorpclog.NewLogger("pfs.BlockAPI.Obj"),
		dir:           dir,
		localServer:   localServer,
		objClient:     objClient,
		objectIndexes: make(map[string]*pfsclient.ObjectIndex),
	}
	server.blockCache = groupcache.NewGroup("block", cacheSize,
		groupcache.GetterFunc(server.blockGetter))
	server.objectCache = groupcache.NewGroup("object", cacheSize,
		groupcache.GetterFunc(server.objectGetter))
	server.tagCache = groupcache.NewGroup("tag", cacheSize,
		groupcache.GetterFunc(server.tagGetter))
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for {
			<-ticker.C
			if err := server.compact(); err != nil {
				protolion.Errorf("error compacting: %s", err.Error())
			}
		}
	}()
	return server, nil
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
	result := &pfsclient.BlockRefs{}
	func() { s.Log(nil, nil, nil, 0) }()
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
		eg.Go(func() (retErr error) {
			backoff.RetryNotify(func() error {
				path := s.localServer.blockPath(blockRef.Block)
				// We don't want to overwrite blocks that already exist, since:
				// 1) blocks are content-addressable, so it will be the same block
				// 2) we risk exceeding the object store's rate limit
				if s.objClient.Exists(path) {
					return nil
				}
				writer, err := s.objClient.Writer(path)
				if err != nil {
					retErr = err
					return nil
				}
				if _, err := writer.Write(data); err != nil {
					retErr = err
					return nil
				}
				if err := writer.Close(); err != nil {
					if s.objClient.IsRetryable(err) {
						return err
					}
					retErr = err
					return nil
				}
				return nil
			}, obj.NewExponentialBackOffConfig(), func(err error, d time.Duration) {
				protolion.Infof("Error writing; retrying in %s: %#v", d, obj.RetryError{
					Err:               err.Error(),
					TimeTillNextRetry: d.String(),
				})
			})
			return
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
	return getBlockServer.Send(&google_protobuf.BytesValue{Value: data})
}

func (s *objBlockAPIServer) DeleteBlock(ctx context.Context, request *pfsclient.DeleteBlockRequest) (response *google_protobuf.Empty, retErr error) {
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
	return google_protobuf.EmptyInstance, nil
}

func (s *objBlockAPIServer) InspectBlock(ctx context.Context, request *pfsclient.InspectBlockRequest) (response *pfsclient.BlockInfo, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	return nil, fmt.Errorf("not implemented")
}

func (s *objBlockAPIServer) ListBlock(ctx context.Context, request *pfsclient.ListBlockRequest) (response *pfsclient.BlockInfos, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return nil, fmt.Errorf("not implemented")
}

func (s *objBlockAPIServer) PutObject(ctx context.Context, request *pfsclient.PutObjectRequest) (response *pfsclient.Object, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	hash := newHash()
	if _, err := hash.Write(request.Value); err != nil {
		return nil, err
	}
	object := &pfsclient.Object{Hash: hex.EncodeToString(hash.Sum(nil))}
	var eg errgroup.Group
	eg.Go(func() (retErr error) {
		objectPath := s.localServer.objectPath(object)
		w, err := s.objClient.Writer(objectPath)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		if _, err := w.Write(request.Value); err != nil {
			return err
		}
		return nil
	})
	for _, tag := range request.Tags {
		tag := tag
		eg.Go(func() (retErr error) {
			index := &pfsclient.ObjectIndex{Tags: map[string]*pfsclient.Object{tag.Name: object}}
			return s.writeProto(s.localServer.tagPath(tag), index)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return object, nil
}

func (s *objBlockAPIServer) GetObject(ctx context.Context, request *pfsclient.Object) (response *google_protobuf.BytesValue, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var data []byte
	sink := groupcache.AllocatingByteSliceSink(&data)
	if err := s.objectCache.Get(ctx, splitKey(request.Hash), sink); err != nil {
		return nil, err
	}
	return &google_protobuf.BytesValue{Value: data}, nil
}

func (s *objBlockAPIServer) GetTag(ctx context.Context, request *pfsclient.Tag) (response *google_protobuf.BytesValue, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	object := &pfsclient.Object{}
	sink := groupcache.ProtoSink(object)
	if err := s.tagCache.Get(ctx, splitKey(request.Name), sink); err != nil {
		return nil, err
	}
	return s.GetObject(ctx, object)
}

func (s *objBlockAPIServer) objectPrefix(prefix string) string {
	return s.localServer.objectPath(&pfsclient.Object{Hash: prefix})
}

func (s *objBlockAPIServer) tagPrefix(prefix string) string {
	return s.localServer.tagPath(&pfsclient.Tag{Name: prefix})
}

func (s *objBlockAPIServer) compact() error {
	lion.Printf("compact\n")
	var eg errgroup.Group
	for i := 0; i < len(alphabet); i++ {
		for j := 0; j < len(alphabet); j++ {
			prefix := fmt.Sprintf("%c%c", alphabet[i], alphabet[j])
			eg.Go(func() error {
				count, err := s.countPrefix(prefix)
				lion.Printf("prefix %s: %d", prefix, count)
				if err != nil {
					return err
				}
				if count < compactionThreshold {
					return nil
				}
				return s.compactPrefix(prefix)
			})
		}
	}
	return eg.Wait()
}

func (s *objBlockAPIServer) countPrefix(prefix string) (int64, error) {
	var count int64
	var eg errgroup.Group
	eg.Go(func() error {
		return s.objClient.Walk(s.objectPrefix(prefix), func(name string) error {
			atomic.AddInt64(&count, 1)
			return nil
		})
	})
	eg.Go(func() error {
		return s.objClient.Walk(s.tagPrefix(prefix), func(name string) error {
			atomic.AddInt64(&count, 1)
			return nil
		})
	})
	if err := eg.Wait(); err != nil {
		return 0, nil
	}
	return count, nil
}

func (s *objBlockAPIServer) compactPrefix(prefix string) (retErr error) {
	var mu sync.Mutex
	var eg errgroup.Group
	objectIndex := &pfsclient.ObjectIndex{
		Objects: make(map[string]*pfsclient.BlockRef),
		Tags:    make(map[string]*pfsclient.Object),
	}
	var toDelete []string
	if err := s.readProto(s.localServer.indexPath(prefix), objectIndex); err != nil && !s.objClient.IsNotExist(err) {
		return err
	}
	block := &pfsclient.Block{Hash: uuid.NewWithoutDashes()}
	blockW, err := s.objClient.Writer(s.localServer.blockPath(block))
	if err != nil {
		return err
	}
	defer func() {
		if err := blockW.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var written uint64
	eg.Go(func() error {
		return s.objClient.Walk(s.objectPrefix(prefix), func(name string) error {
			eg.Go(func() (retErr error) {
				r, err := s.objClient.Reader(name, 0, 0)
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
				mu.Lock()
				defer mu.Unlock()
				if _, err := blockW.Write(object); err != nil {
					return err
				}
				objectIndex.Objects[filepath.Base(name)] = &pfsclient.BlockRef{
					Block: block,
					Range: &pfsclient.ByteRange{
						Lower: written,
						Upper: written + uint64(len(object)),
					},
				}
				written += uint64(len(object))
				toDelete = append(toDelete, name)
				return nil
			})
			return nil
		})
	})
	eg.Go(func() error {
		return s.objClient.Walk(s.tagPrefix(prefix), func(name string) error {
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
	if err := s.writeProto(s.localServer.indexPath(prefix), objectIndex); err != nil {
		return err
	}
	eg = errgroup.Group{}
	for _, file := range toDelete {
		file := file
		eg.Go(func() error {
			lion.Printf("deleting: %s\n", file)
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
	splitKey := strings.Split(key, ".")
	if len(splitKey) != 2 {
		return fmt.Errorf("invalid key %s (this is likely a bug)", key)
	}
	prefix := splitKey[0]
	object := &pfsclient.Object{Hash: strings.Join(splitKey, "")}
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
		return s.readObj(s.localServer.blockPath(blockRef.Block), blockRef.Range.Lower, blockRef.Range.Upper-blockRef.Range.Lower, dest)
	}
	// Try reading the object from its object path, this happens for recently
	// written objects that haven't been incorporated into an index yet.
	// Note that we tolerate NotExist errors here because the object may have
	// been incorporated into an index and thus deleted.
	if err := s.readObj(s.localServer.objectPath(object), 0, 0, dest); err != nil && !s.objClient.IsNotExist(err) {
		return err
	} else if err == nil {
		// We found the file so we can return.
		return nil
	}
	// The last chance to find this object is to update the index since the
	// object may have been recently incorporated into it.
	if !updated {
		if err := s.readObjectIndex(prefix); err != nil {
			return err
		}
		if blockRef, ok := s.objectIndexes[prefix].Objects[object.Hash]; ok {
			return s.readObj(s.localServer.blockPath(blockRef.Block), blockRef.Range.Lower, blockRef.Range.Upper-blockRef.Range.Lower, dest)
		}
	}
	return fmt.Errorf("objectGetter: object %s not found", object.Hash)
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

func (s *objBlockAPIServer) readObj(path string, offset uint64, size uint64, dest groupcache.Sink) (retErr error) {
	var reader io.ReadCloser
	var err error
	backoff.RetryNotify(func() error {
		reader, err = s.objClient.Reader(path, offset, size)
		if err != nil && s.objClient.IsRetryable(err) {
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
