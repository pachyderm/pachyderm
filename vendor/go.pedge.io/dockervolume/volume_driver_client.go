package dockervolume

import (
	"errors"

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
	response, err := v.apiClient.Create(
		context.Background(),
		&CreateRequest{
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

func (v *volumeDriverClient) Remove(name string) error {
	response, err := v.apiClient.Remove(
		context.Background(),
		&RemoveRequest{
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

func (v *volumeDriverClient) Path(name string) (string, error) {
	response, err := v.apiClient.Path(
		context.Background(),
		&PathRequest{
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

func (v *volumeDriverClient) Mount(name string) (string, error) {
	response, err := v.apiClient.Mount(
		context.Background(),
		&MountRequest{
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

func (v *volumeDriverClient) Unmount(name string) error {
	response, err := v.apiClient.Unmount(
		context.Background(),
		&UnmountRequest{
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

func (v *volumeDriverClient) Cleanup() ([]*RemoveVolumeAttempt, error) {
	response, err := v.apiClient.Cleanup(
		context.Background(),
		&google_protobuf.Empty{},
	)
	if err != nil {
		return nil, err
	}
	return response.RemoveVolumeAttempt, nil
}

func (v *volumeDriverClient) GetVolume(name string) (*Volume, error) {
	return v.apiClient.GetVolume(
		context.Background(),
		&GetVolumeRequest{
			Name: name,
		},
	)
}

func (v *volumeDriverClient) ListVolumes() ([]*Volume, error) {
	response, err := v.apiClient.ListVolumes(
		context.Background(),
		&google_protobuf.Empty{},
	)
	if err != nil {
		return nil, err
	}
	return response.Volume, nil
}

func (v *volumeDriverClient) GetEventsByVolume(name string) ([]*Event, error) {
	response, err := v.apiClient.GetEventsByVolume(
		context.Background(),
		&GetEventsByVolumeRequest{
			VolumeName: name,
		},
	)
	if err != nil {
		return nil, err
	}
	return response.Event, nil
}

func (v *volumeDriverClient) ListEvents() ([]*Event, error) {
	response, err := v.apiClient.ListEvents(
		context.Background(),
		&google_protobuf.Empty{},
	)
	if err != nil {
		return nil, err
	}
	return response.Event, nil
}
