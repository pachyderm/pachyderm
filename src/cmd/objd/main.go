package main

import (
	"io/ioutil"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/obj"
	"go.pedge.io/env"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
)

type appEnv struct {
	StorageRoot string `env:"OBJ_ROOT,required"`
	Port        uint16 `env:"OBJ_PORT,default=652"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	var blockAPIServer pfs.BlockAPIServer
	if err := func() error {
		bucket, err := ioutil.ReadFile("/amazon-secret/bucket")
		if err != nil {
			return err
		}
		id, err := ioutil.ReadFile("/amazon-secret/id")
		if err != nil {
			return err
		}
		secret, err := ioutil.ReadFile("/amazon-secret/secret")
		if err != nil {
			return err
		}
		token, err := ioutil.ReadFile("/amazon-secret/token")
		if err != nil {
			return err
		}
		region, err := ioutil.ReadFile("/amazon-secret/region")
		if err != nil {
			return err
		}
		objClient, err := obj.NewAmazonClient(string(bucket), string(id), string(secret), string(token), string(region))
		if err != nil {
			return err
		}
		blockAPIServer, err = server.NewObjBlockAPIServer(appEnv.StorageRoot, objClient)
		if err != nil {
			return err
		}
		return nil
	}(); err != nil {
		protolion.Errorf("failed to create obj backend, falling back to local")
		blockAPIServer, err = server.NewLocalBlockAPIServer(appEnv.StorageRoot)
		if err != nil {
			return err
		}
	}

	return protoserver.Serve(
		func(s *grpc.Server) {
			pfs.RegisterBlockAPIServer(s, blockAPIServer)
		},
		protoserver.ServeOptions{
			Version: pachyderm.Version,
		},
		protoserver.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
}
