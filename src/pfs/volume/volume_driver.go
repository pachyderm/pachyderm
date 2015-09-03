package volume

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/satori/go.uuid"
)

const (
	defaultShard   = 0
	defaultModulus = 1
)

type volume struct {
	repository string
	commitID   string
	shard      uint64
	modulus    uint64
	mountpoint string
}

type volumeDriver struct {
	mounter        fuse.Mounter
	baseMountpoint string

	nameToVolume map[string]*volume
	lock         *sync.RWMutex
}

func newVolumeDriver(mounter fuse.Mounter, baseMountpoint string) *volumeDriver {
	return &volumeDriver{
		mounter,
		baseMountpoint,
		make(map[string]*volume),
		&sync.RWMutex{},
	}
}

func (v *volumeDriver) Create(name string, opts map[string]string) error {
	repository, err := getRequiredString(opts, "repository")
	if err != nil {
		return err
	}
	commitID, err := getRequiredString(opts, "commit_id")
	if err != nil {
		return err
	}
	shard, err := getOptionalUint64(opts, "shard", defaultShard)
	if err != nil {
		return err
	}
	modulus, err := getOptionalUint64(opts, "modulus", defaultModulus)
	if err != nil {
		return err
	}
	dir := filepath.Join(v.baseMountpoint, strings.Replace(uuid.NewV4().String(), "-", "", -1))
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	volume := &volume{
		repository,
		commitID,
		shard,
		modulus,
		dir,
	}
	v.lock.Lock()
	if _, ok := v.nameToVolume[name]; ok {
		v.lock.Unlock()
		return fmt.Errorf("pfs-volume-driver: volume already exists: %s", name)
	}
	v.nameToVolume[name] = volume
	v.lock.Unlock()
	return nil
}

func (v *volumeDriver) Remove(name string) error {
	v.lock.Lock()
	if _, ok := v.nameToVolume[name]; !ok {
		v.lock.Unlock()
		return fmt.Errorf("pfs-volume-driver: volume does not exist: %s", name)
	}
	delete(v.nameToVolume, name)
	v.lock.Unlock()
	return nil
}

func (v *volumeDriver) Path(name string) (string, error) {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("pfs-volume-driver: volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	return volume.mountpoint, nil
}

func (v *volumeDriver) Mount(name string) (string, error) {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("pfs-volume-driver: volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	if err := v.mounter.Mount(
		volume.repository,
		volume.commitID,
		volume.mountpoint,
		volume.shard,
		volume.modulus,
	); err != nil {
		return "", err
	}
	return volume.mountpoint, nil
}

func (v *volumeDriver) Unmount(name string) error {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return fmt.Errorf("pfs-volume-driver: volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	if err := v.mounter.Unmount(volume.mountpoint); err != nil {
		return err
	}
	return v.mounter.Wait(volume.mountpoint)
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
