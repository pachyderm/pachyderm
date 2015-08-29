package volume

import "github.com/pachyderm/pachyderm/src/pfs"

type volumeDriver struct {
	apiClient pfs.ApiClient
}

func newVolumeDriver(apiClient pfs.ApiClient) *volumeDriver {
	return &volumeDriver{
		apiClient,
	}
}

func (v *volumeDriver) Create(name string, opts map[string]string) error {
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
