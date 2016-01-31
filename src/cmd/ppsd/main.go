package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	persistserver "github.com/pachyderm/pachyderm/src/pps/persist/server"
	"github.com/pachyderm/pachyderm/src/pps/pipelineserver"
	"go.pedge.io/env"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type appEnv struct {
	Port            uint16 `env:"PPS_PORT,default=651"`
	DatabaseAddress string `env:"RETHINK_PORT_28015_TCP_ADDR,required"`
	DatabaseName    string `env:"PPS_DATABASE_NAME,default=pachyderm"`
	PfsdAddress     string `env:"PFSD_PORT_650_TCP_ADDR,required"`
	KubeAddress     string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rethinkAPIServer, err := getRethinkAPIServer(appEnv)
	if err != nil {
		return err
	}
	pfsdAddress := getPfsdAddress(appEnv)
	clientConn, err := grpc.Dial(pfsdAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	pfsAPIClient := pfs.NewAPIClient(clientConn)
	kubeClient, err := getKubeClient(appEnv)
	if err != nil {
		return err
	}
	jobAPIServer := jobserver.NewAPIServer(
		pfsAPIClient,
		rethinkAPIServer,
		kubeClient,
	)
	jobAPIClient := pps.NewLocalJobAPIClient(jobAPIServer)
	pipelineAPIServer := pipelineserver.NewAPIServer(pfsAPIClient, jobAPIClient, rethinkAPIServer)
	if err := pipelineAPIServer.Start(); err != nil {
		return err
	}
	return protoserver.Serve(
		func(s *grpc.Server) {
			pps.RegisterJobAPIServer(s, jobAPIServer)
			pps.RegisterInternalJobAPIServer(s, jobAPIServer)
			pps.RegisterPipelineAPIServer(s, pipelineAPIServer)
		},
		protoserver.ServeOptions{
			Version: pachyderm.Version,
		},
		protoserver.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
}

func getRethinkAPIServer(env *appEnv) (persist.APIServer, error) {
	if err := persistserver.InitDBs(getRethinkAddress(env), env.DatabaseName); err != nil {
		return nil, err
	}
	return persistserver.NewRethinkAPIServer(getRethinkAddress(env), env.DatabaseName)
}

func getRethinkAddress(env *appEnv) string {
	return fmt.Sprintf("%s:28015", env.DatabaseAddress)
}

func getPfsdAddress(env *appEnv) string {
	return fmt.Sprintf("%s:650", env.PfsdAddress)
}

func getKubeAddress(env *appEnv) string {
	return fmt.Sprintf("%s:443", env.KubeAddress)
}

func getKubeClient(env *appEnv) (*kube.Client, error) {
	kubeClient, err := kube.NewInCluster()
	if err != nil {
		protolion.Errorf("Falling back to insecure kube client due to error from NewInCluster: %s", err.Error())
	} else {
		return kubeClient, err
	}
	config := &kube.Config{
		Host:     getKubeAddress(env),
		Insecure: true,
	}
	return kube.New(config)
}
