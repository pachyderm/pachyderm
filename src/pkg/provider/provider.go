package provider

type Provider interface {
	CreateDisk(name string, sizeGb int64) error
}

func NewGoogleProvider(ctx context.Context, project string, zone string) (Client, error) {
	return newGoogleClient(ctx, project, zone)
}
