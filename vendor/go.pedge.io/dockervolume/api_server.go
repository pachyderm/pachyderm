package dockervolume

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type volume struct {
	name       string
	opts       map[string]string
	mountpoint string
}

type apiServer struct {
	volumeDriver     VolumeDriver
	volumeDriverName string
	nameToVolume     map[string]*volume
	lock             *sync.RWMutex
}

func newAPIServer(volumeDriver VolumeDriver, volumeDriverName string) *apiServer {
	return &apiServer{
		volumeDriver,
		volumeDriverName,
		make(map[string]*volume),
		&sync.RWMutex{},
	}
}

func (a *apiServer) Activate(_ context.Context, request *google_protobuf.Empty) (response *ActivateResponse, err error) {
	start := time.Now()
	defer func() { a.log("Activate", request, response, err, start) }()
	return &ActivateResponse{
		Implements: []string{
			"VolumeDriver",
		},
	}, nil
}

func (a *apiServer) Create(_ context.Context, request *CreateRequest) (response *CreateResponse, err error) {
	start := time.Now()
	defer func() { a.log("Create", request, response, err, start) }()
	volume := &volume{
		request.Name,
		request.Opts,
		"",
	}
	a.lock.Lock()
	if _, ok := a.nameToVolume[request.Name]; ok {
		a.lock.Unlock()
		return &CreateResponse{
			Err: fmt.Sprintf("dockervolume: volume already created: %s", request.Name),
		}, nil
	}
	if err := a.volumeDriver.Create(request.Name, request.Opts); err != nil {
		a.lock.Unlock()
		return &CreateResponse{
			Err: err.Error(),
		}, nil
	}
	a.nameToVolume[request.Name] = volume
	a.lock.Unlock()
	return &CreateResponse{}, nil
}

func (a *apiServer) Remove(_ context.Context, request *RemoveRequest) (response *RemoveResponse, err error) {
	start := time.Now()
	defer func() { a.log("Remove", request, response, err, start) }()
	a.lock.Lock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		a.lock.Unlock()
		return &RemoveResponse{
			Err: fmt.Sprintf("dockervolume: volume does not exist: %s", request.Name),
		}, nil
	}
	delete(a.nameToVolume, request.Name)
	if err := a.volumeDriver.Remove(volume.name, copyStringStringMap(volume.opts), volume.mountpoint); err != nil {
		a.lock.Unlock()
		return &RemoveResponse{
			Err: err.Error(),
		}, nil
	}
	a.lock.Unlock()
	return &RemoveResponse{}, nil
}

func (a *apiServer) Path(_ context.Context, request *PathRequest) (response *PathResponse, err error) {
	start := time.Now()
	defer func() { a.log("Path", request, response, err, start) }()
	a.lock.RLock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		a.lock.RUnlock()
		return &PathResponse{
			Err: fmt.Sprintf("dockervolume: volume does not exist: %s", request.Name),
		}, nil
	}
	a.lock.RUnlock()
	return &PathResponse{
		Mountpoint: volume.mountpoint,
	}, nil
}

func (a *apiServer) Mount(_ context.Context, request *MountRequest) (response *MountResponse, err error) {
	start := time.Now()
	defer func() { a.log("Mount", request, response, err, start) }()
	a.lock.Lock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		a.lock.Unlock()
		return &MountResponse{
			Err: fmt.Sprintf("dockervolume: volume does not exist: %s", request.Name),
		}, nil
	}
	if volume.mountpoint != "" {
		a.lock.Unlock()
		return &MountResponse{
			Err: fmt.Sprintf("dockervolume: volume already mounted: %s at %s", volume.name, volume.mountpoint),
		}, nil
	}
	mountpoint, err := a.volumeDriver.Mount(volume.name, copyStringStringMap(volume.opts))
	volume.mountpoint = mountpoint
	if err != nil {
		a.lock.Unlock()
		return &MountResponse{
			Mountpoint: mountpoint,
			Err:        err.Error(),
		}, nil
	}
	a.lock.Unlock()
	return &MountResponse{
		Mountpoint: mountpoint,
	}, nil
}

func (a *apiServer) Unmount(_ context.Context, request *UnmountRequest) (response *UnmountResponse, err error) {
	start := time.Now()
	defer func() { a.log("Unmount", request, response, err, start) }()
	a.lock.Lock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		a.lock.Unlock()
		return &UnmountResponse{
			Err: fmt.Sprintf("dockervolume: volume does not exist: %s", request.Name),
		}, nil
	}
	if volume.mountpoint == "" {
		a.lock.Unlock()
		return &UnmountResponse{
			Err: fmt.Sprintf("dockervolume: volume not mounted: %s at %s", volume.name, volume.mountpoint),
		}, nil
	}
	mountpoint := volume.mountpoint
	volume.mountpoint = ""
	if err := a.volumeDriver.Unmount(volume.name, copyStringStringMap(volume.opts), mountpoint); err != nil {
		a.lock.Unlock()
		return &UnmountResponse{
			Err: err.Error(),
		}, nil
	}
	a.lock.Unlock()
	return &UnmountResponse{}, nil
}

func (a *apiServer) Cleanup(_ context.Context, request *google_protobuf.Empty) (response *CleanupResponse, err error) {
	start := time.Now()
	defer func() { a.log("Cleanup", request, response, err, start) }()
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return &CleanupResponse{
			Err: err.Error(),
		}, nil
	}
	allVolumes, err := client.ListVolumes(docker.ListVolumesOptions{})
	if err != nil {
		return &CleanupResponse{
			Err: err.Error(),
		}, nil
	}
	var driverVolumes []docker.Volume
	for _, volume := range allVolumes {
		if volume.Driver == a.volumeDriverName {
			driverVolumes = append(driverVolumes, volume)
		}
	}
	var volumes []docker.Volume
	a.lock.RLock()
	for _, volume := range driverVolumes {
		if _, ok := a.nameToVolume[volume.Name]; ok {
			volumes = append(volumes, volume)
		}
	}
	a.lock.RUnlock()
	removeVolumeAttempts := make([]*RemoveVolumeAttempt, len(volumes))
	for i, volume := range volumes {
		removeVolumeAttempt := &RemoveVolumeAttempt{
			Name: volume.Name,
		}
		if err := client.RemoveVolume(volume.Name); err != nil {
			removeVolumeAttempt.Err = err.Error()
		}
		removeVolumeAttempts[i] = removeVolumeAttempt
	}
	return &CleanupResponse{
		RemoveVolumeAttempt: removeVolumeAttempts,
	}, nil
}

func (a *apiServer) log(methodName string, request proto.Message, response proto.Message, err error, start time.Time) {
	protorpclog.Info("dockervolume.API", methodName, request, response, err, time.Since(start))
}

func copyStringStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	n := make(map[string]string, len(m))
	for key, value := range m {
		n[key] = value
	}
	return n
}
