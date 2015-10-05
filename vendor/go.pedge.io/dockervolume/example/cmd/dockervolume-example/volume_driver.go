package main

import (
	"os"
	"path/filepath"

	"go.pedge.io/dockervolume"
)

type volumeDriver struct {
	baseDirPath string
}

func newVolumeDriver(baseDirPath string) *volumeDriver {
	return &volumeDriver{baseDirPath}
}

func (v *volumeDriver) Create(_ string, _ dockervolume.Opts) error {
	return nil
}

func (v *volumeDriver) Remove(_ string, _ dockervolume.Opts, _ string) error {
	return nil
}

func (v *volumeDriver) Mount(name string, _ dockervolume.Opts) (string, error) {
	dirPath := filepath.Join(v.baseDirPath, name)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", err
	}
	return dirPath, nil
}

func (v *volumeDriver) Unmount(_ string, _ dockervolume.Opts, mountpoint string) error {
	return os.RemoveAll(mountpoint)
}
