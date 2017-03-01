package pps

import (
	"google.golang.org/grpc"
)

const (
	// JobDataPath specifies ephemereal storage on the host for job data.
	JobDataPath = "/var/pachyderm/job-data"
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
