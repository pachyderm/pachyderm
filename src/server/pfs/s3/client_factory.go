package s3

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
)

// ClientFactory implementors provide a way for the s3gateway to instantiate
// and configure a request-scoped client.
type ClientFactory interface {
	// Client creates a pachyderm client.
	Client(authToken string) (*client.APIClient, error)
}

// LocalClientFactory creates clients that connect to localhost on a
// configurable port.
type LocalClientFactory struct {
	port uint16
}

// NewLocalClientFactory creates a new LocalClientFactory, using the given
// port.
func NewLocalClientFactory(port uint16) *LocalClientFactory {
	return &LocalClientFactory{
		port: port,
	}
}

// Client creates a pachyderm client.
func (f *LocalClientFactory) Client(authToken string) (*client.APIClient, error) {
	pc, err := client.NewFromAddress(fmt.Sprintf("localhost:%d", f.port))
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		pc.SetAuthToken(authToken)
	}
	return pc, nil
}
