package server

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/pps"
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
		Logger:    log.NewLogger("pps.SidecarAPI"),
		env:       env,
	}
}

// InspectJob implements the protobuf pps.InspectJob RPC
func (a *sidecarAPIServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx, true)
	jobInfo, err := pachClient.PpsAPIClient.InspectJob(ctx, request)
	if err != nil {
		return nil, err
	}
	return jobInfo, nil
}
