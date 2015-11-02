package storage

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type googleProvider struct {
	service *compute.Service
	project string
	zone    string
}

func newGoogleProvider(ctx context.Context, project string, zone string) (*googleProvider, error) {
	httpClient, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	service, err := compute.New(httpClient)
	if err != nil {
		return nil, err
	}
	return &googleClient{
		service: service,
		project: project,
		zone:    zone,
	}, nil
}

func (p *googleProvider) CreateDisk(name string, sizeGb int64) error {
	disk, err := p.service.Disks.Insert(
		project,
		zone,
		&compute.Disk{
			Name:   name,
			SizeGb: sizeGb,
			Type:   "pd-ssd",
		},
	).Do()
	if err != nil {
		return err
	}
}
