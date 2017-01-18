package server

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"
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

type objBlockAPIServer struct {
	protorpclog.Logger
	dir         string
	localServer *localBlockAPIServer
	objClient   obj.Client
	cache       *groupcache.Group
}

func newObjBlockAPIServer(dir string, cacheBytes int64, objClient obj.Client) (*objBlockAPIServer, error) {
	localServer, err := newLocalBlockAPIServer(dir)
	if err != nil {
		return nil, err
	}
	server := &objBlockAPIServer{
		Logger:      protorpclog.NewLogger("pfs.BlockAPI.Obj"),
		dir:         dir,
		localServer: localServer,
		objClient:   objClient,
	}
	server.cache = groupcache.NewGroup("block", 1024*1024*1024*10,
		groupcache.GetterFunc(server.blockGetter))
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
	if err := s.cache.Get(getBlockServer.Context(), request.Block.Hash, sink); err != nil {
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
	object := &pfsclient.Object{Hash: base64.URLEncoding.EncodeToString(newHash().Sum(request.Value))}
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
	objectPath := s.localServer.objectPath(request)
	r, err := s.objClient.Reader(objectPath, 0, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	value, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &google_protobuf.BytesValue{Value: value}, nil
}

func (s *objBlockAPIServer) GetTag(ctx context.Context, request *pfsclient.Tag) (response *google_protobuf.BytesValue, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	objectIndex := &pfsclient.ObjectIndex{}
	if err := s.readProto(s.localServer.tagPath(request), objectIndex); err != nil {
		return nil, err
	}
	return s.GetObject(ctx, objectIndex.Tags[request.Name])
}

func (s *objBlockAPIServer) objectPrefix(prefix string) string {
	return s.localServer.objectPath(&pfsclient.Object{Hash: prefix})
}

func (s *objBlockAPIServer) tagPrefix(prefix string) string {
	return s.localServer.tagPath(&pfsclient.Tag{Name: prefix})
}

func (s *objBlockAPIServer) compactPrefix(ctx context.Context, prefix string) (retErr error) {
	var mu sync.Mutex
	var eg errgroup.Group
	objectIndex := &pfsclient.ObjectIndex{}
	if err := s.readProto(s.localServer.indexPath(prefix), objectIndex); err != nil {
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
		s.objClient.Walk(s.objectPrefix(prefix), func(name string) error {
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
				return nil
			})
			return nil
		})
		return nil
	})
	eg.Go(func() error {
		s.objClient.Walk(s.tagPrefix(prefix), func(name string) error {
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
				return nil
			})
			return nil
		})
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	return s.writeProto(s.localServer.indexPath(prefix), objectIndex)
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
	var reader io.ReadCloser
	var err error
	backoff.RetryNotify(func() error {
		reader, err = s.objClient.Reader(s.localServer.blockPath(client.NewBlock(key)), 0, 0)
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
	block, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return dest.SetBytes(block)
}
