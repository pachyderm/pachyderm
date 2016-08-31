package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	return &objBlockAPIServer{
		Logger:      protorpclog.NewLogger("pachyderm.pfsclient.objBlockAPIServer"),
		dir:         dir,
		localServer: localServer,
		objClient:   objClient,
		cache: groupcache.NewGroup("block", 1024*1024*1024*10,
			groupcache.GetterFunc(func(ctx groupcache.Context, key string, dest groupcache.Sink) (retErr error) {
				var reader io.ReadCloser
				var err error
				backoff.RetryNotify(func() error {
					reader, err = objClient.Reader(localServer.blockPath(client.NewBlock(key)), 0, 0)
					if err != nil && objClient.IsRetryable(err) {
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
			})),
	}, nil
}

func newAmazonBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	bucket, err := ioutil.ReadFile("/amazon-secret/bucket")
	if err != nil {
		return nil, err
	}
	id, err := ioutil.ReadFile("/amazon-secret/id")
	if err != nil {
		return nil, err
	}
	secret, err := ioutil.ReadFile("/amazon-secret/secret")
	if err != nil {
		return nil, err
	}
	token, err := ioutil.ReadFile("/amazon-secret/token")
	if err != nil {
		return nil, err
	}
	region, err := ioutil.ReadFile("/amazon-secret/region")
	if err != nil {
		return nil, err
	}
	objClient, err := obj.NewAmazonClient(string(bucket), string(id), string(secret), string(token), string(region))
	if err != nil {
		return nil, err
	}
	return newObjBlockAPIServer(dir, cacheBytes, objClient)
}

func newGoogleBlockAPIServer(dir string, cacheBytes int64) (*objBlockAPIServer, error) {
	bucket, err := ioutil.ReadFile("/google-secret/bucket")
	if err != nil {
		return nil, err
	}
	objClient, err := obj.NewGoogleClient(context.Background(), string(bucket))
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
			path := s.localServer.blockPath(blockRef.Block)
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
			defer func() {
				if err := writer.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if _, err := writer.Write(data); err != nil {
				return err
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

func (s *objBlockAPIServer) CreateDiff(ctx context.Context, request *pfsclient.DiffInfo) (response *google_protobuf.Empty, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	writer, err := s.objClient.Writer(s.localServer.diffPath(request.Diff))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (s *objBlockAPIServer) InspectDiff(ctx context.Context, request *pfsclient.InspectDiffRequest) (response *pfsclient.DiffInfo, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return s.readDiff(request.Diff)
}

func (s *objBlockAPIServer) ListDiff(request *pfsclient.ListDiffRequest, listDiffServer pfsclient.BlockAPI_ListDiffServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	if err := s.objClient.Walk(s.localServer.diffDir(), func(path string) error {
		diff := s.localServer.pathToDiff(path)
		if diff == nil {
			return fmt.Errorf("couldn't parse %s", path)
		}
		if diff.Shard == request.Shard {
			diffInfo, err := s.readDiff(diff)
			if err != nil {
				return err
			}
			if err := listDiffServer.Send(diffInfo); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *objBlockAPIServer) DeleteDiff(ctx context.Context, request *pfsclient.DeleteDiffRequest) (response *google_protobuf.Empty, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := s.objClient.Delete(s.localServer.diffPath(request.Diff)); err != nil && !s.objClient.IsNotExist(err) {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (s *objBlockAPIServer) readDiff(diff *pfsclient.Diff) (*pfsclient.DiffInfo, error) {
	reader, err := s.objClient.Reader(s.localServer.diffPath(diff), 0, 0)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	result := &pfsclient.DiffInfo{}
	if err := proto.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}
