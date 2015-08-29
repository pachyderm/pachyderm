package volume

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type volume struct {
	repository string
	commitID   string
	shard      uint64
	modulus    uint64
}

type volumeDriver struct {
	apiClient pfs.ApiClient

	nameToVolume map[string]*volume
	lock         *sync.Mutex
}

func newVolumeDriver(apiClient pfs.ApiClient) *volumeDriver {
	return &volumeDriver{
		apiClient,
		make(map[string]*volume),
		&sync.Mutex{},
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
	}
	v.lock.Lock()
	defer v.lock.Unlock()
	if _, ok := v.nameToVolume[name]; ok {
		return fmt.Errorf("volume already exists: %s", name)
	}
	v.nameToVolume[name] = volume
	return nil
}

func (v *volumeDriver) Remove(name string) error {
	return nil
}

func (v *volumeDriver) Path(name string) (string, error) {
	return "", nil
}

func (v *volumeDriver) Mount(name string) (string, error) {
	return "", nil
}

func (v *volumeDriver) Unmount(name string) error {
	return nil
}
