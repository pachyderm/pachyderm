package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
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
		Logger: protorpclog.NewLogger("pachyderm.pfsclient.localBlockAPIServer"),
		dir:    dir,
	}
	if err := os.MkdirAll(server.tmpDir(), 0777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(server.diffDir(), 0777); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(server.blockDir(), 0777); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *localBlockAPIServer) PutBlock(putBlockServer pfsclient.BlockAPI_PutBlockServer) (retErr error) {
	result := &pfsclient.BlockRefs{}
	defer func(start time.Time) { s.Log(nil, result, retErr, time.Since(start)) }(time.Now())
	reader := bufio.NewReader(protostream.NewStreamingBytesReader(putBlockServer))
	for {
		blockRef, err := s.putOneBlock(reader)
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
	reader := io.NewSectionReader(file, int64(request.OffsetBytes), int64(request.SizeBytes))
	return protostream.WriteToStreamingBytesServer(reader, getBlockServer)
}

func (s *localBlockAPIServer) DeleteBlock(ctx context.Context, request *pfsclient.DeleteBlockRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return google_protobuf.EmptyInstance, s.deleteBlock(request.Block)
}

func (s *localBlockAPIServer) InspectBlock(ctx context.Context, request *pfsclient.InspectBlockRequest) (response *pfsclient.BlockInfo, retErr error) {
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
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return nil, fmt.Errorf("not implemented")
}

func (s *localBlockAPIServer) CreateDiff(ctx context.Context, request *pfsclient.DiffInfo) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path.Dir(s.diffPath(request.Diff)), 0777); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.diffPath(request.Diff), data, 0666); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (s *localBlockAPIServer) InspectDiff(ctx context.Context, request *pfsclient.InspectDiffRequest) (response *pfsclient.DiffInfo, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return s.readDiff(request.Diff)
}

func (s *localBlockAPIServer) ListDiff(request *pfsclient.ListDiffRequest, listDiffServer pfsclient.BlockAPI_ListDiffServer) (retErr error) {
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	if err := filepath.Walk(s.diffDir(), func(path string, info os.FileInfo, err error) error {
		diff := s.pathToDiff(path)
		if diff == nil {
			// likely a directory
			return nil
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

func (s *localBlockAPIServer) DeleteDiff(ctx context.Context, request *pfsclient.DeleteDiffRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return google_protobuf.EmptyInstance, os.Remove(s.diffPath(request.Diff))
}

func (s *localBlockAPIServer) tmpDir() string {
	return filepath.Join(s.dir, "tmp")
}

func (s *localBlockAPIServer) blockDir() string {
	return filepath.Join(s.dir, "block")
}

func (s *localBlockAPIServer) blockPath(block *pfsclient.Block) string {
	return filepath.Join(s.blockDir(), block.Hash)
}

func (s *localBlockAPIServer) diffDir() string {
	return filepath.Join(s.dir, "diff")
}

func (s *localBlockAPIServer) diffPath(diff *pfsclient.Diff) string {
	commitID := diff.Commit.ID
	if commitID == "" {
		// each repo creates a diff per shard with an empty commit
		// so it works as a path we make that an underscore
		commitID = "_"
	}
	return filepath.Join(s.diffDir(), diff.Commit.Repo.Name, strconv.FormatUint(diff.Shard, 10), commitID)
}

// pathToDiff parses a path as a diff, it returns nil when parse fails
func (s *localBlockAPIServer) pathToDiff(path string) *pfsclient.Diff {
	repoCommitShard := strings.Split(strings.TrimPrefix(path, s.diffDir()+"/"), "/")
	if len(repoCommitShard) < 3 {
		return nil
	}
	commitID := repoCommitShard[2]
	if commitID == "_" {
		commitID = ""
	}
	shard, err := strconv.ParseUint(repoCommitShard[1], 10, 64)
	if err != nil {
		return nil
	}
	return &pfsclient.Diff{
		Commit: &pfsclient.Commit{
			Repo: &pfsclient.Repo{Name: repoCommitShard[0]},
			ID:   commitID,
		},
		Shard: shard,
	}
}

func (s *localBlockAPIServer) readDiff(diff *pfsclient.Diff) (*pfsclient.DiffInfo, error) {
	data, err := ioutil.ReadFile(s.diffPath(diff))
	if err != nil {
		return nil, err
	}
	result := &pfsclient.DiffInfo{}
	if err := proto.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func readBlock(reader *bufio.Reader) (*pfsclient.BlockRef, []byte, error) {
	var buffer bytes.Buffer
	var bytesWritten int
	hash := newHash()
	for {
		// We don't use ReadBytes or ReadString because they allocate new buffers
		// whereas ReadLine only uses an internal buffer
		bytes, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		// Append a newline if we have reached the end of the line
		if !isPrefix {
			bytes = append(bytes, '\n')
		}
		buffer.Write(bytes)
		hash.Write(bytes)
		bytesWritten += len(bytes)
		// len(bytes) == 0 when we've
		if bytesWritten > blockSize {
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

func (s *localBlockAPIServer) putOneBlock(reader *bufio.Reader) (*pfsclient.BlockRef, error) {
	blockRef, data, err := readBlock(reader)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(s.blockPath(blockRef.Block)); os.IsNotExist(err) {
		ioutil.WriteFile(s.blockPath(blockRef.Block), data, 0666)
	}
	return blockRef, nil
}

func (s *localBlockAPIServer) deleteBlock(block *pfsclient.Block) error {
	return os.Remove(s.blockPath(block))
}
