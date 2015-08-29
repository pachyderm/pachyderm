package volume

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs/fuse"
)

const (
	tempdirPrefix = "pfs-docker-volume"
)

type volume struct {
	repository string
	commitID   string
	shard      uint64
	modulus    uint64
	mountpoint string
	lock       *sync.RWMutex
}

type volumeDriver struct {
	mounter fuse.Mounter

	nameToVolume map[string]*volume
	lock         *sync.RWMutex
}

func newVolumeDriver(mounter fuse.Mounter) *volumeDriver {
	return &volumeDriver{
		mounter,
		make(map[string]*volume),
		&sync.RWMutex{},
	}
}

func (v *volumeDriver) Create(name string, opts map[string]string) error {
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
	volume := &volume{
		repository,
		commitID,
		shard,
		modulus,
		"",
		&sync.RWMutex{},
	}
	v.lock.Lock()
	if _, ok := v.nameToVolume[name]; ok {
		v.lock.Unlock()
		return fmt.Errorf("volume already exists: %s", name)
	}
	v.nameToVolume[name] = volume
	v.lock.Unlock()
	return nil
}

func (v *volumeDriver) Remove(name string) error {
	v.lock.Lock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.Unlock()
		return fmt.Errorf("volume does not exist: %s", name)
	}
	delete(v.nameToVolume, name)
	v.lock.Unlock()

	volume.lock.RLock()
	if volume.mountpoint != "" {
		v.lock.Lock()
		if _, ok := v.nameToVolume[name]; ok {
			v.lock.Unlock()
			return fmt.Errorf("volume already exists: %s", name)
		}
		v.nameToVolume[name] = volume
		volume.lock.RUnlock()
		v.lock.Unlock()
		return fmt.Errorf("volume %s still has mountpoint %s", name, volume.mountpoint)
	}
	volume.lock.RUnlock()
	return nil
}

func (v *volumeDriver) Path(name string) (string, error) {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("volume does not exist: %s", name)
	}
	v.lock.RUnlock()
	// TODO(pedge): should this return an error if volume.mountpoint == ""?
	return volume.mountpoint, nil
}

func (v *volumeDriver) Mount(name string) (string, error) {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return "", fmt.Errorf("volume does not exist: %s", name)
	}
	v.lock.RUnlock()

	volume.lock.Lock()
	defer volume.lock.Unlock()
	if volume.mountpoint != "" {
		return "", fmt.Errorf("volume %s already mounted at %s", name, volume.mountpoint)
	}
	tempdir, err := ioutil.TempDir("", tempdirPrefix)
	if err != nil {
		return "", err
	}
	if err := v.mounter.Mount(
		volume.repository,
		volume.commitID,
		tempdir,
		volume.shard,
		volume.modulus,
	); err != nil {
		return "", err
	}
	volume.mountpoint = tempdir
	return tempdir, nil
}

func (v *volumeDriver) Unmount(name string) error {
	v.lock.RLock()
	volume, ok := v.nameToVolume[name]
	if !ok {
		v.lock.RUnlock()
		return fmt.Errorf("volume does not exist: %s", name)
	}
	v.lock.RUnlock()

	volume.lock.Lock()
	defer volume.lock.Unlock()
	if volume.mountpoint == "" {
		return fmt.Errorf("volume %s not mounted", name)
	}
	mountpoint := volume.mountpoint
	if err := v.mounter.Unmount(mountpoint); err != nil {
		return err
	}
	volume.mountpoint = ""
	return v.mounter.Wait(mountpoint)
}
