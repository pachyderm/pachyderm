package server

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

// Env is the set of dependencies required by an APIServer
type Env struct {
	ClusterID string
	Config    *pachconfig.Configuration
	PFSServer pfs.APIServer
	Paused    bool
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, paused bool) Env {
	return Env{
		ClusterID: senv.ClusterID(),
		Config:    senv.Config(),
		PFSServer: senv.PfsServer(),
		Paused:    paused,
	}
}

// APIServer represents an APIServer
type APIServer interface {
	admin.APIServer
}

// NewAPIServer returns a new admin.APIServer
func NewAPIServer(env Env) APIServer {
	var host string
	var tls bool
	if pachd := env.Config.PachdSpecificConfiguration; pachd != nil {
		host = pachd.ProxyHost
		tls = pachd.ProxyTLS
	}
	return &apiServer{
		clusterInfo: &admin.ClusterInfo{
			Id:           env.ClusterID,
			DeploymentId: env.Config.DeploymentID,
			WarningsOk:   true,
			ProxyHost:    host,
			ProxyTls:     tls,
			Paused:       env.Paused,
		},
		pfsServer: env.PFSServer,
	}
}

type apiServer struct {
	admin.UnimplementedAPIServer
	clusterInfo *admin.ClusterInfo
	pfsServer   pfs.APIServer
}

const (
	msgNoVersionReq        = "WARNING: The client used to connect to Pachyderm did not send its version, which means that it is likely too old.  Please upgrade it."
	msgClientTooOld        = "WARNING: The client used to connect to Pachyderm is much older than the server; please upgrade the client."
	msgServerTooOld        = "WARNING: The client used to connect to Pachyderm is much newer than the server; please use a version of the client that matches the server."
	fmtServerIsPreview     = "WARNING: The client used to connect to Pachyderm is not the same version as the server; only %s is compatible because the server is running a pre-release version."
	fmtClientIsPreview     = "WARNING: The client used to connect to Pachyderm is a pre-release version not compatible with the server; please use a released version compatible with %s."
	fmtInspectProjectError = "WARNING: Could not inspect project %q: %v"
)

func (a *apiServer) InspectCluster(ctx context.Context, request *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
	response := &admin.ClusterInfo{}
	proto.Merge(response, a.clusterInfo)

	serverVersion := version.Version
	if serverVersion == nil {
		// This is very much a "can't happen".
		log.Error(ctx, "internal error: no version information set in version.Version; rebuild Pachyderm")
		return response, nil
	}

	clientVersion := request.GetClientVersion()
	if clientVersion == nil {
		log.Info(ctx, "version skew: client called InspectCluster without sending its version; it is probably outdated and needs to be upgraded")
		response.Warnings = append(response.Warnings, msgNoVersionReq)
		return response, nil
	}

	if err := versionpb.IsCompatible(clientVersion, serverVersion); err != nil {
		log.Info(ctx, "version skew: client is using an incompatible version", zap.Error(err), zap.String("clientVersion", clientVersion.Canonical()), zap.String("serverVersion", serverVersion.Canonical()))
		if errors.Is(err, versionpb.ErrClientTooOld) {
			response.Warnings = append(response.Warnings, msgClientTooOld)
		}
		if errors.Is(err, versionpb.ErrServerTooOld) {
			response.Warnings = append(response.Warnings, msgServerTooOld)
		}
		if errors.Is(err, versionpb.ErrIncompatiblePreview) {
			if serverVersion.Additional != "" {
				response.Warnings = append(response.Warnings, fmt.Sprintf(fmtServerIsPreview, serverVersion.Canonical()))
			} else if clientVersion.Additional != "" {
				response.Warnings = append(response.Warnings, fmt.Sprintf(fmtClientIsPreview, serverVersion.Canonical()))
			}
		}
	}

	if n := request.GetCurrentProject().GetName(); n != "" {
		if a.pfsServer == nil {
			response.Warnings = append(response.Warnings, fmt.Sprintf("PFS server not running; cannot check existence of project %s", request.GetCurrentProject()))
		} else if _, err := a.pfsServer.InspectProject(ctx, &pfs.InspectProjectRequest{Project: request.GetCurrentProject()}); err != nil {
			response.Warnings = append(response.Warnings, fmt.Sprintf(fmtInspectProjectError, request.GetCurrentProject(), err))
		}
	}
	return response, nil
}
