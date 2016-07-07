package pps

import (
	"google.golang.org/grpc"
)

// NewInternalJobAPIClientFromAddress creates an InternalJobAPIClient
// connecting to pachd at pachAddr.
func NewInternalJobAPIClientFromAddress(pachAddr string) (InternalJobAPIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := NewInternalJobAPIClient(clientConn)

	return client, nil
}
