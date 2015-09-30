package dockervolume

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
)

const (
	// TODO(pedge): rethink events, it's a little silly how it is implemented
	maxEventsSize = 65536
)

type apiServer struct {
	volumeDriver     VolumeDriver
	volumeDriverName string
	nameToVolume     map[string]*Volume
	noEvents         bool
	nameToEvents     map[string][]*Event
	numEvents        int
	lock             *sync.RWMutex
}

func newAPIServer(volumeDriver VolumeDriver, volumeDriverName string, noEvents bool) *apiServer {
	return &apiServer{
		volumeDriver,
		volumeDriverName,
		make(map[string]*Volume),
		noEvents,
		make(map[string][]*Event),
		0,
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
	volume := &Volume{
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
	a.addEvent(EventType_EVENT_TYPE_CREATED, volume)
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
	a.addEvent(EventType_EVENT_TYPE_REMOVED, volume)
	if err := a.volumeDriver.Remove(volume.Name, copyStringStringMap(volume.Opts), volume.Mountpoint); err != nil {
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
		Mountpoint: volume.Mountpoint,
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
	if volume.Mountpoint != "" {
		a.lock.Unlock()
		return &MountResponse{
			Err: fmt.Sprintf("dockervolume: volume already mounted: %s at %s", volume.Name, volume.Mountpoint),
		}, nil
	}
	mountpoint, err := a.volumeDriver.Mount(volume.Name, copyStringStringMap(volume.Opts))
	volume.Mountpoint = mountpoint
	a.addEvent(EventType_EVENT_TYPE_MOUNTED, volume)
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
	if volume.Mountpoint == "" {
		a.lock.Unlock()
		return &UnmountResponse{
			Err: fmt.Sprintf("dockervolume: volume not mounted: %s at %s", volume.Name, volume.Mountpoint),
		}, nil
	}
	a.addEvent(EventType_EVENT_TYPE_UNMOUNTED, volume)
	mountpoint := volume.Mountpoint
	volume.Mountpoint = ""
	if err := a.volumeDriver.Unmount(volume.Name, copyStringStringMap(volume.Opts), mountpoint); err != nil {
		a.lock.Unlock()
		return &UnmountResponse{
			Err: err.Error(),
		}, nil
	}
	a.lock.Unlock()
	return &UnmountResponse{}, nil
}

func (a *apiServer) Cleanup(_ context.Context, request *google_protobuf.Empty) (response *RemoveVolumeAttempts, err error) {
	start := time.Now()
	defer func() { a.log("Cleanup", request, response, err, start) }()
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
			volumes = append(volumes, volume)
		}
	}
	a.lock.RUnlock()
	removeVolumeAttempts := make([]*RemoveVolumeAttempt, len(volumes))
	for i, volume := range volumes {
		removeVolumeAttempt := &RemoveVolumeAttempt{
			Volume: copyVolume(volume),
		}
		if err := client.RemoveVolume(volume.Name); err != nil {
			removeVolumeAttempt.Err = err.Error()
		}
		removeVolumeAttempts[i] = removeVolumeAttempt
	}
	return &RemoveVolumeAttempts{
		RemoveVolumeAttempt: removeVolumeAttempts,
	}, nil
}

func (a *apiServer) GetVolume(_ context.Context, request *GetVolumeRequest) (response *Volume, err error) {
	start := time.Now()
	defer func() { a.log("GetVolume", request, response, err, start) }()
	a.lock.RLock()
	volume, ok := a.nameToVolume[request.Name]
	if !ok {
		a.lock.RUnlock()
		return nil, grpc.Errorf(codes.NotFound, request.Name)
	}
	volume = copyVolume(volume)
	a.lock.RUnlock()
	return volume, nil
}

func (a *apiServer) ListVolumes(_ context.Context, request *google_protobuf.Empty) (response *Volumes, err error) {
	start := time.Now()
	defer func() { a.log("ListVolumes", request, response, err, start) }()
	volumes := &Volumes{
		Volume: make([]*Volume, 0),
	}
	a.lock.RLock()
	for _, volume := range a.nameToVolume {
		volumes.Volume = append(volumes.Volume, copyVolume(volume))
	}
	a.lock.RUnlock()
	return volumes, nil
}

func (a *apiServer) GetEventsByVolume(_ context.Context, request *GetEventsByVolumeRequest) (response *Events, err error) {
	start := time.Now()
	defer func() { a.log("GetEventsByVolume", request, response, err, start) }()
	a.lock.RLock()
	events, ok := a.nameToEvents[request.VolumeName]
	if !ok {
		a.lock.RUnlock()
		return nil, grpc.Errorf(codes.NotFound, request.VolumeName)
	}
	response = &Events{
		Event: copyEvents(events),
	}
	a.lock.RUnlock()
	return response, nil
}

func (a *apiServer) ListEvents(_ context.Context, request *google_protobuf.Empty) (response *Events, err error) {
	start := time.Now()
	defer func() { a.log("ListEvents", request, response, err, start) }()
	response = &Events{
		Event: make([]*Event, 0),
	}
	a.lock.RLock()
	for _, events := range a.nameToEvents {
		if len(events) > 0 {
			response.Event = append(response.Event, copyEvents(events)...)
		}
	}
	a.lock.RUnlock()
	return response, nil
}

func (a *apiServer) log(methodName string, request proto.Message, response proto.Message, err error, start time.Time) {
	protorpclog.Info("dockervolume.API", methodName, request, response, err, time.Since(start))
}

// assumes write lock acquired
func (a *apiServer) addEvent(eventType EventType, volume *Volume) {
	if a.noEvents {
		return
	}
	if a.numEvents >= maxEventsSize {
		a.nameToEvents = make(map[string][]*Event, 0)
		a.numEvents = 0
	}
	if _, ok := a.nameToEvents[volume.Name]; !ok {
		a.nameToEvents[volume.Name] = make([]*Event, 0)
	}
	a.nameToEvents[volume.Name] = append(a.nameToEvents[volume.Name], newEvent(eventType, volume))
	a.numEvents++
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

func copyTimestamp(timestamp *google_protobuf.Timestamp) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: timestamp.Seconds,
		Nanos:   timestamp.Nanos,
	}
}

func copyVolume(volume *Volume) *Volume {
	if volume == nil {
		return nil
	}
	return &Volume{
		Name:       volume.Name,
		Opts:       copyStringStringMap(volume.Opts),
		Mountpoint: volume.Mountpoint,
	}
}

func copyEvent(event *Event) *Event {
	if event == nil {
		return nil
	}
	return &Event{
		EventType: event.EventType,
		Volume:    copyVolume(event.Volume),
		Timestamp: copyTimestamp(event.Timestamp),
	}
}

func copyEvents(events []*Event) []*Event {
	e := make([]*Event, len(events))
	for i, event := range events {
		e[i] = copyEvent(event)
	}
	return e
}

func newEvent(eventType EventType, volume *Volume) *Event {
	return &Event{
		EventType: eventType,
		Volume:    copyVolume(volume),
		Timestamp: now(),
	}
}

func now() *google_protobuf.Timestamp {
	return prototime.TimeToTimestamp(time.Now().UTC())
}
