package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

type localBlockAPIServer struct {
	protorpclog.Logger
	dir string
}

func newLocalBlockAPIServer(dir string) (*localBlockAPIServer, error) {
	server := &localBlockAPIServer{
		Logger: protorpclog.NewLogger("pfs.BlockAPIServer.Local"),
		dir:    dir,
	}
	if err := os.MkdirAll(server.blockDir(), 0777); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *localBlockAPIServer) PutBlock(putBlockServer pfsclient.BlockAPI_PutBlockServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	result := &pfsclient.BlockRefs{}
	defer func(start time.Time) { s.Log(nil, result, retErr, time.Since(start)) }(time.Now())
	defer drainBlockServer(putBlockServer)

	putBlockRequest, err := putBlockServer.Recv()
	if err != nil {
		if err != io.EOF {
			return err
		}
		// Allow empty PutBlock requests, in this case we don't create any actual blockRefs
		return putBlockServer.SendAndClose(result)
	}

	reader := bufio.NewReader(&putBlockReader{
		server: putBlockServer,
		buffer: bytes.NewBuffer(putBlockRequest.Value),
	})

	decoder := json.NewDecoder(reader)

	for {
		blockRef, err := s.putOneBlock(putBlockRequest.Delimiter, reader, decoder)
		if err != nil {
			return err
		}
		result.BlockRef = append(result.BlockRef, blockRef)
		if (blockRef.Range.Upper - blockRef.Range.Lower) < uint64(blockSize) {
			break
		}
	}
	return putBlockServer.SendAndClose(result)
}

func (s *localBlockAPIServer) blockFile(block *pfsclient.Block) (*os.File, error) {
	return os.Open(s.blockPath(block))
}

func (s *localBlockAPIServer) GetBlock(request *pfsclient.GetBlockRequest, getBlockServer pfsclient.BlockAPI_GetBlockServer) (retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	file, err := s.blockFile(request.Block)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var reader io.Reader
	if request.SizeBytes == 0 {
		_, err = file.Seek(int64(request.OffsetBytes), 0)
		if err != nil {
			return err
		}
		reader = file
	} else {
		reader = io.NewSectionReader(file, int64(request.OffsetBytes), int64(request.SizeBytes))
	}
	return protostream.WriteToStreamingBytesServer(reader, getBlockServer)
}

func (s *localBlockAPIServer) DeleteBlock(ctx context.Context, request *pfsclient.DeleteBlockRequest) (response *google_protobuf.Empty, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return google_protobuf.EmptyInstance, s.deleteBlock(request.Block)
}

func (s *localBlockAPIServer) InspectBlock(ctx context.Context, request *pfsclient.InspectBlockRequest) (response *pfsclient.BlockInfo, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	stat, err := os.Stat(s.blockPath(request.Block))
	if err != nil {
		return nil, err
	}
	return &pfsclient.BlockInfo{
		Block: request.Block,
		Created: prototime.TimeToTimestamp(
			stat.ModTime(),
		),
		SizeBytes: uint64(stat.Size()),
	}, nil
}

func (s *localBlockAPIServer) ListBlock(ctx context.Context, request *pfsclient.ListBlockRequest) (response *pfsclient.BlockInfos, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return nil, fmt.Errorf("not implemented")
}

func (s *localBlockAPIServer) PutObject(ctx context.Context, request *pfsclient.PutObjectRequest) (response *pfsclient.Object, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	object := pfsclient.Object{Hash: base64.URLEncoding.EncodeToString(newHash().Sum(request.Value))}
	objectPath := s.objectPath(object)
	if err := ioutil.WriteFile(objectPath, request.Value, 0666); err != nil {
		return nil, err
	}
	for _, tag := range request.Tags {
		if err := os.Symlink(objectPath, s.objectPath(tag)); err != nil {
			return nil, err
		}
	}
	return object, nil
}

func (s *localBlockAPIServer) GetObject(ctx context.Context, request *pfsclient.PutObjectRequest) (response *google_protobuf.BytesValue, retErr error) {
	func() { s.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	value, err := ioutil.ReadFile(objectPath(request))
	if err != nil {
		return nil, err
	}
	return &google_protobuf.BytesValue{Value: value}, nil
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
	return filepath.Join(objectDir(), object.Hash)
}

func (s *localBlockAPIServer) tagDir() string {
	return filepath.Join(s.dir, "tag")
}

func (s *localBlockAPIServer) tagPath(tag string) string {
	return filepath.Join(s.tagDir(), tag)
}

func (s *localBlockAPIServer) indexDir() string {
	return filepath.Join(s.dir, "index")
}

func (s *localBlockAPIServer) index(prefix string) string {
	return filepath.Join(s.indexDir(), prefix)
}

func readBlock(delimiter pfsclient.Delimiter, reader *bufio.Reader, decoder *json.Decoder) (*pfsclient.BlockRef, []byte, error) {
	var buffer bytes.Buffer
	var bytesWritten int
	hash := newHash()
	EOF := false
	var value []byte

	for !EOF {
		var err error
		if delimiter == pfsclient.Delimiter_JSON {
			var jsonValue json.RawMessage
			err = decoder.Decode(&jsonValue)
			value = jsonValue
		} else if delimiter == pfsclient.Delimiter_NONE {
			value = make([]byte, 1000)
			n, e := reader.Read(value)
			err = e
			value = value[:n]
		} else {
			value, err = reader.ReadBytes('\n')
		}
		if err != nil {
			if err == io.EOF {
				EOF = true
			} else {
				return nil, nil, err
			}
		}
		buffer.Write(value)
		hash.Write(value)
		bytesWritten += len(value)
		if bytesWritten > blockSize && delimiter != pfsclient.Delimiter_NONE {
			break
		}
	}

	return &pfsclient.BlockRef{
		Block: getBlock(hash),
		Range: &pfsclient.ByteRange{
			Lower: 0,
			Upper: uint64(buffer.Len()),
		},
	}, buffer.Bytes(), nil
}

func (s *localBlockAPIServer) putOneBlock(delimiter pfsclient.Delimiter, reader *bufio.Reader, decoder *json.Decoder) (*pfsclient.BlockRef, error) {
	blockRef, data, err := readBlock(delimiter, reader, decoder)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(s.blockPath(blockRef.Block)); os.IsNotExist(err) {
		ioutil.WriteFile(s.blockPath(blockRef.Block), data, 0666)
	}
	return blockRef, nil
}

func (s *localBlockAPIServer) deleteBlock(block *pfsclient.Block) error {
	return os.RemoveAll(s.blockPath(block))
}

type putBlockReader struct {
	server pfsclient.BlockAPI_PutBlockServer
	buffer *bytes.Buffer
}

func (r *putBlockReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		// Buffer.Write cannot error
		r.buffer.Write(request.Value)
	}
	return r.buffer.Read(p)
}

func drainBlockServer(putBlockServer pfsclient.BlockAPI_PutBlockServer) {
	for {
		if _, err := putBlockServer.Recv(); err != nil {
			break
		}
	}
}
