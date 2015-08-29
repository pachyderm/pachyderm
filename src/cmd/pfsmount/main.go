package main

import (
	"errors"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"google.golang.org/grpc"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	Address            string `env:"PFS_ADDRESS"`
	Repository         string `env:"PFS_REPOSITORY,required"`
	CommitID           string `env:"PFS_COMMIT_ID"`
	Mountpoint         string `env:"PFS_MOUNTPOINT,required"`
	Shard              uint64 `env:"PFS_SHARD,required"`
	Modulus            uint64 `env:"PFS_MODULUS,required"`
}

func main() {
	mainutil.Main(do, &appEnv{}, nil)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := appEnv.Address
	if address == "" {
		if appEnv.PachydermPfsd1Port == "" {
			return errors.New("PFS_ADDRESS and PACHYDERM_PFSD_1_PORT both not set")
		}
		address = strings.Replace(appEnv.PachydermPfsd1Port, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	return fuse.NewMounter().Mount(
		pfs.NewApiClient(clientConn),
		appEnv.Repository,
		appEnv.CommitID,
		appEnv.Mountpoint,
		appEnv.Shard,
		appEnv.Modulus,
	)
}
