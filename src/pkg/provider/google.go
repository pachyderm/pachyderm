package storage

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type googleProvider struct {
	service *compute.Service
}

func newGoogleProvider(ctx context.Context) (*googleProvider, error) {
	httpClient, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	service, err := compute.New(httpClient)
	if err != nil {
		return nil, err
	}
	return &googleClient{service: service}, nil
}
