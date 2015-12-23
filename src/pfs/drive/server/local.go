package server

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
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

func (s *localAPIServer) GetBlock(*drive.GetBlockRequest, drive.API_GetBlockServer) error {
	return nil
}

func (s *localAPIServer) InspectBlock(context.Context, *drive.InspectBlockRequest) (*drive.BlockInfo, error) {
	return nil, nil
}

func (s *localAPIServer) ListBlock(context.Context, *drive.ListBlockRequest) (*drive.BlockInfos, error) {
	return nil, nil
}

func (s *localAPIServer) CreateDiff(context.Context, *drive.CreateDiffRequest) (*google_protobuf.Empty, error) {
	return nil, nil
}

func (s *localAPIServer) InspectDiff(context.Context, *drive.InspectDiffRequest) (*drive.DiffInfo, error) {
	return nil, nil
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
