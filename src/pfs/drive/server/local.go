package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

type localAPIServer struct {
	protorpclog.Logger
	dir string
}

func newLocalAPIServer(dir string) (*localAPIServer, error) {
	server := &localAPIServer{
		Logger: protorpclog.NewLogger("pachyderm.pfs.drive.localAPIServer"),
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

func (s *localAPIServer) PutBlock(putBlockServer drive.API_PutBlockServer) (retErr error) {
	var result *drive.Block
	defer func(start time.Time) { s.Log(nil, result, retErr, time.Since(start)) }(time.Now())
	hash := newHash()
	tmp, err := ioutil.TempFile(s.tmpDir(), "block")
	if err != nil {
		return err
	}
	defer func() {
		if err := tmp.Close(); err != nil && retErr == nil {
			retErr = err
			return
		}
		if result == nil {
			return
		}
		// Check if it's a new block, if so rename it, otherwise remove.
		if _, err := os.Stat(s.blockPath(result)); !os.IsNotExist(err) {
			if err := os.Remove(tmp.Name()); err != nil && retErr == nil {
				retErr = err
				return
			}
			return
		}
		if err := os.Rename(tmp.Name(), s.blockPath(result)); err != nil && retErr == nil {
			retErr = err
			return
		}
	}()
	r := io.TeeReader(protostream.NewStreamingBytesReader(putBlockServer), hash)
	if _, err := io.Copy(tmp, r); err != nil {
		return err
	}
	result = getBlock(hash)
	return putBlockServer.SendAndClose(result)
}

func (s *localAPIServer) GetBlock(request *drive.GetBlockRequest, getBlockServer drive.API_GetBlockServer) (retErr error) {
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	file, err := os.Open(s.blockPath(request.Block))
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := file.Seek(int64(request.OffsetBytes), 0); err != nil {
		return err
	}
	return protostream.WriteToStreamingBytesServer(file, getBlockServer)
}

func (s *localAPIServer) InspectBlock(ctx context.Context, request *drive.InspectBlockRequest) (response *drive.BlockInfo, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	stat, err := os.Stat(s.blockPath(request.Block))
	if err != nil {
		return nil, err
	}
	return &drive.BlockInfo{
		Block: request.Block,
		Created: prototime.TimeToTimestamp(
			stat.ModTime(),
		),
		SizeBytes: uint64(stat.Size()),
	}, nil
}

func (s *localAPIServer) ListBlock(ctx context.Context, request *drive.ListBlockRequest) (response *drive.BlockInfos, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return nil, fmt.Errorf("not implemented")
}

func (s *localAPIServer) CreateDiff(ctx context.Context, request *drive.CreateDiffRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	data, err := proto.Marshal(&drive.DiffInfo{
		Diff:         request.Diff,
		ParentCommit: request.ParentCommit,
		Appends:      request.Appends,
		LastRefs:     request.LastRefs,
		NewPaths:     request.NewPaths,
	})
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.diffPath(request.Diff), data, 0666); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (s *localAPIServer) InspectDiff(ctx context.Context, request *drive.InspectDiffRequest) (response *drive.DiffInfo, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return s.readDiff(request.Diff)
}

func (s *localAPIServer) ListDiff(request *drive.ListDiffRequest, listDiffServer drive.API_ListDiffServer) (retErr error) {
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

func (s *localAPIServer) DeleteDiff(ctx context.Context, request *drive.DeleteDiffRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { s.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return google_protobuf.EmptyInstance, os.Remove(s.diffPath(request.Diff))
}

func (s *localAPIServer) tmpDir() string {
	return filepath.Join(s.dir, "tmp")
}

func (s *localAPIServer) blockDir() string {
	return filepath.Join(s.dir, "block")
}

func (s *localAPIServer) blockPath(block *drive.Block) string {
	return filepath.Join(s.blockDir(), block.Hash)
}

func (s *localAPIServer) diffDir() string {
	return filepath.Join(s.dir, "diff")
}

func (s *localAPIServer) diffPath(diff *drive.Diff) string {
	return filepath.Join(s.diffDir(), diff.Commit.Repo.Name, diff.Commit.Id, strconv.FormatUint(diff.Shard, 10))
}

// pathToDiff parses a path as a diff, it returns nil when parse fails
func (s *localAPIServer) pathToDiff(path string) *drive.Diff {
	repoCommitShard := strings.Split(strings.TrimPrefix(path, s.diffDir()), "/")
	if len(repoCommitShard) < 3 {
		return nil
	}
	shard, err := strconv.ParseUint(repoCommitShard[2], 10, 64)
	if err != nil {
		return nil
	}
	return &drive.Diff{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{Name: repoCommitShard[0]},
			Id:   repoCommitShard[1],
		},
		Shard: shard,
	}
}

func (s *localAPIServer) readDiff(diff *drive.Diff) (*drive.DiffInfo, error) {
	data, err := ioutil.ReadFile(s.diffPath(diff))
	if err != nil {
		return nil, err
	}
	result := &drive.DiffInfo{}
	if err := proto.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}
