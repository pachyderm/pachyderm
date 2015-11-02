package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.pedge.io/dockervolume"
	"go.pedge.io/env"
	"go.pedge.io/pkg/map"

	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
)

const (
	defaultShard   = 0
	defaultModulus = 1
)

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS":        "0.0.0.0:650",
		"BASE_MOUNTPOINT":    "/tmp/pfs-volume-driver",
		"GRPC_PORT":          "2150",
		"HTTP_PORT":          "1950",
		"VOLUME_DRIVER_NAME": "pfs",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	BaseMountpoint     string `env:"BASE_MOUNTPOINT"`
	GRPCPort           int    `env:"GRPC_PORT"`
	HTTPPort           int    `env:"HTTP_PORT"`
	VolumeDriverName   string `env:"VOLUME_DRIVER_NAME"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	return newServer(appEnvObj.(*appEnv)).Serve()
}

func newServer(appEnv *appEnv) dockervolume.Server {
	return dockervolume.NewTCPServer(
		newVolumeDriver(
			getPFSAddress(appEnv),
			fuse.NewMounterProvider(getPFSAddress(appEnv)),
			appEnv.BaseMountpoint,
		),
		appEnv.VolumeDriverName,
		fmt.Sprintf(":%d", appEnv.HTTPPort),
		dockervolume.ServerOptions{
			GRPCPort: uint16(appEnv.GRPCPort),
		},
	)
}

func getPFSAddress(appEnv *appEnv) string {
	address := appEnv.PachydermPfsd1Port
	if address == "" {
		return appEnv.PfsAddress
	}
	return strings.Replace(address, "tcp://", "", -1)
}

type volumeDriver struct {
	pfsAddress      string
	mounterProvider fuse.MounterProvider
	baseMountpoint  string
}

func newVolumeDriver(
	pfsAddress string,
	mounterProvider fuse.MounterProvider,
	baseMountpoint string,
) *volumeDriver {
	return &volumeDriver{
		pfsAddress,
		mounterProvider,
		baseMountpoint,
	}
}

func (v *volumeDriver) Create(_ string, _ pkgmap.StringStringMap) error {
	return nil
}

func (v *volumeDriver) Remove(_ string, _ pkgmap.StringStringMap, _ string) error {
	return nil
}

func (v *volumeDriver) Mount(name string, opts pkgmap.StringStringMap) (string, error) {
	mount, err := getMount(opts, v.baseMountpoint)
	if err != nil {
		return "", err
	}
	if err := mount.init(); err != nil {
		return "", err
	}
	mounter, err := v.mounterProvider.Get()
	if err != nil {
		return "", err
	}
	if err := mounter.Mount(
		mount.mountpoint,
		mount.shard,
		mount.modulus,
	); err != nil {
		return "", err
	}
	return mount.mountpoint, nil
}

func (v *volumeDriver) Unmount(_ string, _ pkgmap.StringStringMap, mountpoint string) error {
	mounter, err := v.mounterProvider.Get()
	if err != nil {
		return err
	}
	return mounter.Unmount(mountpoint)
}

type mount struct {
	repository string
	commitID   string
	shard      uint64
	modulus    uint64
	mountpoint string
}

func getMount(opts pkgmap.StringStringMap, baseMountpoint string) (*mount, error) {
	repository, err := opts.GetRequiredString("repository")
	if err != nil {
		return nil, err
	}
	commitID, err := opts.GetRequiredString("commit_id")
	if err != nil {
		return nil, err
	}
	shard, err := opts.GetUint64("shard")
	if err != nil {
		return nil, err
	}
	if shard == 0 {
		shard = defaultShard
	}
	modulus, err := opts.GetUint64("modulus")
	if err != nil {
		return nil, err
	}
	if modulus == 0 {
		modulus = defaultModulus
	}
	return &mount{
		repository,
		commitID,
		shard,
		modulus,
		filepath.Join(baseMountpoint, fmt.Sprintf("%s-%s-%d-%d-%s", repository, commitID, shard, modulus, uuid.New())),
	}, nil
}

func (m *mount) init() error {
	return os.MkdirAll(m.mountpoint, 0777)
}
