package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

type localAPIServer struct {
	dir string
}

func newLocalAPIServer(dir string) *localAPIServer {
	return &localAPIServer{dir: dir}
}

func (s *localAPIServer) PutBlock(putBlockServer drive.API_PutBlockServer) (retErr error) {
	var result *drive.Block
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
	file, err := os.Open(s.blockPath(request.Block))
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protostream.WriteToStreamingBytesServer(file, getBlockServer)
}

func (s *localAPIServer) InspectBlock(ctx context.Context, request *drive.InspectBlockRequest) (*drive.BlockInfo, error) {
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
	return nil, nil
}

func (s *localAPIServer) ListBlock(context.Context, *drive.ListBlockRequest) (*drive.BlockInfos, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *localAPIServer) CreateDiff(ctx context.Context, request *drive.CreateDiffRequest) (_ *google_protobuf.Empty, retErr error) {
	data, err := proto.Marshal(&drive.DiffInfo{
		Diff:         request.Diff,
		ParentCommit: request.ParentCommit,
		Appends:      request.Appends,
		LastRef:      request.LastRef,
	})
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.diffPath(request.Diff), data, 0666); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *localAPIServer) InspectDiff(ctx context.Context, request *drive.InspectDiffRequest) (*drive.DiffInfo, error) {
	data, err := ioutil.ReadFile(s.diffPath(request.Diff))
	if err != nil {
		return nil, err
	}
	result := &drive.DiffInfo{}
	if err := proto.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *localAPIServer) ListDiff(context.Context, *drive.ListDiffRequest) (*drive.DiffInfos, error) {
	return nil, nil
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
