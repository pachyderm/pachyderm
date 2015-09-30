package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"go.pedge.io/dockervolume"
	"go.pedge.io/env"
	"go.pedge.io/protolog/logrus"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

const (
	volumeDriverName = "pfs"

	defaultShard   = 0
	defaultModulus = 1
)

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS":     "0.0.0.0:650",
		"BASE_MOUNTPOINT": "/tmp/pfs-volume-driver",
		"GRPC_PORT":       "2150",
		"HTTP_PORT":       "1950",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS"`
	BaseMountpoint     string `env:"BASE_MOUNTPOINT"`
	GRPCPort           int    `env:"GRPC_PORT"`
	HTTPPort           int    `env:"HTTP_PORT"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.PfsAddress
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	return dockervolume.NewTCPServer(
		newVolumeDriver(
			func() (fuse.Mounter, error) {
				clientConn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					return nil, err
				}
				return fuse.NewMounter(
					pfs.NewApiClient(
						clientConn,
					),
				), nil
			},
			appEnv.BaseMountpoint,
		),
		volumeDriverName,
		fmt.Sprintf(":%d", appEnv.HTTPPort),
		dockervolume.ServerOptions{
			GRPCPort: uint16(appEnv.GRPCPort),
		},
	).Serve()
}

type volumeDriver struct {
	mounterProvider func() (fuse.Mounter, error)
	baseMountpoint  string

	mounterOnce  *sync.Once
	mounterValue *atomic.Value
	mounterErr   *atomic.Value
}

func newVolumeDriver(
	mounterProvider func() (fuse.Mounter, error),
	baseMountpoint string,
) *volumeDriver {
	return &volumeDriver{
		mounterProvider,
		baseMountpoint,
		&sync.Once{},
		&atomic.Value{},
		&atomic.Value{},
	}
}

func (v *volumeDriver) Create(_ string, _ map[string]string) error {
	return nil
}

func (v *volumeDriver) Remove(_ string, _ map[string]string, _ string) error {
	return nil
}

func (v *volumeDriver) Mount(name string, opts map[string]string) (string, error) {
	repository, err := getRequiredString(opts, "repository")
	if err != nil {
		return "", err
	}
	commitID, err := getRequiredString(opts, "commit_id")
	if err != nil {
		return "", err
	}
	shard, err := getOptionalUint64(opts, "shard", defaultShard)
	if err != nil {
		return "", err
	}
	modulus, err := getOptionalUint64(opts, "modulus", defaultModulus)
	if err != nil {
		return "", err
	}
	mountpoint := filepath.Join(v.baseMountpoint, fmt.Sprintf("name-%s", strings.Replace(uuid.NewV4().String(), "-", "", -1)))
	if err := os.MkdirAll(mountpoint, 0777); err != nil {
		return "", err
	}
	mounter, err := v.getMounter()
	if err != nil {
		return "", err
	}
	if err := mounter.Mount(
		repository,
		commitID,
		mountpoint,
		shard,
		modulus,
	); err != nil {
		return "", err
	}
	return mountpoint, nil
}

func (v *volumeDriver) Unmount(_ string, _ map[string]string, mountpoint string) error {
	mounter, err := v.getMounter()
	if err != nil {
		return err
	}
	if err := mounter.Unmount(mountpoint); err != nil {
		return err
	}
	return mounter.Wait(mountpoint)
}

func (v *volumeDriver) getMounter() (fuse.Mounter, error) {
	v.mounterOnce.Do(func() {
		value, err := v.mounterProvider()
		if value != nil {
			v.mounterValue.Store(value)
		}
		if err != nil {
			v.mounterErr.Store(err)
		}
	})
	mounterObj := v.mounterValue.Load()
	errObj := v.mounterErr.Load()
	if errObj != nil {
		return nil, errObj.(error)
	}
	return mounterObj.(fuse.Mounter), nil
}

func getOptionalString(m map[string]string, key string, defaultValue string) (string, error) {
	value, ok := m[key]
	if !ok {
		return defaultValue, nil
	}
	return value, nil
}

func getRequiredString(m map[string]string, key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", fmt.Errorf("pfs-volume-driver: must pass opt %s (--opt %s=VALUE)", key, key)
	}
	return value, nil
}

func getOptionalUint64(m map[string]string, key string, defaultValue uint64) (uint64, error) {
	valueObj, err := getOptionalString(m, key, strconv.FormatUint(defaultValue, 10))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(valueObj, 10, 64)
}

func getRequiredUint64(m map[string]string, key string) (uint64, error) {
	valueObj, err := getRequiredString(m, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(valueObj, 10, 64)
}
