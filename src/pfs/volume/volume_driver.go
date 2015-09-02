package volume

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"go.pedge.io/protolog"

	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/satori/go.uuid"
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
	protolog.Infof("Create(name:%s, opts:%v)", name, opts)
	repository, ok := opts["repository"]
	if !ok {
		return fmt.Errorf("option repository not found in %v", opts)
	}
	commitID, ok := opts["commit_id"]
	if !ok {
		return fmt.Errorf("option commit_id not found in %v", opts)
	}
	shardObj, ok := opts["shard"]
	if !ok {
		return fmt.Errorf("option shard not found in %v", opts)
	}
	shard, err := strconv.ParseUint(shardObj, 10, 64)
	if err != nil {
		return err
	}
	modulusObj, ok := opts["modulus"]
	if !ok {
		return fmt.Errorf("option modulus not found in %v", opts)
	}
	modulus, err := strconv.ParseUint(modulusObj, 10, 64)
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
		return fmt.Errorf("volume already exists: %s", name)
	}
	v.nameToVolume[name] = volume
	v.lock.Unlock()
	protolog.Infof("Create name:%s volume:%+v", name, volume)
	return nil
}

func (v *volumeDriver) Remove(name string) error {
	protolog.Infof("Remove(name:%s)", name)
	v.lock.Lock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.Unlock()
		return fmt.Errorf("volume does not exist: %s", name)
	}
	delete(v.nameToVolume, name)
	v.lock.Unlock()
	protolog.Infof("Remove name:%s volume:%+v", name, volume)
	return nil
}

func (v *volumeDriver) Path(name string) (string, error) {
	protolog.Infof("Path(name:%s)", name)
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	protolog.Infof("Path name:%s volume:%+v", name, volume)
	return volume.mountpoint, nil
}

func (v *volumeDriver) Mount(name string) (string, error) {
	protolog.Infof("Mount(name:%s)", name)
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("volume does not exist: %s", name)
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
	protolog.Infof("Mount name:%s volume:%+v", name, volume)
	return volume.mountpoint, nil
}

func (v *volumeDriver) Unmount(name string) error {
	protolog.Infof("Unmount(name:%s)", name)
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return fmt.Errorf("volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	if err := v.mounter.Unmount(volume.mountpoint); err != nil {
		return err
	}
	if err := v.mounter.Wait(volume.mountpoint); err != nil {
		return err
	}
	protolog.Infof("Unmount name:%s volume:%+v", name, volume)
	return nil
}
