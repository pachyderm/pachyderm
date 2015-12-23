package server

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

type localAPIServer struct {
	dir string
}

func newLocalAPIServer(dir string) *localAPIServer {
	return &localAPIServer{dir: dir}
}

func (s *localAPIServer) PutBlock(drive.API_PutBlockServer) error {
	return nil
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
