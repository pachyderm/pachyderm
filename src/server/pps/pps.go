package pps

import (
	"google.golang.org/grpc"
)

// NewInternalPodAPIClientFromAddress creates an InternalPodAPIClient
// connecting to pachd at pachAddr.
func NewInternalPodAPIClientFromAddress(pachAddr string) (InternalPodAPIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := NewInternalPodAPIClient(clientConn)

	return client, nil
}
