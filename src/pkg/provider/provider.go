package provider

type Provider interface {
}

func NewProvider(ctx context.Context) (Client, error) {
	return newGoogleClient(ctx)
}
