package main

import (
	"os"
	"path/filepath"

	"go.pedge.io/pkg/map"
)

type volumeDriver struct {
	baseDirPath string
}

func newVolumeDriver(baseDirPath string) *volumeDriver {
	return &volumeDriver{baseDirPath}
}

func (v *volumeDriver) Create(_ string, _ pkgmap.StringStringMap) error {
	return nil
}

func (v *volumeDriver) Remove(_ string, _ pkgmap.StringStringMap, _ string) error {
	return nil
}

func (v *volumeDriver) Mount(name string, _ pkgmap.StringStringMap) (string, error) {
	dirPath := filepath.Join(v.baseDirPath, name)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", err
	}
	return dirPath, nil
}

func (v *volumeDriver) Unmount(_ string, _ pkgmap.StringStringMap, mountpoint string) error {
	return os.RemoveAll(mountpoint)
}
