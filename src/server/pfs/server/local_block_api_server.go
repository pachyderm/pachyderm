package server

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/types"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"

	"golang.org/x/net/context"
)

type localBlockAPIServer struct {
	log.Logger
	dir string
}

func newLocalBlockAPIServer(dir string) (*localBlockAPIServer, error) {
	server := &localBlockAPIServer{
		Logger: log.NewLogger("pfs.BlockAPIServer.Local"),
		dir:    dir,
	}
	if err := os.MkdirAll(server.blockDir(), 0777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(server.objectDir(), 0777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(server.tagDir(), 0777); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *localBlockAPIServer) PutObject(server pfsclient.ObjectAPI_PutObjectServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	defer drainObjectServer(server)
	hash := newHash()
	tmpPath := filepath.Join(s.objectDir(), uuid.NewWithoutDashes())
	putObjectReader := &putObjectReader{
		server: server,
	}
	r := io.TeeReader(putObjectReader, hash)
	if err := func() error {
		w, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		if _, err := io.Copy(w, r); err != nil {
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
	objectPath := s.objectPath(object)
	if err := os.Rename(tmpPath, objectPath); err != nil && retErr == nil {
		retErr = err
	}
	if _, err := s.TagObject(server.Context(), &pfsclient.TagObjectRequest{
		Object: object,
		Tags:   putObjectReader.tags,
	}); err != nil {
		return err
	}
	return nil
}

func (s *localBlockAPIServer) GetObject(request *pfsclient.Object, getObjectServer pfsclient.ObjectAPI_GetObjectServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	file, err := os.Open(s.objectPath(request))
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return grpcutil.WriteToStreamingBytesServer(file, getObjectServer)
}

func (s *localBlockAPIServer) GetObjects(request *pfsclient.GetObjectsRequest, getObjectsServer pfsclient.ObjectAPI_GetObjectsServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	offsetBytes := int64(request.OffsetBytes)
	sizeBytes := int64(request.SizeBytes)
	for _, object := range request.Objects {
		fileInfo, err := os.Stat(s.objectPath(object))
		if err != nil {
			return err
		}
		if fileInfo.Size() < offsetBytes {
			offsetBytes -= fileInfo.Size()
			continue
		}
		if err := func() error {
			file, err := os.Open(s.objectPath(object))
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if request.SizeBytes == 0 {
				sizeBytes = fileInfo.Size() - offsetBytes
			}
			return grpcutil.WriteToStreamingBytesServer(io.NewSectionReader(file, offsetBytes, sizeBytes), getObjectsServer)
		}(); err != nil {
			return err
		}
		if request.SizeBytes != 0 {
			sizeBytes -= (fileInfo.Size() - offsetBytes)
			if sizeBytes <= 0 {
				break
			}
		}
		offsetBytes = 0
	}
	return nil
}

func (s *localBlockAPIServer) TagObject(ctx context.Context, request *pfsclient.TagObjectRequest) (response *types.Empty, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	objectPath := s.objectPath(request.Object)
	for _, tag := range request.Tags {
		if err := os.RemoveAll(s.tagPath(tag)); err != nil {
			return nil, err
		}
		if err := os.Symlink(objectPath, s.tagPath(tag)); err != nil {
			return nil, err
		}
	}
	return &types.Empty{}, nil
}

func (s *localBlockAPIServer) InspectObject(ctx context.Context, request *pfsclient.Object) (response *pfsclient.ObjectInfo, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	fileInfo, err := os.Stat(s.objectPath(request))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object: %s not found", request.Hash)
		}
		return nil, err
	}
	return &pfsclient.ObjectInfo{
		Object: request,
		BlockRef: &pfsclient.BlockRef{
			Range: &pfsclient.ByteRange{
				Upper: uint64(fileInfo.Size()),
			},
		},
	}, nil
}

func (s *localBlockAPIServer) CheckObject(ctx context.Context, request *pfsclient.CheckObjectRequest) (response *pfsclient.CheckObjectResponse, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	_, err := os.Stat(s.objectPath(request.Object))
	if err != nil {
		if os.IsNotExist(err) {
			return &pfsclient.CheckObjectResponse{
				Exists: false,
			}, nil
		}
		return nil, err
	}
	return &pfsclient.CheckObjectResponse{
		Exists: true,
	}, nil
}

func (s *localBlockAPIServer) ListObjects(request *pfsclient.ListObjectsRequest, listObjectsServer pfsclient.ObjectAPI_ListObjectsServer) (retErr error) {
	return errors.New("TODO")
}

func (s *localBlockAPIServer) ListTags(request *pfsclient.ListTagsRequest, server pfsclient.ObjectAPI_ListTagsServer) (retErr error) {
	return errors.New("TODO")
}

func (s *localBlockAPIServer) DeleteTags(ctx context.Context, request *pfsclient.DeleteTagsRequest) (response *pfsclient.DeleteTagsResponse, retErr error) {
	return nil, errors.New("TODO")
}

func (s *localBlockAPIServer) DeleteObjects(ctx context.Context, request *pfsclient.DeleteObjectsRequest) (response *pfsclient.DeleteObjectsResponse, retErr error) {
	return nil, errors.New("TODO")
}

func (s *localBlockAPIServer) GetTag(request *pfsclient.Tag, getTagServer pfsclient.ObjectAPI_GetTagServer) (retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	file, err := os.Open(s.tagPath(request))
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return grpcutil.WriteToStreamingBytesServer(file, getTagServer)
}

func (s *localBlockAPIServer) InspectTag(ctx context.Context, request *pfsclient.Tag) (response *pfsclient.ObjectInfo, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	objectPath, err := os.Readlink(s.tagPath(request))
	if err != nil {
		return nil, err
	}
	return s.InspectObject(ctx, &pfsclient.Object{Hash: filepath.Base(objectPath)})
}

func (s *localBlockAPIServer) Compact(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	return &types.Empty{}, nil
}

func (s *localBlockAPIServer) blockDir() string {
	return filepath.Join(s.dir, "block")
}

func (s *localBlockAPIServer) blockPath(block *pfsclient.Block) string {
	return filepath.Join(s.blockDir(), block.Hash)
}

func (s *localBlockAPIServer) objectDir() string {
	return filepath.Join(s.dir, "object")
}

func (s *localBlockAPIServer) objectPath(object *pfsclient.Object) string {
	return filepath.Join(s.objectDir(), object.Hash)
}

func (s *localBlockAPIServer) tagDir() string {
	return filepath.Join(s.dir, "tag")
}

func (s *localBlockAPIServer) tagPath(tag *pfsclient.Tag) string {
	return filepath.Join(s.tagDir(), tag.Name)
}

func (s *localBlockAPIServer) indexDir() string {
	return filepath.Join(s.dir, "index")
}

func (s *localBlockAPIServer) indexPath(prefix string) string {
	return filepath.Join(s.indexDir(), prefix)
}

type putObjectReader struct {
	server pfsclient.ObjectAPI_PutObjectServer
	buffer bytes.Buffer
	tags   []*pfsclient.Tag
}

func (r *putObjectReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		// buffer.Write cannot error
		r.buffer.Write(request.Value)
		r.tags = append(r.tags, request.Tags...)
	}
	return r.buffer.Read(p)
}

func drainObjectServer(putObjectServer pfsclient.ObjectAPI_PutObjectServer) {
	for {
		if _, err := putObjectServer.Recv(); err != nil {
			break
		}
	}
}
