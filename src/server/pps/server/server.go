package server

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

// connectAndInitSpecRepo initialize the GRPC connection to pachd, and then
// uses that connection to create the 'spec' repo that PPS needs to exist on
// startup
func (a *apiServer) connectAndInitSpecRepo() {
	pachClient := a.env.GetPachClient(context.Background()) // no creds on startup
	// Initialize spec repo
	if err := a.sudo(pachClient, func(superUserClient *client.APIClient) error {
		if err := superUserClient.CreateRepo(ppsconsts.SpecRepo); err != nil && !isAlreadyExistsErr(err) {
			return err
		}
		return nil
	}); err != nil {
		panic(fmt.Sprintf("could not create pipeline spec repo: %v", err))
	}
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env *serviceenv.ServiceEnv,
	etcdPrefix string,
	namespace string,
	workerImage string,
	workerSidecarImage string,
	workerImagePullPolicy string,
	storageRoot string,
	storageBackend string,
	storageHostPath string,
	iamRole string,
	imagePullSecret string,
	reporter *metrics.Reporter,
) (ppsclient.APIServer, error) {
	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API"),
		env:                   env,
		etcdPrefix:            etcdPrefix,
		namespace:             namespace,
		workerImage:           workerImage,
		workerSidecarImage:    workerSidecarImage,
		workerImagePullPolicy: workerImagePullPolicy,
		storageRoot:           storageRoot,
		storageBackend:        storageBackend,
		storageHostPath:       storageHostPath,
		iamRole:               iamRole,
		imagePullSecret:       imagePullSecret,
		reporter:              reporter,
		pipelines:             ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix),
		jobs:                  ppsdb.Jobs(env.GetEtcdClient(), etcdPrefix),
		monitorCancels:        make(map[string]func()),
	}
	apiServer.validateKube()
	go apiServer.connectAndInitSpecRepo()
	go apiServer.master()
	return apiServer, nil
}

// NewSidecarAPIServer creates an APIServer that has limited functionalities
// and is meant to be run as a worker sidecar.  It cannot, for instance,
// create pipelines.
func NewSidecarAPIServer(
	env *serviceenv.ServiceEnv,
	etcdPrefix string,
	iamRole string,
	reporter *metrics.Reporter,
) (ppsclient.APIServer, error) {
	apiServer := &apiServer{
		Logger:     log.NewLogger("pps.API"),
		env:        env,
		etcdPrefix: etcdPrefix,
		iamRole:    iamRole,
		reporter:   reporter,
		pipelines:  ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix),
		jobs:       ppsdb.Jobs(env.GetEtcdClient(), etcdPrefix),
	}
	go apiServer.connectAndInitSpecRepo()
	return apiServer, nil
}
