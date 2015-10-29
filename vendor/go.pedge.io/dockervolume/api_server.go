package dockervolume

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/fsouza/go-dockerclient"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/map"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	volumeDriver     VolumeDriver
	volumeDriverName string
	nameToVolume     map[string]*Volume
	lock             *sync.RWMutex
}

func newAPIServer(volumeDriver VolumeDriver, volumeDriverName string, noEvents bool) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("dockervolume.API"),
		volumeDriver,
		volumeDriverName,
		make(map[string]*Volume),
		&sync.RWMutex{},
	}
}

func (a *apiServer) Activate(_ context.Context, request *google_protobuf.Empty) (response *ActivateResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return &ActivateResponse{
		Implements: []string{
			"VolumeDriver",
		},
	}, nil
}

func (a *apiServer) Create(_ context.Context, request *NameOptsRequest) (response *ErrResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return doNameOptsToErr(request, a.create)
}

func (a *apiServer) create(name string, opts map[string]string) error {
	volume := &Volume{
		name,
		opts,
		"",
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	if _, ok := a.nameToVolume[name]; ok {
		return fmt.Errorf("dockervolume: volume already created: %s", name)
	}
	if err := a.volumeDriver.Create(name, pkgmap.StringStringMap(opts)); err != nil {
		return err
	}
	a.nameToVolume[name] = volume
	return nil
}

func (a *apiServer) Remove(_ context.Context, request *NameRequest) (response *ErrResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return doNameToErr(request, a.remove)
}

func (a *apiServer) remove(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	volume, ok := a.nameToVolume[name]
	if !ok {
		return fmt.Errorf("dockervolume: volume does not exist: %s", name)
	}
	delete(a.nameToVolume, name)
	return a.volumeDriver.Remove(volume.Name, pkgmap.StringStringMap(volume.Opts).Copy(), volume.Mountpoint)
}

func (a *apiServer) Path(_ context.Context, request *NameRequest) (response *MountpointErrResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return doNameToMountpointErr(request, a.path)
}

func (a *apiServer) path(name string) (string, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	volume, ok := a.nameToVolume[name]
	if !ok {
		return "", fmt.Errorf("dockervolume: volume does not exist: %s", name)
	}
	return volume.Mountpoint, nil
}

func (a *apiServer) Mount(_ context.Context, request *NameRequest) (response *MountpointErrResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return doNameToMountpointErr(request, a.mount)
}

func (a *apiServer) mount(name string) (string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	volume, ok := a.nameToVolume[name]
	if !ok {
		return "", fmt.Errorf("dockervolume: volume does not exist: %s", name)
	}
	if volume.Mountpoint != "" {
		return "", fmt.Errorf("dockervolume: volume already mounted: %s at %s", volume.Name, volume.Mountpoint)
	}
	mountpoint, err := a.volumeDriver.Mount(volume.Name, pkgmap.StringStringMap(volume.Opts).Copy())
	volume.Mountpoint = mountpoint
	return mountpoint, err
}

func (a *apiServer) Unmount(_ context.Context, request *NameRequest) (response *ErrResponse, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return doNameToErr(request, a.unmount)
}

func (a *apiServer) unmount(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	volume, ok := a.nameToVolume[name]
	if !ok {
		return fmt.Errorf("dockervolume: volume does not exist: %s", name)
	}
	if volume.Mountpoint == "" {
		return fmt.Errorf("dockervolume: volume not mounted: %s at %s", volume.Name, volume.Mountpoint)
	}
	mountpoint := volume.Mountpoint
	volume.Mountpoint = ""
	return a.volumeDriver.Unmount(volume.Name, pkgmap.StringStringMap(volume.Opts).Copy(), mountpoint)
}

func (a *apiServer) Cleanup(_ context.Context, request *google_protobuf.Empty) (response *Volumes, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	allVolumes, err := client.ListVolumes(docker.ListVolumesOptions{})
	if err != nil {
		return nil, err
	}
	var driverVolumes []docker.Volume
	for _, volume := range allVolumes {
		if volume.Driver == a.volumeDriverName {
			driverVolumes = append(driverVolumes, volume)
		}
	}
	var volumes []*Volume
	a.lock.RLock()
	for _, dockerVolume := range driverVolumes {
		if volume, ok := a.nameToVolume[dockerVolume.Name]; ok {
			volumes = append(volumes, copyVolume(volume))
		}
	}
	a.lock.RUnlock()
	var errs []error
	for _, volume := range volumes {
		if err := client.RemoveVolume(volume.Name); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		err = grpc.Errorf(codes.Internal, "%v", errs)
	}
	return &Volumes{
		Volume: volumes,
	}, err
}

func (a *apiServer) GetVolume(_ context.Context, request *NameRequest) (response *Volume, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	a.lock.RLock()
	defer a.lock.RUnlock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		return nil, grpc.Errorf(codes.NotFound, request.Name)
	}
	return copyVolume(volume), nil
}

func (a *apiServer) ListVolumes(_ context.Context, request *google_protobuf.Empty) (response *Volumes, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	a.lock.RLock()
	defer a.lock.RUnlock()
	volumes := make([]*Volume, len(a.nameToVolume))
	i := 0
	for _, volume := range a.nameToVolume {
		volumes[i] = copyVolume(volume)
		i++
	}
	return &Volumes{
		Volume: volumes,
	}, nil
}

func fromNameOptsRequest(request *NameOptsRequest) (string, map[string]string) {
	return request.Name, pkgmap.StringStringMap(request.Opts).Copy()
}

func fromNameRequest(request *NameRequest) string {
	return request.Name
}

func toErrResponse(err error) (*ErrResponse, error) {
	response := &ErrResponse{}
	if err != nil {
		response.Err = err.Error()
	}
	return response, nil
}

func toMountpointErrResponse(mountpoint string, err error) (*MountpointErrResponse, error) {
	response := &MountpointErrResponse{
		Mountpoint: mountpoint,
	}
	if err != nil {
		response.Err = err.Error()
	}
	return response, nil
}

func doNameOptsToErr(request *NameOptsRequest, f func(string, map[string]string) error) (*ErrResponse, error) {
	name, opts := fromNameOptsRequest(request)
	return toErrResponse(f(name, opts))
}

func doNameToErr(request *NameRequest, f func(string) error) (*ErrResponse, error) {
	return toErrResponse(f(fromNameRequest(request)))
}

func doNameToMountpointErr(request *NameRequest, f func(string) (string, error)) (*MountpointErrResponse, error) {
	mountpoint, err := f(fromNameRequest(request))
	return toMountpointErrResponse(mountpoint, err)
}

func copyVolume(volume *Volume) *Volume {
	if volume == nil {
		return nil
	}
	return &Volume{
		Name:       volume.Name,
		Opts:       pkgmap.StringStringMap(volume.Opts).Copy(),
		Mountpoint: volume.Mountpoint,
	}
}

func now() *google_protobuf.Timestamp {
	return prototime.TimeToTimestamp(time.Now().UTC())
}
