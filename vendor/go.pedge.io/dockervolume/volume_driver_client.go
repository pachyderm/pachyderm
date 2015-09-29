package dockervolume

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type volumeDriverClient struct {
	apiClient APIClient
}

func newVolumeDriverClient(clientConn *grpc.ClientConn) *volumeDriverClient {
	return &volumeDriverClient{NewAPIClient(clientConn)}
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
