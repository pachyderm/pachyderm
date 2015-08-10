package grpcversion

import (
	"fmt"

	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

func NewAPIServer(version *Version) ApiServer {
	return newAPIServer(version)
}

func VersionString(version *Version) string {
	return fmt.Sprintf("%d.%d.%d%s", version.Major, version.Minor, version.Micro, version.Additional)
}

func GetVersion(apiClient ApiClient) (*Version, error) {
	getVersionResponse, err := apiClient.GetVersion(
		context.Background(),
		emptyInstance,
	)
	if err != nil {
		return nil, err
	}
	return getVersionResponse.Version, nil
}
