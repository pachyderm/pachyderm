package dockervolume

import (
	"errors"

	"google.golang.org/grpc"

	"go.pedge.io/google-protobuf"

	"golang.org/x/net/context"
)

type volumeDriverClient struct {
	apiClient APIClient
}

func newVolumeDriverClient(apiClient APIClient) *volumeDriverClient {
	return &volumeDriverClient{apiClient}
}

func (v *volumeDriverClient) Create(name string, opts map[string]string) error {
	return callNameOptsToErr(name, opts, v.apiClient.Create)
}

func (v *volumeDriverClient) Remove(name string) error {
	return callNameToErr(name, v.apiClient.Remove)
}

func (v *volumeDriverClient) Path(name string) (string, error) {
	return callNameToMountpointErr(name, v.apiClient.Path)
}

func (v *volumeDriverClient) Mount(name string) (string, error) {
	return callNameToMountpointErr(name, v.apiClient.Mount)
}

func (v *volumeDriverClient) Unmount(name string) error {
	return callNameToErr(name, v.apiClient.Unmount)
}

func (v *volumeDriverClient) Cleanup() ([]*Volume, error) {
	response, err := v.apiClient.Cleanup(
		context.Background(),
		google_protobuf.EmptyInstance,
	)
	if err != nil {
		return nil, err
	}
	return response.Volume, nil
}

func (v *volumeDriverClient) GetVolume(name string) (*Volume, error) {
	return v.apiClient.GetVolume(
		context.Background(),
		&NameRequest{
			Name: name,
		},
	)
}

func (v *volumeDriverClient) ListVolumes() ([]*Volume, error) {
	response, err := v.apiClient.ListVolumes(
		context.Background(),
		google_protobuf.EmptyInstance,
	)
	if err != nil {
		return nil, err
	}
	return response.Volume, nil
}

func callNameOptsToErr(
	name string,
	opts map[string]string,
	f func(context.Context, *NameOptsRequest, ...grpc.CallOption) (*ErrResponse, error),
) error {
	response, err := f(
		context.Background(),
		&NameOptsRequest{
			Name: name,
			Opts: opts,
		},
	)
	if err != nil {
		return err
	}
	if response.Err != "" {
		return errors.New(response.Err)
	}
	return nil
}

func callNameToErr(
	name string,
	f func(context.Context, *NameRequest, ...grpc.CallOption) (*ErrResponse, error),
) error {
	response, err := f(
		context.Background(),
		&NameRequest{
			Name: name,
		},
	)
	if err != nil {
		return err
	}
	if response.Err != "" {
		return errors.New(response.Err)
	}
	return nil
}

func callNameToMountpointErr(
	name string,
	f func(context.Context, *NameRequest, ...grpc.CallOption) (*MountpointErrResponse, error),
) (string, error) {
	response, err := f(
		context.Background(),
		&NameRequest{
			Name: name,
		},
	)
	if err != nil {
		return response.Mountpoint, err
	}
	if response.Err != "" {
		return response.Mountpoint, errors.New(response.Err)
	}
	return response.Mountpoint, nil
}
