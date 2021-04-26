package server

import (
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/net/context"
)

var _ APIServer = &sidecarAPIServer{}

type sidecarAPIServer struct {
	APIServer
	log.Logger
	env serviceenv.ServiceEnv
}

func newSidecarAPIServer(embeddedServer APIServer, env serviceenv.ServiceEnv) *sidecarAPIServer {
	return &sidecarAPIServer{
		APIServer: embeddedServer,
		Logger:    log.NewLogger("pfs.SidecarAPI"),
		env:       env,
	}
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *sidecarAPIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx, true)
	commitInfo, err := pachClient.PfsAPIClient.InspectCommit(ctx, request)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

// FlushCommit implements the protobuf pfs.FlushCommit RPC
func (a *sidecarAPIServer) FlushCommit(request *pfs.FlushCommitRequest, server pfs.API_FlushCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	ctx := server.Context()
	pachClient := a.env.GetPachClient(ctx, true)
	client, err := pachClient.PfsAPIClient.FlushCommit(ctx, request)
	if err != nil {
		return err
	}
	for {
		ci, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := server.Send(ci); err != nil {
			return err
		}
	}
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *sidecarAPIServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	ctx := server.Context()
	pachClient := a.env.GetPachClient(ctx, true)
	client, err := pachClient.PfsAPIClient.SubscribeCommit(ctx, request)
	if err != nil {
		return err
	}
	for {
		ci, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := server.Send(ci); err != nil {
			return err
		}
	}
}
