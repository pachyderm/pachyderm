package s3

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
)

type ClientFactory interface {
	Client(authToken string) (*client.APIClient, error)
}

type LocalClientFactory struct {
	port uint16
}

func NewLocalClientFactory(port uint16) *LocalClientFactory {
	return &LocalClientFactory{
		port: port,
	}
}

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
