package provider

import (
	"golang.org/x/net/context"
)

type Provider interface {
	CreateDisk(name string, sizeGb int64) error
}

func NewGoogleProvider(ctx context.Context, project string, zone string) (Provider, error) {
	return newGoogleProvider(ctx, project, zone)
}
