package dockervolume

import (
	"time"

	"github.com/golang/protobuf/proto"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
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

func (a *apiServer) Activate(_ context.Context, request *google_protobuf.Empty) (response *ActivateResponse, err error) {
	start := time.Now()
	defer func() { a.log("Activate", request, response, err, start) }()
	return activateResponse, nil
}

func (a *apiServer) Create(_ context.Context, request *CreateRequest) (response *CreateResponse, err error) {
	start := time.Now()
	defer func() { a.log("Create", request, response, err, start) }()
	if err := a.volumeDriver.Create(request.Name, request.Opts); err != nil {
		return &CreateResponse{
			Err: err.Error(),
		}, nil
	}
	return &CreateResponse{}, nil
}

func (a *apiServer) Remove(_ context.Context, request *RemoveRequest) (response *RemoveResponse, err error) {
	start := time.Now()
	defer func() { a.log("Remove", request, response, err, start) }()
	if err := a.volumeDriver.Remove(request.Name); err != nil {
		return &RemoveResponse{
			Err: err.Error(),
		}, nil
	}
	return &RemoveResponse{}, nil
}

func (a *apiServer) Path(_ context.Context, request *PathRequest) (response *PathResponse, err error) {
	start := time.Now()
	defer func() { a.log("Path", request, response, err, start) }()
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

func (a *apiServer) Mount(_ context.Context, request *MountRequest) (response *MountResponse, err error) {
	start := time.Now()
	defer func() { a.log("Mount", request, response, err, start) }()
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

func (a *apiServer) Unmount(_ context.Context, request *UnmountRequest) (response *UnmountResponse, err error) {
	start := time.Now()
	defer func() { a.log("Unmount", request, response, err, start) }()
	if err := a.volumeDriver.Unmount(request.Name); err != nil {
		return &UnmountResponse{
			Err: err.Error(),
		}, nil
	}
	return &UnmountResponse{}, nil
}

func (a *apiServer) log(methodName string, request proto.Message, response proto.Message, err error, start time.Time) {
	protorpclog.Info("dockervolume.API", methodName, request, response, err, time.Since(start))
}
