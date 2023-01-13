package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/sirupsen/logrus"
)

// Env is the set of dependencies required by an APIServer
type Env struct {
	ClusterID string
	Config    *serviceenv.Configuration
	Logger    *logrus.Logger
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv) Env {
	return Env{
		ClusterID: senv.ClusterID(),
		Config:    senv.Config(),
		Logger:    senv.Logger(),
	}
}

// APIServer represents an APIServer
type APIServer interface {
	admin.APIServer
}

// NewAPIServer returns a new admin.APIServer
func NewAPIServer(env Env) APIServer {
	return &apiServer{
		clusterInfo: &admin.ClusterInfo{
			ID:                env.ClusterID,
			DeploymentID:      env.Config.DeploymentID,
			VersionWarningsOk: true,
		},
	}
}

type apiServer struct {
	clusterInfo *admin.ClusterInfo
}

const (
	msgNoVersionReq    = "WARNING: The client used to connect to Pachyderm did not send its version, which means that it is likely too old.  Please upgrade it."
	msgClientTooOld    = "WARNING: The client used to connect to Pachyderm is much older than the server; please upgrade the client."
	msgServerTooOld    = "WARNING: The client used to connect to Pachyderm is much newer than the server; please use a version of the client that matches the server."
	fmtServerIsPreview = "WARNING: The client used to connect to Pachyderm is not the same version as the server; only %s is compatible because the server is running a pre-release version."
	fmtClientIsPreview = "WARNING: The client used to connect to Pachyderm is a pre-release version not compatible with the server; please use a released version compatible with %s."
)

func (a *apiServer) InspectCluster(ctx context.Context, request *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
	response := &admin.ClusterInfo{}
	proto.Merge(response, a.clusterInfo)

	serverVersion := version.Version
	if serverVersion == nil {
		// This is very much a "can't happen".
		logrus.Error("internal error: no version information set in version.Version; rebuild Pachyderm")
		return response, nil
	}

	clientVersion := request.GetClientVersion()
	if clientVersion == nil {
		// In 2.4.x, this doesn't log anything on the server side.  In 2.5.x, this logs a
		// message that the client didn't include any version info.
		response.VersionWarnings = append(response.VersionWarnings, msgNoVersionReq)
		return response, nil
	}

	if err := versionpb.IsCompatible(clientVersion, serverVersion); err != nil {
		logrus.Infof("version skew: client is using an incompatible version: %v, client=%v, server=%v", err, clientVersion.Canonical(), serverVersion.Canonical())
		if errors.Is(err, versionpb.ErrClientTooOld) {
			response.VersionWarnings = append(response.VersionWarnings, msgClientTooOld)
		}
		if errors.Is(err, versionpb.ErrServerTooOld) {
			response.VersionWarnings = append(response.VersionWarnings, msgServerTooOld)
		}
		if errors.Is(err, versionpb.ErrIncompatiblePreview) {
			if serverVersion.Additional != "" {
				response.VersionWarnings = append(response.VersionWarnings, fmt.Sprintf(fmtServerIsPreview, serverVersion.Canonical()))
			} else if clientVersion.Additional != "" {
				response.VersionWarnings = append(response.VersionWarnings, fmt.Sprintf(fmtClientIsPreview, serverVersion.Canonical()))
			}
		}
	}
	return response, nil
}
