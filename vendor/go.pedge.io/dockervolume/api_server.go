package dockervolume

import (
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

var (
	activateResponse = &ActivateResponse{
		Implements: []string{
			"VolumeDriver",
		},
	}
)

type apiServer struct {
	volumeDriver VolumeDriver
}

func newAPIServer(volumeDriver VolumeDriver) *apiServer {
	return &apiServer{volumeDriver}
}

func (a *apiServer) Activate(_ context.Context, _ *google_protobuf.Empty) (*ActivateResponse, error) {
	return activateResponse, nil
}

func (a *apiServer) Create(_ context.Context, request *CreateRequest) (*CreateResponse, error) {
	if err := a.volumeDriver.Create(request.Name, request.Opts); err != nil {
		return &CreateResponse{
			Err: err.Error(),
		}, nil
	}
	return &CreateResponse{}, nil
}

func (a *apiServer) Remove(_ context.Context, request *RemoveRequest) (*RemoveResponse, error) {
	if err := a.volumeDriver.Remove(request.Name); err != nil {
		return &RemoveResponse{
			Err: err.Error(),
		}, nil
	}
	return &RemoveResponse{}, nil
}

func (a *apiServer) Path(_ context.Context, request *PathRequest) (*PathResponse, error) {
	mountpoint, err := a.volumeDriver.Path(request.Name)
	if err != nil {
		return &PathResponse{
			Mountpoint: mountpoint,
			Err:        err.Error(),
		}, nil
	}
	return &PathResponse{
		Mountpoint: mountpoint,
	}, nil
}

func (a *apiServer) Mount(_ context.Context, request *MountRequest) (*MountResponse, error) {
	mountpoint, err := a.volumeDriver.Mount(request.Name)
	if err != nil {
		return &MountResponse{
			Mountpoint: mountpoint,
			Err:        err.Error(),
		}, nil
	}
	return &MountResponse{
		Mountpoint: mountpoint,
	}, nil
}

func (a *apiServer) Unmount(_ context.Context, request *UnmountRequest) (*UnmountResponse, error) {
	if err := a.volumeDriver.Unmount(request.Name); err != nil {
		return &UnmountResponse{
			Err: err.Error(),
		}, nil
	}
	return &UnmountResponse{}, nil
}
