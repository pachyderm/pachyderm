package provider

import (
	"golang.org/x/net/context"
)

// Provider represents a service provider, such as google.
// TODO: clarify what this type is for
type Provider interface {
	CreateDisk(name string, sizeGb int64) error
}

// NewGoogleProvider returns a google provider with the given project and zone.
func NewGoogleProvider(ctx context.Context, project string, zone string) (Provider, error) {
	return newGoogleProvider(ctx, project, zone)
}
