package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

func GetVersion(apiClient pps.ApiClient) (*pps.GetVersionResponse, error) {
	return apiClient.GetVersion(
		context.Background(),
		&google_protobuf.Empty{},
	)
}
